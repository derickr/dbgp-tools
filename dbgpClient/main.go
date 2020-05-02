package main

import (
	"crypto/tls"
	"fmt"
	"github.com/bitbored/go-ansicon" // BSD-3
	"github.com/chzyer/readline"     // MIT
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/protocol"
	"github.com/derickr/dbgp-tools/lib/xml"
	. "github.com/logrusorgru/aurora" // WTFPL
	"github.com/pborman/getopt/v2"    // BSD-3
	"net"
	"os"
	"os/signal"
)

var clientVersion = "0.2"

func displayHelp() {
	fmt.Fprintf(output, `
This is a DBGp client. DBGp is a common debugging protocol described at
https://xdebug.org/docs/dbgp

The client reads DBGp commands on the command line, sends them to the
DBGp debugging engine, reads the XML response, and formats that response
by interpreting the XML.

A short overview of commands is also available in the online
documentation at https://xdebug.org/docs/debugclient

You can use <tab> for auto completing commands, and find out which one
exist.
`)
}

type CommandRunner interface {
	AddCommandToRun(string)
	IsInConversation() bool
	SignalAbort()
}

func setupSignalHandler(protocol CommandRunner) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			if protocol.IsInConversation() {
				protocol.AddCommandToRun("break")
			} else {
				protocol.SignalAbort()
			}
		}
	}()
}

func handleConnection(c net.Conn, rl *readline.Instance) (bool, error) {
	var lastCommand string

	reader := protocol.NewDbgpClient(c, smartClient, logOutput)

	setupSignalHandler(reader)
	defer signal.Reset()

	for {
		var formattedResponse protocol.Response

		response, err, timedOut := reader.ReadResponse()

		if timedOut {
			if reader.HasAbortBeenSignalled() {
				return true, nil
			}
			if reader.HasCommandsToRun() {
				err = nil // set err to nil, so it resets the timeout
				goto ReadInput
			}
			continue
		}

		if err != nil { // reading failed
			return false, err
		}

		if showXML {
			fmt.Fprintf(output, "%s\n", Faint(response))
		}

		if !dbgpXml.IsValidXml(response) {
			return false, fmt.Errorf("The received XML is not valid, closing connection: %s", response)
		}

		formattedResponse = reader.FormatXML(response)

		if formattedResponse == nil {
			return false, fmt.Errorf("Could not interpret XML, closing connection.")
		}
		fmt.Fprintln(output, formattedResponse)

		if formattedResponse.ExpectMoreResponses() {
			if !formattedResponse.IsSuccess() {
				return false, fmt.Errorf("Another response expected, but it wasn't a successful response")
			}
			continue
		}

		if formattedResponse.ShouldCloseConnection() {
			fmt.Fprintf(output, "%s\n", BrightRed("The connection should be closed."))
			return false, nil
		}

	ReadInput:
		line, found := reader.GetNextCommand()

		if !found { // if there was no command in the queue, read from the command line
			line, err = rl.Readline()
		}

		if err != nil { // io.EOF
			return false, err
		}

		if line == "help" {
			displayHelp()
			goto ReadInput
		}

		if line == "" {
			line = lastCommand
		}

		err = reader.SendCommand(line)
		if err != nil {
			return false, err
		}

		lastCommand = line
	}

	return false, nil
}

var (
	cloudUser   = ""
	disCloudUser   = ""
	CloudDomain = "cloud.xdebug.com"
	CloudPort   = "9021"
	help        = false
	once        = false
	port        = 9000
	proxy       = "localhost:9001"
	register    = ""
	showXML     = false
	smartClient = false
	ssl         = false
	sslPort     = 9010
	sslProxy    = "localhost:9011"
	version     = false
	unregister  = ""
	output      = ansicon.Convert(os.Stdout)
	logOutput   = logger.NewConsoleLogger(output)
)

func printStartUp() {
	fmt.Fprintf(output, "Xdebug Simple DBGp client (%s)\n", Bold(clientVersion))
	fmt.Fprintf(output, "Copyright 2019-2020 by Derick Rethans\n")

	if !smartClient {
		fmt.Fprintf(output, "In dumb client mode\n")
	}

	fmt.Fprintf(output, "\n")
}

func handleArguments() {
	getopt.Flag(&help, 'h', "Show this help")
	getopt.Flag(&port, 'p', "Specify the port to listen on")
	getopt.Flag(&smartClient, 'f', "Whether act as fully featured DBGp client")
	getopt.Flag(&version, 'v', "Show version number and exit")
	getopt.Flag(&showXML, 'x', "Show protocol XML")
	getopt.Flag(&once, '1', "Debug once and then exit")

	handleProxyFlags()
	handleCloudFlags()

	getopt.Parse()

	if help {
		getopt.PrintUsage(os.Stdout)
		os.Exit(1)
	}
	if version {
		os.Exit(0)
	}

	if cloudUser == "" && disCloudUser == "" {
		handleProxyArguments()
	}
	handleCloudArguments()
}

func accept(l net.Listener) (net.Conn, error) {
	c, err := l.Accept()

	if err != nil {
		return nil, err
	}

	if ssl {
		cert, err := tls.LoadX509KeyPair("certs/fullchain.pem", "certs/privkey.pem")
		if err != nil {
			fmt.Printf("server: loadkeys: %s", err)
			panic(err)
		}
		config := tls.Config{Certificates: []tls.Certificate{cert}}

		return tls.Server(c, &config), nil
	} else {
		return c, nil
	}
}

func doNormalConnectionLoop(l net.Listener, rl *readline.Instance) {
	c, err := accept(l)
	if err != nil {
		fmt.Fprintln(output, err)
		return
	}

	defer c.Close()
	defer fmt.Fprintf(output, "Disconnect\n")

	fmt.Fprintf(output, "Connect from %s\n", c.RemoteAddr().String())

	_, err = handleConnection(c, rl)
	if err != nil {
		fmt.Fprintf(output, "%s: %s\n", BrightRed("Error while handling connection"), BrightRed(err.Error()))
	}
}

func runAsNormalClient() {
	portString := fmt.Sprintf(":%v", port)
	l, err := net.Listen("tcp", portString)
	if err != nil {
		fmt.Fprintf(output, "%v", err)
		return
	}
	defer l.Close()

	fmt.Fprintf(output, "Waiting for debug server to connect on port %d.\n", port)

	rl := initReadline()
	defer rl.Close()

	for {
		doNormalConnectionLoop(l, rl)

		if once {
			break
		}
	}
}

func runAsCloudClient(logger logger.Logger) {
	conn, err := connections.ConnectToCloud(CloudDomain, CloudPort, cloudUser, logger)

	if err != nil {
		fmt.Fprintf(output, "%s '%s': %s\n", BrightRed("Can not connect to Xdebug Cloud at"), BrightYellow(CloudDomain), BrightRed(err))
		return
	}
	defer conn.Close()
	defer fmt.Fprintf(output, "\n%s\n", BrightYellow(Bold("Shutting down client")))

	protocol := protocol.NewDbgpClient(conn, false, logger)

	rl := initReadline()
	defer rl.Close()

	command := "cloudinit -u " + cloudUser
	protocol.SendCommand(command)

	for {
		abortClient, err := handleConnection(conn, rl)
		if err != nil {
			if err == readline.ErrInterrupt {
				fmt.Fprintf(output, "%s: %s\n", BrightYellow("Interrupt, sending detach"), BrightRed(err.Error()))
				command := "detach"
				protocol.SendCommand(command)
			} else {
				fmt.Fprintf(output, "%s: %s\n", BrightRed("Error while handling connection"), BrightRed(err.Error()))
				break
			}
		}

		if abortClient {
			command = "cloudstop -u " + cloudUser
			protocol.SendCommand(command)
			protocol.ReadResponse()
			return
		}

		fmt.Fprintf(output, "\n%s\n", BrightGreen(Bold("Waiting for incoming connection...")))
	}
}

func main() {
	handleArguments()
	printStartUp()

	logger := logger.NewConsoleLogger(os.Stdout)

	if disCloudUser != "" {
		protocol.UnregisterCloudClient(CloudDomain, CloudPort, disCloudUser, output, logger)
		if cloudUser == "" {
			return
		}
	}
	if cloudUser != "" {
		runAsCloudClient(logger)
	} else {
		runAsNormalClient()
	}
}
