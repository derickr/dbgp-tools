package dbgp

import (
	"fmt"
	"io"
	"bufio"
	"net"
)

type dbgpReader struct {
	reader *bufio.Reader
}

func NewDbgpReader(reader io.Reader) *dbgpReader {
	var tmp dbgpReader
	tmp.reader = bufio.NewReader(reader)

	return &tmp
}

func (dbgp *dbgpReader) ReadResponse() (string, error) {
	/* Read length */
	_, err := dbgp.reader.ReadBytes('\000');

	if err != nil {
		fmt.Println("Error reading length:", err.Error())
		return "", err
	}

	/* Read data */
	data, err := dbgp.reader.ReadBytes('\000');

	if err != nil {
		fmt.Println("Error reading data:", err.Error())
		return "", err
	}

	return string(data), nil
}

func SendCommand(c net.Conn, line string) error {
	_, err := c.Write([]byte(line))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return err
	}

	_, err = c.Write([]byte("\000"))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return err
	}

	return nil
}
