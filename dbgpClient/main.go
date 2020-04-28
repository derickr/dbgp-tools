package main

import (
	"crypto/tls"
	"fmt"
	"github.com/bitbored/go-ansicon" // BSD-3
	"github.com/chzyer/readline"     // MIT
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/protocol"
	"github.com/derickr/dbgp-tools/lib/server"
	. "github.com/logrusorgru/aurora" // WTFPL
	"github.com/pborman/getopt/v2"    // BSD-3
	"net"
	"os"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
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

func isValidXml(xml string) bool {
	return strings.HasPrefix(xml, "<?xml")
}

func handleConnection(c net.Conn, rl *readline.Instance) error {
	var lastCommand string

	reader := protocol.NewDbgpClient(c, smartClient, logger)

	if smartClient {
		setupSignalHandler(reader)
		defer signal.Reset()
	}

	for {
		var formattedResponse protocol.Response

		response, err, timedOut := reader.ReadResponse()

		if timedOut {
			if reader.HasAbortBeenSignalled() {
				return nil
			}
			if reader.HasCommandsToRun() {
				err = nil // set err to nil, so it resets the timeout
				goto ReadInput
			}
			continue
		}

		if err != nil { // reading failed
			return err
		}

		if !isValidXml(response) {
			return fmt.Errorf("The received XML is not valid, closing connection: %s", response)
		}

		if showXML {
			fmt.Fprintf(output, "%s\n", Faint(response))
		}

		formattedResponse = reader.FormatXML(response)

		if formattedResponse == nil {
			return fmt.Errorf("Could not interpret XML, closing connection.")
		}
		fmt.Fprintln(output, formattedResponse)

		if formattedResponse.ExpectMoreResponses() {
			if !formattedResponse.IsSuccess() {
				return fmt.Errorf("Another response expected, but it wasn't a successful response")
			}
			continue
		}

		if formattedResponse.ShouldCloseConnection() {
			fmt.Fprintf(output, "%s\n", BrightRed("The connection should be closed."))
			return nil
		}

	ReadInput:
		line, found := reader.GetNextCommand()

		if !found { // if there was no command in the queue, read from the command line
			line, err = rl.Readline()
		}

		if err != nil { // io.EOF
			return err
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
			return err
		}

		lastCommand = line
	}

	return nil
}

var (
	cloudUser   = ""
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
	logger      = server.NewConsoleLogger(output)
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

	if cloudUser == "" {
		handleProxyArguments()
	}
	handleCloudArguments()
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("breakpoint_get -d"),
	readline.PcItem("breakpoint_list"),
	readline.PcItem("breakpoint_remove -d"),
	readline.PcItem("breakpoint_get -d"),
	readline.PcItem("breakpoint_set",
		readline.PcItem("-t line",
			readline.PcItem("-f"),
			readline.PcItem("-n"),
		),
		readline.PcItem("-t conditional",
			readline.PcItem("-f"),
			readline.PcItem("-n"),
			readline.PcItem("--"),
		),
		readline.PcItem("-t call",
			readline.PcItem("-a"),
			readline.PcItem("-m"),
		),
		readline.PcItem("-t return",
			readline.PcItem("-a"),
			readline.PcItem("-m"),
		),
		readline.PcItem("-t exception",
			readline.PcItem("-x"),
		),
		readline.PcItem("-t watch"),
		readline.PcItem("-h"),
		readline.PcItem("-o >="),
		readline.PcItem("-o =="),
		readline.PcItem("-o %"),
		readline.PcItem("-s enabled"),
		readline.PcItem("-s disabled"),
	),
	readline.PcItem("breakpoint_update -d",
		readline.PcItem("-n"),
		readline.PcItem("-h"),
		readline.PcItem("-o >="),
		readline.PcItem("-o =="),
		readline.PcItem("-o %"),
		readline.PcItem("-s enabled"),
		readline.PcItem("-s disabled"),
	),

	readline.PcItem("context_get",
		readline.PcItem("-c"),
		readline.PcItem("-d"),
	),
	readline.PcItem("context_names"),

	readline.PcItem("eval",
		readline.PcItem("-p"),
		readline.PcItem("--"),
	),
	readline.PcItem("feature_get -n",
		readline.PcItem("breakpoint_languages"),
		readline.PcItem("breakpoint_types"),
		readline.PcItem("data_encoding"),
		readline.PcItem("encoding"),
		readline.PcItem("extended_properties"),
		readline.PcItem("language_name"),
		readline.PcItem("language_supports_threads"),
		readline.PcItem("language_version"),
		readline.PcItem("max_children"),
		readline.PcItem("max_data"),
		readline.PcItem("max_depth"),
		readline.PcItem("notify_ok"),
		readline.PcItem("protocol_version"),
		readline.PcItem("resolved_breakpoints"),
		readline.PcItem("show_hidden"),
		readline.PcItem("supported_encodings"),
		readline.PcItem("supports_async"),
		readline.PcItem("supports_postmortem"),
	),
	readline.PcItem("feature_set -n",
		readline.PcItem("encoding -v"),
		readline.PcItem("extended_properties -v"),
		readline.PcItem("max_children -v"),
		readline.PcItem("max_data -v"),
		readline.PcItem("max_depth -v"),
		readline.PcItem("notify_ok -v"),
		readline.PcItem("resolved_breakpoints -v"),
		readline.PcItem("show_hidden -v"),
	),

	readline.PcItem("typemap_get"),
	readline.PcItem("property_get",
		readline.PcItem("-c"),
		readline.PcItem("-d"),
		readline.PcItem("-m"),
		readline.PcItem("-n"),
		readline.PcItem("-p"),
	),
	readline.PcItem("property_set",
		readline.PcItem("-c"),
		readline.PcItem("-d"),
		readline.PcItem("-n"),
		readline.PcItem("-p"),
		readline.PcItem("--"),
	),
	readline.PcItem("property_value",
		readline.PcItem("-c"),
		readline.PcItem("-d"),
		readline.PcItem("-n"),
		readline.PcItem("-p"),
	),

	readline.PcItem("source",
		readline.PcItem("-f"),
		readline.PcItem("-b"),
		readline.PcItem("-e"),
	),
	readline.PcItem("stack_depth"),
	readline.PcItem("stack_get",
		readline.PcItem("-d"),
	),
	readline.PcItem("status"),

	readline.PcItem("stderr"),
	readline.PcItem("stdout -c",
		readline.PcItem("0"),
		readline.PcItem("1"),
	),

	readline.PcItem("run"),
	readline.PcItem("step_into"),
	readline.PcItem("step_out"),
	readline.PcItem("step_over"),

	readline.PcItem("stop"),
	readline.PcItem("detach"),

	readline.PcItem("help"),
)

func initReadline() *readline.Instance {
	usr, _ := user.Current()
	dir := usr.HomeDir

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("%s", Bold("(cmd) ")),
		Stdout:          output,
		HistoryFile:     dir + "/.xdebug-debugclient.hist",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
	})
	if err != nil {
		panic(err)
	}

	return rl
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

	err = handleConnection(c, rl)
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

func runAsCloudClient(logger server.Logger) {
	conn, err := connections.ConnectToCloud(CloudDomain, CloudPort, cloudUser, logger)

	if err != nil {
		fmt.Fprintf(output, "%s '%s': %s\n", BrightRed("Can not connect to Xdebug cloud at"), BrightYellow(CloudDomain), BrightRed(err))
		return
	}
	defer conn.Close()
	defer fmt.Fprintf(output, "Disconnect\n")

	protocol := protocol.NewDbgpClient(conn, false, logger)

	rl := initReadline()
	defer rl.Close()

	command := "cloudinit -u " + cloudUser
	protocol.SendCommand(command)

	for {
		err = handleConnection(conn, rl)
		if err != nil {
			fmt.Fprintf(output, "%s: %s\n", BrightRed("Error while handling connection"), BrightRed(err.Error()))
			break
		}

		fmt.Fprintf(output, "\n%s\n", BrightGreen(Bold("Waiting for incoming connection...")))
	}
}

func main() {
	handleArguments()
	printStartUp()

	logger := server.NewConsoleLogger(os.Stdout)

	if cloudUser != "" {
		runAsCloudClient(logger)
	} else {
		runAsNormalClient()
	}
}
