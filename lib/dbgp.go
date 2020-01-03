package dbgp

import (
    "fmt"
	"net"
	"strings"
)

func ReadResponse(c net.Conn) (string, error) {
	buf := make([]byte, 2048)

	_, err := c.Read(buf)

	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return "", err
	}

	/* Strip out everything up until the first leading \0 */
	initial := strings.IndexByte(string(buf), 0)
	xml := string(buf[initial+1:])
	final := strings.IndexByte(xml, 0)

	if final == -1 {
		fmt.Println("Error reading: couldn't find end '\\0'")
		return "", err
	}

	return xml[:final], nil
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
