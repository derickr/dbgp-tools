package dbgpxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
	"strconv"
)

type ProxyInitError struct {
	XMLName xml.Name `xml:"error"`
	ID      string   `xml:"id,attr"`
	Message string   `xml:"message"`
}
type ProxyInit struct {
	XMLName     xml.Name        `xml:"proxyinit"`
	XmlNS       string          `xml:"xmlns,attr"`
	XmlNSXdebug string          `xml:"xmlns:xdebug,attr"`
	Success     int             `xml:"success,attr"`
	IDEKey      string          `xml:"idekey,attr"`
	Address     string          `xml:"address,attr"`
	Port        string          `xml:"port,attr"`
	SSL         bool            `xml:"ssl,attr"`
	Error       *ProxyInitError `xml:"error,omitempty"`
}

func NewProxyInit(success bool, ideKey string, address string, port int, ssl bool, initError *ProxyInitError) *ProxyInit {
	successStr := 1
	if !success {
		successStr = 0
	}

	return &ProxyInit{
		XmlNS:       "urn:debugger_protocol_v1",
		XmlNSXdebug: "https://xdebug.org/dbgp/xdebug",
		Success:     successStr,
		IDEKey:      ideKey,
		Address:     address,
		Port:        strconv.Itoa(port),
		SSL:         ssl,
		Error:       initError,
	}
}

func (proxyInit *ProxyInit) AsXML() (string, error) {
	var output bytes.Buffer

	encoder := xml.NewEncoder(&output)

	err := encoder.Encode(proxyInit)

	if err != nil {
		return "", err
	}

	return xml.Header + output.String(), nil
}

func (init ProxyInit) IsSuccess() bool {
	return !!(init.Success == 1)
}

func (init ProxyInit) GetErrorMessage() string {
	if init.Error != nil {
		return init.Error.Message
	} else {
		return "no error"
	}
}

func (init ProxyInit) ExpectMoreResponses() bool {
	return false
}

func (init ProxyInit) ShouldCloseConnection() bool {
	return false
}

func (init ProxyInit) String() string {
	if init.Success == 0 {
		return fmt.Sprintf("%s | %s: %s\n", Yellow(Bold("proxyinit")), Bold(Red("failure")), BrightRed(init.Error.Message))
	} else {
		return fmt.Sprintf("%s | %s\n", Yellow(Bold("proxyinit")), Bold(Green("success")))
	}
}
