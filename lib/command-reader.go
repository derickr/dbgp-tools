package dbgp

import (
	"bufio"
	"fmt"
	"github.com/xdebug/dbgp-tools/lib/command"
	"github.com/xdebug/dbgp-tools/lib/connections"
	"io"
	"net"
	"strconv"
	"strings"
)

type commandReader struct {
	connection     net.Conn
	connectionList *connections.ConnectionList
	reader         *bufio.Reader
	writer         io.Writer
}

func NewCommandReader(c net.Conn, connectionList *connections.ConnectionList) *commandReader {
	var tmp commandReader

	tmp.connection = c
	tmp.connectionList = connectionList
	tmp.reader = bufio.NewReader(c)
	tmp.writer = c

	return &tmp
}

func (dbgp *commandReader) parseLine(data string) (command.DbgpCommand, error) {
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

func (dbgp *commandReader) ReadCommand() (command.DbgpCommand, error) {
	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		fmt.Println("Error reading data:", err.Error())
		return nil, err
	}

	return dbgp.parseLine(strings.TrimRight(string(data), "\000"))
}

func (dbgp *commandReader) SendResponse(xml string) error {
	_, err := dbgp.writer.Write([]byte(strconv.Itoa(len(xml)) + "\000" + xml + "\000"))

	return err
}
