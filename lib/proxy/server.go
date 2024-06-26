package proxy

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/protocol"
	"net"
)

type ClientHandler struct {
	logger         logger.Logger
	connectionList *connections.ConnectionList
}

func NewClientHandler(connectionList *connections.ConnectionList, logger logger.Logger) *ClientHandler {
	return &ClientHandler{connectionList: connectionList, logger: logger}
}

func (handler *ClientHandler) Handle(conn net.Conn) error {
	reader := protocol.NewDbgpServer(conn, handler.connectionList, handler.logger)

	cmd, err := reader.ReadCommand()
	if err != nil {
		return fmt.Errorf("Error reading command: %v", err)
	}

	xml, err := cmd.Handle()

	if err != nil {
		return err
	}

	return reader.SendResponse(xml)
}
