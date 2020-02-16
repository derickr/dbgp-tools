package proxy

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib"
	"github.com/derickr/dbgp-tools/lib/connections"
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

func (handler *ServerHandler) setupForwarder(conn net.Conn, initialPacket []byte, clientConnection *connections.Connection) error {
	fmt.Printf("  - Connecting to %s\n", clientConnection.FullAddress())
	client, err := net.Dial("tcp", clientConnection.FullAddress())

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
	go func() {
		_, err = io.Copy(client, conn)
		clientChan <- err
	}()

	go func() {
		_, err = io.Copy(conn, client)
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
	reader := dbgp.NewDbgpClient(conn)

	response, err := reader.ReadResponse()
	if err != nil {
		fmt.Printf("Error reading command: %v\n", err)
		return err
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
