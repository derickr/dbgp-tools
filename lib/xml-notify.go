package dbgp

import (
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
	"golang.org/x/net/html/charset"
	"strings"
)

/*
<notify xmlns="urn:debugger_protocol_v1"
xmlns:xdebug="https://xdebug.org/dbgp/xdebug"
name="breakpoint_resolved"><breakpoint type="line" resolved="resolved"
filename="file:///tmp/xdebug-test.php" lineno="13" state="enabled"
hit_count="0" hit_value="0" id="161070001"></breakpoint></notify>
*/
type Notify struct {
	XMLName     xml.Name   `xml:"notify"`
	XmlNS       string     `xml:"xmlns,attr"`
	XmlNSXdebug string     `xml:"xdebug,attr"`
	Name        string     `xml:"name,attr"`
	Breakpoint  Breakpoint `xml:"breakpoint"`
}

func (dbgp *dbgpReader) parseNotifyXML(rawXmlData string) (Notify, error) {
	notify := Notify{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&notify)

	if err != nil {
		return notify, err
	}

	return notify, nil
}

func (notify Notify) String() string {
	output := fmt.Sprintf("%s\n", Bold(BrightYellow(notify.Name)))

	switch notify.Name {
	case "breakpoint_resolved":
		output += notify.Breakpoint.String()
	}

	return output
}
