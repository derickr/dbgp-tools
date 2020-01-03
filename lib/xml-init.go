package dbgp

import (
    "encoding/xml"
	"golang.org/x/net/html/charset"
	. "github.com/logrusorgru/aurora" // WTFPL
    "fmt"
	"strings"
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
type Engine struct {
	XMLName xml.Name `xml:"engine"`
	Version string   `xml:"version,attr"`
	Value   string   `xml:",cdata"`
}
type Author struct {
	XMLName xml.Name `xml:"author"`
	Value   string   `xml:",cdata"`
}
type URL struct {
	XMLName xml.Name `xml:"url"`
	Value   string   `xml:",cdata"`
}
type Copyright struct {
	XMLName xml.Name `xml:"copyright"`
	Value   string   `xml:",cdata"`
}
type Init struct {
    XMLName xml.Name `xml:"init"`
	XmlNS   string   `xml:"xmlns,attr"`
	XmlNSXdebug string `xml:"xdebug,attr"`
    FileURI string   `xml:"fileuri,attr"`
    Language string   `xml:"language,attr"`
    LanguageVersion string   `xml:"language_version,attr"`
    ProtocolVersion string   `xml:"protocol_version,attr"`
    AppID string   `xml:"appid,attr"`
    IDEKey string   `xml:"idekey,attr"`
	Engine          Engine   `xml:"engine"`
	Author          Author   `xml:"author"`
	URL             URL   `xml:"url"`
	Copyright       Copyright   `xml:"copyright"`
}

func ParseInitXML(rawXmlData string) (Init, error) {
	init := Init{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&init)

	if err != nil {
		return init, err
	}

	return init, nil
}

func (init Init) String() string {
	return fmt.Sprintf("DBGp/%s: %s %s — For %s %s\nDebugging %v (ID: %s/%s)",
		Bold(Green(init.ProtocolVersion)),
		Bold("Xdebug"), Bold(Green(init.Engine.Version)),
		Bold(init.Language), Bold(Green(init.LanguageVersion)),
		Bold(BrightYellow(init.FileURI)),
		BrightYellow(init.AppID), BrightYellow(init.IDEKey))
}
