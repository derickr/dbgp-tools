package proxy

import (
	"crypto/tls"
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/protocol"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
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

func (handler *ServerHandler) setupForwarder(conn net.Conn, initialPacket []byte, clientConnection *connections.Connection) error {
	handler.logger.LogUserInfo("proxy-client", clientConnection.GetKey(), "Connecting to %s", clientConnection.FullAddress())
	client, err := connectToIDE(clientConnection)

	if err != nil {
		handler.logger.LogUserError("proxy-client", clientConnection.GetKey(), "IDE not connected: %s", err)
		return err
	}

	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {
			handler.logger.LogUserError("proxy-client", clientConnection.GetKey(), "Closer didn't work: %v", err)
		}
	}(client)

	if err != nil {
		return err
	}

	handler.logger.LogUserInfo("proxy-client", clientConnection.GetKey(), "IDE connected")
	reassembledPacket := fmt.Sprintf("%d\000%s", len(initialPacket)-1, initialPacket)
	_, err = client.Write([]byte(reassembledPacket))
	if err != nil {
		return err
	}

	handler.logger.LogUserInfo("proxy-client", clientConnection.GetKey(), "Init forwarded, start pipe")
	clientChan := make(chan error)
	serverChan := make(chan error)

	conn.SetDeadline(time.Now().Add(time.Second * 2))

	go func() {
	restartCopy:
		_, err = io.Copy(client, conn)
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			conn.SetDeadline(time.Now().Add(time.Second * 2))
			goto restartCopy
		}
		clientChan <- err
	}()

	go func() {
	restartCopy:
		_, err = io.Copy(conn, client)
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			conn.SetDeadline(time.Now().Add(time.Second * 2))
			goto restartCopy
		}
		serverChan <- err
	}()

	for {
		select {
		case err = <-serverChan:
			handler.logger.LogUserInfo("proxy-client", clientConnection.GetKey(), "IDE closed connection")
			return nil
		case err = <-clientChan:
			handler.logger.LogUserInfo("proxy-client", clientConnection.GetKey(), "Xdebug connection closed")
			return nil
		default:
			time.Sleep(sleepTimeout)
		}
	}
}

func (handler *ServerHandler) Handle(conn net.Conn) error {
	var key string
	var connType string

	/* Set up interrupt handler */
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	reader := protocol.NewDbgpClient(conn, false, handler.logger)

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

		if err != nil {
			return fmt.Errorf("Error reading response: %v", err)
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
			handler.setupForwarder(conn, []byte(response), client)
		} else {
			handler.logger.LogUserInfo("proxy-client", key, "Could not find IDE connection for %s '%s': %s", connType, key, client.FullAddress())
		}
	}

	return nil
}
