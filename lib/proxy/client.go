package proxy

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/protocol"
)

const sleepTimeout = time.Millisecond * 50

type ServerHandler struct {
	logger         logger.Logger
	connectionList *connections.ConnectionList
}

func NewServerHandler(connectionList *connections.ConnectionList, logger logger.Logger) *ServerHandler {
	return &ServerHandler{connectionList: connectionList, logger: logger}
}

func connectToIDE(clientConnection *connections.Connection) (net.Conn, error) {
	if clientConnection.IsSSL() {
		cert, err := tls.LoadX509KeyPair("client-certs/client.pem", "client-certs/client.key")
		if err != nil {
			return nil, err
		}
		config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
		return tls.Dial("tcp", clientConnection.FullAddress(), &config)
	} else {
		return net.Dial("tcp", clientConnection.FullAddress())
	}
}

// Connects to IDE, sends the Init packet and proxies messages between IDE and Xdebug
func (handler *ServerHandler) setupForwarder(conn net.Conn, initialPacket []byte, clientConnection *connections.Connection) error {
	handler.logger.LogUserInfo("proxy-client", clientConnection.GetKey(), "Connecting to %s", clientConnection.FullAddress())
	client, err := connectToIDE(clientConnection)

	if err != nil {
		handler.logger.LogUserError("proxy-client", clientConnection.GetKey(), "IDE not connected: %s", err)
		return err
	}

	handler.logger.LogUserInfo("proxy-client", clientConnection.GetKey(), "IDE connected")
	reassembledPacket := fmt.Sprintf("%d\000%s", len(initialPacket)-1, initialPacket)
	_, err = client.Write([]byte(reassembledPacket))
	if err != nil {
		_ = client.Close()
		return err
	}

	handler.logger.LogUserInfo("proxy-client", clientConnection.GetKey(), "Init forwarded, start pipe")
	serverChan := make(chan error)

	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {
			handler.logger.LogUserError("proxy-client", clientConnection.GetKey(), "Closer didn't work: %v", err)
		}
		<-serverChan
	}(client)

	go func() {
		// client read loop
		if _, err = io.Copy(conn, client); err != nil {
			serverChan <- err
		}
		close(serverChan)
	}()

	// server read loop
	reader := protocol.NewDbgpClient(conn, handler.logger)
	for {
		response, err, timeout := reader.ReadResponseWithTimeout(2 * time.Second)

		if timeout {
			// has client disconnected or had a fatal error?
			select {
			case err := <-serverChan:
				handler.logger.LogUserInfo("proxy-client", clientConnection.GetKey(), "IDE closed connection")
				return fmt.Errorf("Client read error: %s", err)
			default:
			}
			continue
		}

		if err != nil {
			// protocol error or EOF
			// force close and return for cleanup
			handler.logger.LogError("proxy-client", "Protocol error reading from server: %s", err)
			conn.Close()
			return nil
		}

		if packet := reader.FormatXML(response); packet != nil && packet.ShouldCloseConnection() {
			// dbgp done, calling function will close connection or read next init packet from cloud
			return nil
		}

		// forward packet
		reassembledPacket := fmt.Sprintf("%d\000%s", len(response)-1, response)
		client.Write([]byte(reassembledPacket))
		// any error will re-appear on the top of the loop
	}
}

func (handler *ServerHandler) sendDetach(conn net.Conn) {
	err := protocol.NewDbgpClient(conn, handler.logger).RunCommand("detach -- \"dbgpProxy has no IDE connected to it\"")
	if err != nil {
		handler.logger.LogError("proxy-client", "Could not send 'detach': %s", err)
		return
	}
}

// The conn received is either a fresh Xdebug connection from a local Listen socket
// or an initialized Cloud connection (cloudinit)
func (handler *ServerHandler) Handle(conn net.Conn) error {
	var key string
	var connType string

	/* Set up interrupt handler */
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	reader := protocol.NewDbgpClient(conn, handler.logger)

ConnectionsLoop:
	for {
		response, err, timeout := reader.ReadResponseWithTimeout(2 * time.Second)

		select {
		case <-signals:
			break ConnectionsLoop
		default:
			if timeout {
				continue
			}
		}

		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return fmt.Errorf("Error reading response: %w", err)
		}

		init, _ := reader.ParseInitXML(response)

		switch {
		case init.CloudUserID != "":
			key = init.CloudUserID
			connType = "Cloud User"
		case init.IDEKey != "":
			key = init.IDEKey
			connType = "IDE Key"
		default:
			return fmt.Errorf("Both IDE Key and Cloud User are unset")
		}

		client, ok := handler.connectionList.FindByKey(key)

		if ok {
			handler.logger.LogUserInfo("proxy-client", key, "Found connection for %s '%s': %s", connType, key, client.FullAddress())
			if err := handler.setupForwarder(conn, []byte(response), client); err != nil {
				// Error indicates
				// - IDE connection failed or init packet send error - Xdebug/Cloud should be "released"
				// - Client disconnected and left Xdebug/Cloud in non-stopped status
				// - Connection to Xdebug/Cloud failed
				handler.logger.LogUserWarning("proxy-client", key, "Removed connection information for '%s': %s", key, err)
				handler.connectionList.RemoveByKey(key)
				handler.sendDetach(conn)
			}
		} else {
			handler.logger.LogUserInfo("proxy-client", key, "Could not find IDE connection for %s '%s'", connType, key)
			handler.sendDetach(conn)
		}
	}

	return nil
}
