package dbgp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

type dbgpReader struct {
	reader  *bufio.Reader
	writer  io.Writer
	counter int
}

func NewDbgpReader(c net.Conn) *dbgpReader {
	var tmp dbgpReader
	tmp.reader = bufio.NewReader(c)
	tmp.writer = c
	tmp.counter = 1

	return &tmp
}

func (dbgp *dbgpReader) ReadResponse() (string, error) {
	/* Read length */
	_, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		fmt.Println("Error reading length:", err.Error())
		return "", err
	}

	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000')

	if err != nil {
		fmt.Println("Error reading data:", err.Error())
		return "", err
	}

	return string(data), nil
}

func (dbgp *dbgpReader) injectIIfNeeded(parts []string) []string {
	for _, item := range parts {
		if item == "-i" {
			return parts
		}
	}

	var newParts []string
	newParts = append(newParts, parts[0])
	newParts = append(newParts, "-i", fmt.Sprintf("%d", dbgp.counter))
	newParts = append(newParts, parts[1:]...)

	dbgp.counter++

	return newParts
}

func (dbgp *dbgpReader) processLine(line string) string {
	parts := strings.Split(strings.TrimSpace(line), " ")

	parts = dbgp.injectIIfNeeded(parts)

	return strings.Join(parts, " ")
}

func (dbgp *dbgpReader) SendCommand(line string) error {
	line = dbgp.processLine(line);

	_, err := dbgp.writer.Write([]byte(line))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return err
	}

	_, err = dbgp.writer.Write([]byte("\000"))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return err
	}

	return nil
}
