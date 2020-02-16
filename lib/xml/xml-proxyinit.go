package dbgpXml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
	"strconv"
)

/*
<init xmlns="urn:debugger_protocol_v1"
xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
fileuri="file:///home/derick/dev/php/derickr-xdebug/tests/debugger/bug01727.inc"
language="PHP" xdebug:language_version="7.4.0-dev" protocol_version="1.0"
appid="105446" idekey="dr"><engine
version="2.9.1-dev"><![CDATA[Xdebug]]></engine><author><![CDATA[Derick
Rethans]]></author><url><![CDATA[https://xdebug.org]]></url><copyright><![CDATA[Copyright
(c) 2002-2019 by Derick Rethans]]></copyright></init>
*/
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
	Error       *ProxyInitError `xml:"error,omitempty"`
}

func NewProxyInit(success bool, ideKey string, address string, port int, initError *ProxyInitError) *ProxyInit {
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

func (init ProxyInit) String() string {
	if init.Success == 0 {
		return fmt.Sprintf("%s | %s: %s\n", Yellow(Bold("proxyinit")), Bold(Red("failure")), BrightRed(init.Error.Message))
	} else {
		return fmt.Sprintf("%s | %s\n", Yellow(Bold("proxyinit")), Bold(Green("success")))
	}
}
