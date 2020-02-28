package main

import (
	"crypto/tls"
	"fmt"
	"github.com/bitbored/go-ansicon" // BSD-3
	"github.com/chzyer/readline"     // MIT
	"github.com/derickr/dbgp-tools/lib"
	. "github.com/logrusorgru/aurora" // WTFPL
	"github.com/pborman/getopt/v2"    // BSD-3
	"net"
	"os"
	"os/signal"
	"os/user"
	"strconv"
)

var clientVersion = "0.1"

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
}

func setupSignalHandler(dbgp CommandRunner) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			dbgp.AddCommandToRun("break")
		}
	}()
}

func handleConnection(c net.Conn, rl *readline.Instance) {
	var lastCommand string

	reader := dbgp.NewDbgpClient(c, smartClient)

	if smartClient {
		setupSignalHandler(reader)
	}

	fmt.Fprintf(output, "Connect from %s\n", c.RemoteAddr().String())

	for {
		var formattedResponse dbgp.Response

		response, err, timedOut := reader.ReadResponse()

		if timedOut {
			if reader.HasCommandsToRun() {
				err = nil // set err to nil, so it resets the timeout
				goto ReadInput
			}
			continue
		}

		if err != nil { // reading failed
			break
		}

		if showXML {
			fmt.Fprintf(output, "%s\n", Faint(response))
		}

		formattedResponse = reader.FormatXML(response)
		fmt.Fprintln(output, formattedResponse)

		if formattedResponse.ExpectMoreResponses() {
			continue
		}

	ReadInput:
		line, found := reader.GetNextCommand()

		if !found { // if there was no command in the queue, read from the command line
			line, err = rl.Readline()
		}

		if err != nil { // io.EOF
			break
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
			break
		}

		lastCommand = line
	}
	c.Close()
	signal.Reset()
	fmt.Fprintf(output, "Disconnect\n")
}

func connectToProxy(address string) (net.Conn, error) {
	var conn net.Conn
	var err error
	var cert tls.Certificate

	if ssl {
		cert, err = tls.LoadX509KeyPair("certs/client.pem", "certs/client.key")
		if err != nil {
			return nil, err
		}
		config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
		conn, err = tls.Dial("tcp", address, &config)

		if err != nil {
			return nil, err
		}
	} else {
		conn, err = net.Dial("tcp", address)

		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func registerWithProxy(address string, idekey string) error {
	conn, err := connectToProxy(address)
	if err != nil {
		return err
	}

	dbgp := dbgp.NewDbgpClient(conn, false)

	command := "proxyinit -m 1 -k " + idekey + " -p " + strconv.Itoa(port)

	dbgp.SendCommand(command)

	response, err, _ := dbgp.ReadResponse()
	if err != nil {
		return fmt.Errorf("proxyinit failed: %s", err)
	}

	if showXML {
		fmt.Fprintf(output, "%s\n", Faint(response))
	}

	formatted := dbgp.FormatXML(response)

	fmt.Fprintln(output, formatted)

	if !formatted.IsSuccess() {
		return fmt.Errorf("proxyinit failed")
	}

	return nil
}

func unregisterWithProxy(address string, idekey string) error {
	conn, err := connectToProxy(address)
	if err != nil {
		return err
	}

	dbgp := dbgp.NewDbgpClient(conn, false)

	command := "proxystop -k " + idekey

	dbgp.SendCommand(command)

	response, err, _ := dbgp.ReadResponse()
	if err != nil {
		return fmt.Errorf("proxystop failed: %s", err)
	}

	if showXML {
		fmt.Fprintf(output, "%s\n", Faint(response))
	}

	formatted := dbgp.FormatXML(response)
	fmt.Fprintln(output, formatted)

	if !formatted.IsSuccess() {
		return fmt.Errorf("proxystop failed")
	}

	return nil
}

var (
	help        = false
	once        = false
	port        = 9000
	proxy       = "localhost:9001"
	register    = ""
	showXML     = false
	smartClient = false
	ssl         = false
	sslPort     = 9030
	sslProxy    = "localhost:9031"
	version     = false
	unregister  = ""
	output      = ansicon.Convert(os.Stdout)
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
	handleProxyFlags()
	getopt.Flag(&smartClient, 'f', "Whether act as fully featured DBGp client")
	getopt.Flag(&version, 'v', "Show version number and exit")
	getopt.Flag(&showXML, 'x', "Show protocol XML")
	getopt.Flag(&once, '1', "Debug once and then exit")

	getopt.Parse()

	if help {
		getopt.PrintUsage(os.Stdout)
		os.Exit(1)
	}
	if version {
		os.Exit(0)
	}

	handleProxyArguments()
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

func main() {
	handleArguments()
	printStartUp()

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
		c, err := l.Accept()
		if err != nil {
			fmt.Fprintln(output, err)
			return
		}

		handleConnection(c, rl)

		if once {
			break
		}
	}
}
