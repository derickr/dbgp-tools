// Go offers built-in support for XML and XML-like
// formats with the `encoding.xml` package.

package main

import (
    "encoding/xml"
    "fmt"
	"net"
	"os"
	"github.com/pborman/getopt/v2" // BSD-3
	"github.com/chzyer/readline" // MIT
	"strings"
)

// This type will be mapped to XML. Similarly to the
// JSON examples, field tags contain directives for the
// encoder and decoder. Here we use some special features
// of the XML package: the `XMLName` field name dictates
// the name of the XML element representing this struct;
// `id,attr` means that the `Id` field is an XML
// _attribute_ rather than a nested element.

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
	XmlNSXdebug string `xml:"xmlns:xdebug,attr"`
    FileURI string   `xml:"fileurl,attr"`
    Language string   `xml:"language,attr"`
    LanguageVersion string   `xml:"xdebug:language_version,attr"`
    ProtocolVersion string   `xml:"protocol_version,attr"`
	Engine          Engine   `xml:"engine"`
	Author          Author   `xml:"author"`
	URL             URL   `xml:"url"`
	Copyright       Copyright   `xml:"copyright"`
}

/*
<response xmlns="urn:debugger_protocol_v1"
xmlns:xdebug="https://xdebug.org/dbgp/xdebug" command="feature_set"
transaction_id="1" feature="resolved_breakpoints" success="1"></response>
*/
type Response struct {
    XMLName xml.Name `xml:"init"`
	XmlNS   string   `xml:"xmlns,attr"`
	XmlNSXdebug string `xml:"xmlns:xdebug,attr"`
    TID  string   `xml:"transaction_id,attr"`
    Command string   `xml:"command,attr,omitempty"`
    Success string   `xml:"success,attr,omitempty"`
	Feature string   `xml:"feature,attr,omitempty"`
}

func handleConnection(c net.Conn, rl *readline.Instance) {
	fmt.Printf("Connect from %s\n", c.RemoteAddr().String())

	buf := make([]byte, 1024)
	for {
		_, err := c.Read(buf)

		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}

		/* Strip out everything up until the first leading \0 */
		initial := strings.IndexByte(string(buf), 0)
		xml := string(buf[initial+1:])
		final := strings.IndexByte(xml, 0)
		fmt.Printf("%s\n", xml[:final])

		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}

		_, err = c.Write([]byte(line))
		if err != nil {
			fmt.Println("Error writing:", err.Error())
			break
		}

		_, err = c.Write([]byte("\000"))
		if err != nil {
			fmt.Println("Error writing:", err.Error())
			break
		}
	}
	c.Close()
	fmt.Printf("Disconnect\n")
}

var (
	help = false
	once = false
	port = 9000
	version = false
)

func printStartUp() {
	fmt.Println("Xdebug Simple DBGp client (Go 0.1)")
	fmt.Println("Copyright 2019-2020 by Derick Rethans")
/*
	if haveLibEdit {
		fmt.Println("- libedit support: enabled")
	}
*/
}

func handleArguments() {
	getopt.Flag(&help, 'h', "Show this help")
	getopt.Flag(&port, 'p', "Specify the port to listen on")
	getopt.Flag(&version, 'v', "Show version number and exit")
	getopt.Flag(&once, '1', "Debug once and then exit")

	getopt.Parse()

	if help {
		getopt.PrintUsage(os.Stdout)
		os.Exit(1)
	}
	if version {
		os.Exit(0)
	}
}

func main() {
	printStartUp()
	handleArguments()

	portString := fmt.Sprintf(":%v", port)
	l, err := net.Listen("tcp", portString)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	defer l.Close()

	fmt.Printf("\nWaiting for debug server to connect on port %d.\n", port)

	rl, err := readline.New("(cmd) ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		handleConnection(c, rl)

		if once {
			break
		}
	}
}
