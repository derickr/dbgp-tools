package dbgpXml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
)

/*
 */
type ProxyStop struct {
	XMLName     xml.Name        `xml:"proxystop"`
	XmlNS       string          `xml:"xmlns,attr"`
	XmlNSXdebug string          `xml:"xmlns:xdebug,attr"`
	Success     int             `xml:"success,attr"`
	IDEKey      string          `xml:"idekey,attr"`
	Error       *ProxyInitError `xml:"error,omitempty"`
}

func NewProxyStop(success bool, ideKey string, stopError *ProxyInitError) *ProxyStop {
	successStr := 1
	if !success {
		successStr = 0
	}

	return &ProxyStop{
		XmlNS:       "urn:debugger_protocol_v1",
		XmlNSXdebug: "https://xdebug.org/dbgp/xdebug",
		Success:     successStr,
		IDEKey:      ideKey,
		Error:       stopError,
	}
}

func (proxyStop *ProxyStop) AsXML() (string, error) {
	var output bytes.Buffer

	encoder := xml.NewEncoder(&output)

	err := encoder.Encode(proxyStop)

	if err != nil {
		return "", err
	}

	return xml.Header + output.String(), nil
}

func (stop ProxyStop) IsSuccess() bool {
	return !!(stop.Success == 1)
}

func (stop ProxyStop) ExpectMoreResponses() bool {
	return false
}

func (stop ProxyStop) ShouldCloseConnection() bool {
	return false
}

func (stop ProxyStop) String() string {
	if stop.Success == 0 {
		return fmt.Sprintf("%s | %s: %s\n", Yellow(Bold("proxystop")), Bold(Red("failure")), BrightRed(stop.Error.Message))
	} else {
		return fmt.Sprintf("%s | %s\n", Yellow(Bold("proxystop")), Bold(Green("success")))
	}
}

/*
func (init Init) String() string {
	return fmt.Sprintf("DBGp/%s: %s %s — For %s %s\nDebugging %v (ID: %s/%s)",
		Bold(Green(init.ProtocolVersion)),
		Bold("Xdebug"), Bold(Green(init.Engine.Version)),
		Bold(init.Language), Bold(Green(init.LanguageVersion)),
		Bold(BrightYellow(init.FileURI)),
		BrightYellow(init.AppID), BrightYellow(init.IDEKey))
}
*/
