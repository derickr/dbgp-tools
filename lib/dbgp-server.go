package dbgp

import (
	"bufio"
	"fmt"
	"github.com/derickr/dbgp-tools/lib/command"
	"github.com/derickr/dbgp-tools/lib/connections"
	"io"
	"net"
	"strconv"
	"strings"
)

type dbgpServer struct {
	connection     net.Conn
	connectionList *connections.ConnectionList
	reader         *bufio.Reader
	writer         io.Writer
}

func NewDbgpServer(c net.Conn, connectionList *connections.ConnectionList) *dbgpServer {
	var tmp dbgpServer

	tmp.connection = c
	tmp.connectionList = connectionList
	tmp.reader = bufio.NewReader(c)
	tmp.writer = c

	return &tmp
}

func (dbgp *dbgpServer) parseLine(data string) (command.DbgpCommand, error) {
	parts := strings.Split(data, " ")

	switch parts[0] {
	case "proxyinit":
		host, _, _ := net.SplitHostPort(dbgp.connection.RemoteAddr().String())
		return command.CreateProxyInit(host, dbgp.connectionList, parts[1:])

	case "proxystop":
		return command.CreateProxyStop(dbgp.connectionList, parts[1:])
	}

	return nil, fmt.Errorf("Don't understand command '%s'", parts)
}

func (dbgp *dbgpServer) parseInitLine(data string) (command.DbgpInitCommand, error) {
	parts := strings.Split(data, " ")

	switch parts[0] {
	case "cloudinit":
		return command.CreateCloudInit(dbgp.connectionList, &dbgp.connection, parts[1:])
	}

	return nil, fmt.Errorf("Don't understand command '%s'", parts)
}

func (dbgp *dbgpServer) ReadCommand() (command.DbgpCommand, error) {
	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		fmt.Println("Error reading data:", err.Error())
		return nil, err
	}

	return dbgp.parseLine(strings.TrimRight(string(data), "\000"))
}

func (dbgp *dbgpServer) ReadInitCommand() (command.DbgpInitCommand, error) {
	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		fmt.Println("Error reading data:", err.Error())
		return nil, err
	}

	return dbgp.parseInitLine(strings.TrimRight(string(data), "\000"))
}

func (dbgp *dbgpServer) SendResponse(xml string) error {
	_, err := dbgp.writer.Write([]byte(strconv.Itoa(len(xml)) + "\000" + xml + "\000"))

	return err
}
