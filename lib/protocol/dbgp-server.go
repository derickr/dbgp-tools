package protocol

import (
	"bufio"
	"fmt"
	"github.com/derickr/dbgp-tools/lib/command"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	"io"
	"net"
	"strconv"
	"strings"
)

type DbgpServer struct {
	logger         logger.Logger
	connection     net.Conn
	connectionList *connections.ConnectionList
	reader         *bufio.Reader
	writer         io.Writer
}

func NewDbgpServer(c net.Conn, connectionList *connections.ConnectionList, logger logger.Logger) *DbgpServer {
	var tmp DbgpServer

	tmp.connection = c
	tmp.connectionList = connectionList
	tmp.logger = logger
	tmp.reader = bufio.NewReader(c)
	tmp.writer = c

	return &tmp
}

func (dbgp *DbgpServer) parseLine(data string) (command.DbgpCommand, error) {
	parts := strings.Split(data, " ")

	switch parts[0] {
	case "proxyinit":
		host, _, _ := net.SplitHostPort(dbgp.connection.RemoteAddr().String())
		return command.CreateProxyInit(host, dbgp.connectionList, parts[1:], dbgp.logger)

	case "proxystop":
		return command.CreateProxyStop(dbgp.connectionList, parts[1:], dbgp.logger)
	}

	// TODO(florin): '%s' can be replaced with %q and the formatting function will handle the quotes
	return nil, fmt.Errorf("Don't understand command '%s'", parts)
}

func (dbgp *DbgpServer) parseCloudLine(data string) (command.DbgpCloudCommand, error) {
	parts := strings.Split(data, " ")

	switch parts[0] {
	case "cloudinit":
		return command.CreateCloudInit(dbgp.connectionList, &dbgp.connection, parts[1:], dbgp.logger)
	case "cloudstop":
		return command.CreateCloudStop(dbgp.connectionList, &dbgp.connection, parts[1:], dbgp.logger)
	}

	return nil, fmt.Errorf("Don't understand command '%s'", parts)
}

// TODO(florin): I would extract this into a readCommand() method and deduplicate ReadCloudCommand
func (dbgp *DbgpServer) ReadCommand() (command.DbgpCommand, error) {
	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			dbgp.logger.LogWarning("dbgp-server", "I/O timeout reading data: %s", err.Error())
		} else {
			dbgp.logger.LogError("dbgp-server", "Error reading data: %s", err.Error())
		}
		return nil, err
	}

	return dbgp.parseLine(strings.TrimRight(string(data), "\000"))
}

func (dbgp *DbgpServer) ReadCloudCommand() (command.DbgpCloudCommand, error) {
	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			dbgp.logger.LogWarning("dbgp-server", "I/O timeout reading data: %s", err.Error())
		} else {
			dbgp.logger.LogError("dbgp-server", "Error reading data: %s", err.Error())
		}
		return nil, err
	}

	return dbgp.parseCloudLine(strings.TrimRight(string(data), "\000"))
}

func (dbgp *DbgpServer) SendResponse(xml string) error {
	_, err := dbgp.writer.Write([]byte(strconv.Itoa(len(xml)) + "\000" + xml + "\000"))

	return err
}
