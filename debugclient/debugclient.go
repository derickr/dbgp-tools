// Go offers built-in support for XML and XML-like
// formats with the `encoding.xml` package.

package main

import (
	"fmt"
	"github.com/chzyer/readline"      // MIT
	. "github.com/logrusorgru/aurora" // WTFPL
	"github.com/pborman/getopt/v2"    // BSD-3
	"github.com/xdebug/dbgp-tools/lib"
	"net"
	"os"
)

func formatXML(rawXmlData string) {
	response, err := dbgp.ParseResponseXML(rawXmlData)

	if err == nil {
		fmt.Println(response)
		return
	} else {
		fmt.Println(err)
	}

	init, err := dbgp.ParseInitXML(rawXmlData)

	if err == nil {
		fmt.Println(init)
		return
	} else {
		fmt.Println(err)
	}
}

func handleConnection(c net.Conn, rl *readline.Instance) {
	var lastCommand string

	fmt.Printf("Connect from %s\n", c.RemoteAddr().String())

	for {

		response, err := dbgp.ReadResponse(c)
		if err != nil { // reading failed
			break
		}
		fmt.Printf("%s\n", response)
		formatXML(response)

		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}

		if line == "" {
			line = lastCommand
		}

		err = dbgp.SendCommand(c, line)
		if err != nil {
			break
		}

		lastCommand = line
	}
	c.Close()
	fmt.Printf("Disconnect\n")
}

var (
	help    = false
	once    = false
	port    = 9000
	version = false
)

func printStartUp() {
	fmt.Println("Xdebug Simple DBGp client (Go 0.1)")
	fmt.Println("Copyright 2019-2020 by Derick Rethans")
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

	rl, err := readline.New(fmt.Sprintf("%s", Bold("(cmd) ")))
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
