package protocol

import (
	"bufio"
	"fmt"
	"github.com/derickr/dbgp-tools/lib/command"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/server"
	"io"
	"net"
	"strconv"
	"strings"
)

type DbgpServer struct {
	logger         server.Logger
	connection     net.Conn
	connectionList *connections.ConnectionList
	reader         *bufio.Reader
	writer         io.Writer
}

func NewDbgpServer(c net.Conn, connectionList *connections.ConnectionList, logger server.Logger) *DbgpServer {
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

	return nil, fmt.Errorf("Don't understand command '%s'", parts)
}

func (dbgp *DbgpServer) parseCloudInitLine(data string) (command.DbgpCloudInitCommand, error) {
	parts := strings.Split(data, " ")

	switch parts[0] {
	case "cloudinit":
		return command.CreateCloudInit(dbgp.connectionList, &dbgp.connection, parts[1:], dbgp.logger)
	}

	return nil, fmt.Errorf("Don't understand command '%s'", parts)
}

func (dbgp *DbgpServer) ReadCommand() (command.DbgpCommand, error) {
	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		dbgp.logger.LogError("conn", "Error reading data: %s", err.Error())
		return nil, err
	}

	return dbgp.parseLine(strings.TrimRight(string(data), "\000"))
}

func (dbgp *DbgpServer) ReadCloudInitCommand() (command.DbgpCloudInitCommand, error) {
	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		dbgp.logger.LogError("conn", "Error reading data: %s", err.Error())
		return nil, err
	}

	return dbgp.parseCloudInitLine(strings.TrimRight(string(data), "\000"))
}

func (dbgp *DbgpServer) SendResponse(xml string) error {
	_, err := dbgp.writer.Write([]byte(strconv.Itoa(len(xml)) + "\000" + xml + "\000"))

	return err
}