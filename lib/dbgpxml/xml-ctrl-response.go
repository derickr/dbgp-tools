package dbgpxml

import (
//	"encoding/base64"
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
//	"strings"
)

/*
<ps>
	<engine version="3.3.0alpha4-dev"><![CDATA[Xdebug]]></engine>
	<filename><![CDATA[/tmp/test.php]]></filename>
	<pid><![CDATA[1943117]]></pid>
</ps>
*/
type PS struct {
	XMLName         xml.Name  `xml:"ps"`
	FileURI         string    `xml:"fileuri,attr"`
	Engine          Engine    `xml:"engine"`
	PID             string    `xml:"pid"`
	FileUri         string    `xml:"fileuri"`
	Time float64 `xml:"time,omitempty"`
	Memory int64 `xml:"memory,omitempty"`
}

type CtrlResponse struct {
	XMLName     xml.Name  `xml:"ctrl-response"`
	XmlNSXdebug string    `xml:"xmlns:xdebug-ctrl,attr"`
	PS PS `xml:"ps,omitempty"`
	Error *Error `xml:"error,omitempty"`

	Value string `xml:",cdata"`
}

func formatPS(ctrlResponse CtrlResponse) string {
	return fmt.Sprintf("%s | %s\n", Red(ctrlResponse.PS), Bold(Green("OK")))
}

func (ctrlResponse CtrlResponse) IsSuccess() bool {
	return true
}

func (ctrlResponse CtrlResponse) formatError() string {
	return fmt.Sprintf("%s(%d): %s\n", Bold(Red("Error")), Red(ctrlResponse.Error.Code), BrightRed(Bold(ctrlResponse.Error.Message.Text)))
}

func (ctrlResponse CtrlResponse) GetErrorMessage() string {
	if ctrlResponse.Error != nil {
		return ctrlResponse.Error.Message.Text
	} else {
		return "no error"
	}
}

func (ctrlResponse CtrlResponse) ExpectMoreResponses() bool {
	return false
}

func (ctrlResponse CtrlResponse) ShouldCloseConnection() bool {
	return false
}

func (ctrlResponse CtrlResponse) String() string {
	output := ""

	if ctrlResponse.Error != nil && ctrlResponse.Error.Code != 0 {
		return output + ctrlResponse.formatError()
	}

	output += fmt.Sprintf("%10s %8d %8.2f %s %s",
		Faint(ctrlResponse.PS.PID),
		ctrlResponse.PS.Memory,
		ctrlResponse.PS.Time,
		BrightWhite(ctrlResponse.PS.FileUri),
		Faint("(" + ctrlResponse.PS.Engine.Value + ctrlResponse.PS.Engine.Version + ")"))

	return output
}
