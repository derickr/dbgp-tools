package dbgpxml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
)

/*
<stream xmlns="urn:debugger_protocol_v1"
xmlns:xdebug="https://xdebug.org/dbgp/xdebug" type="stdout"
encoding="base64"><![CDATA[aW4gY2FsbGVkX2Z1bmN0aW9uCg==]]></stream>
*/
type Stream struct {
	XMLName     xml.Name `xml:"stream"`
	XmlNS       string   `xml:"xmlns,attr"`
	XmlNSXdebug string   `xml:"xdebug,attr"`
	Type        string   `xml:"type,attr"`
	Encoding    string   `xml:"encoding,attr"`
	Value       string   `xml:",cdata"`
}

func (stream Stream) IsSuccess() bool {
	return true
}

func (stream Stream) GetErrorMessage() string {
	return ""
}

func (stream Stream) ExpectMoreResponses() bool {
	return true
}

func (stream Stream) ShouldCloseConnection() bool {
	return false
}

func (stream Stream) String() string {
	output := fmt.Sprintf("%s\n", Bold(BrightYellow(stream.Type)))

	value := []byte(stream.Value)
	if stream.Encoding == "base64" {
		value, _ = base64.StdEncoding.DecodeString(string(value))
	}
	output += fmt.Sprintf("%s %s", Bold(BrightYellow("←")), BrightYellow(value))

	return output
}
