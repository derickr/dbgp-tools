package proxy

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib"
	"github.com/derickr/dbgp-tools/lib/connections"
	"net"
)

type ClientHandler struct {
	connectionList *connections.ConnectionList
}

func NewClientHandler(connectionList *connections.ConnectionList) *ClientHandler {
	return &ClientHandler{connectionList: connectionList}
}

func (handler *ClientHandler) Handle(conn net.Conn) error {
	reader := dbgp.NewDbgpServer(conn, handler.connectionList)

	cmd, err := reader.ReadCommand()
	if err != nil {
		fmt.Printf("Error reading command: %v\n", err)
		return err
	}

	xml, err := cmd.Handle()

	if err != nil {
		return err
	}

	return reader.SendResponse(xml)
}
