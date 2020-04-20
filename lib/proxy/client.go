package proxy

import (
	"crypto/tls"
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/protocol"
	"github.com/derickr/dbgp-tools/lib/server"
	"io"
	"net"
	"time"
)

const sleepTimeout = time.Millisecond * 50

type ServerHandler struct {
	connectionList *connections.ConnectionList
}

func NewServerHandler(connectionList *connections.ConnectionList) *ServerHandler {
	return &ServerHandler{connectionList: connectionList}
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
	fmt.Printf("  - Connecting to %s\n", clientConnection.FullAddress())
	client, err := connectToIDE(clientConnection)

	if err != nil {
		fmt.Printf("    - IDE not connected: %s\n", err)
		return err
	}

	defer func(closer io.Closer) {
		err := closer.Close()
		if err != nil {
			fmt.Printf("%v", err)
		}
	}(client)

	if err != nil {
		return err
	}

	fmt.Println("    - IDE connected")
	reassembledPacket := fmt.Sprintf("%d\000%s", len(initialPacket)-1, initialPacket)
	_, err = client.Write([]byte(reassembledPacket))
	if err != nil {
		return err
	}

	fmt.Println("    - Init forwarded, start pipe")
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
			fmt.Println("  - IDE closed connection")
			return nil
		case err = <-clientChan:
			fmt.Println("  - Xdebug connection closed")
			return nil
		default:
			time.Sleep(sleepTimeout)
		}
	}
}

func (handler *ServerHandler) Handle(conn net.Conn) error {
	reader := protocol.NewDbgpClient(conn, false)

	response, err, _ := reader.ReadResponse()
	if err != nil {
		return fmt.Errorf("Error reading response: %v", err)
	}

	init, _ := reader.ParseInitXML(response)

	client, ok := handler.connectionList.FindByKey(init.IDEKey)

	if ok {
		fmt.Printf("  - Found connection for IDE Key '%s': %s\n", init.IDEKey, client)
		handler.setupForwarder(conn, []byte(response), client)
	} else {
		return fmt.Errorf("Could not find connection for IDE Key '%s'", init.IDEKey)
	}

	return nil
}
