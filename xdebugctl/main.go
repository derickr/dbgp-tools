package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/bitbored/go-ansicon" // BSD-3
	"github.com/derickr/dbgp-tools/lib/dbgpxml"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/protocol"
	. "github.com/logrusorgru/aurora" // WTFPL
	"github.com/pborman/getopt/v2"    // BSD-3
)

var clientVersion = "0.3.0"
var clientYear    = "2025"

var (
	command   = ""
	help      = false
	pid       = 0
	showXML   = false
	version   = false
	output    = ansicon.Convert(os.Stdout)
	logOutput = logger.NewConsoleLogger(output)
)

func printVersion() {
	fmt.Fprintf(output, "Xdebug Controller (%s)\n", Bold(clientVersion))
	fmt.Fprintf(output, "Copyright 2023-%s by Derick Rethans\n", clientYear)
}

func printCommandList() {
	fmt.Fprintf(output, "\n")
	fmt.Fprintf(output, "Commands:\n\n")
	fmt.Fprintf(output, " ps        Lists all Xdebug enabled PHP scripts\n")
	fmt.Fprintf(output, " pause     Instructs Xdebug to initiate a debugging connection or breakpoint\n")
	fmt.Fprintf(output, "\n")
}

func printStartUp() {
	fmt.Fprintf(output, "\n")
}

func handleArguments() {
	getopt.Flag(&help, 'h', "Show this help")
	getopt.Flag(&pid, 'p', "Specify the PID to operate on")
	getopt.Flag(&version, 'v', "Show version number and exit")
	getopt.Flag(&showXML, 'x', "Show protocol XML")

	getopt.SetParameters("[command]")
	getopt.Parse()

	if version {
		printVersion()
		exit(0)
	}

	if help || getopt.NArgs() != 1 {
		printVersion()
		printStartUp()
		getopt.PrintUsage(os.Stdout)
		printCommandList()
		exit(1)
	}

	command = getopt.Arg(0)
}

func sendCmd(ctrl_socket string, scriptPid int, command string) string {
	xml := ""
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	unknownResponseStr := fmt.Sprintf("%10d %s: %s: %s\n", Faint(scriptPid), BrightRed("Error"), "No response on", Faint(ctrl_socket))

	conn, err := dialCtrlSocket(ctx, ctrl_socket)
	if err != nil {
		return unknownResponseStr
	}
	defer conn.Close()

	bCommand := []byte(command)
	if _, err := conn.Write(bCommand); err != nil {
		log.Fatal(err)
	}

	conn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
	response, err := bufio.NewReader(conn).ReadString('\000')
	if err != nil {
		return unknownResponseStr
	}
	if showXML {
		xml = fmt.Sprintf("%s\n", Faint(response))
	}

	if !dbgpxml.IsValidXml(response) {
		fmt.Errorf("The received XML is not valid, closing connection: %s", response)
		return ""
	}

	reader := protocol.NewDbgpClient(conn, logOutput)
	formattedResponse := reader.FormatXML(response)

	if formattedResponse == nil {
		return unknownResponseStr
	}
	return fmt.Sprintf("%s%s\n", xml, formattedResponse)
}

// Give ansicon a chance to write all output before exiting
func exit(code int) {
	if runtime.GOOS == "windows" {
		time.Sleep(time.Duration(time.Microsecond * 50))
	}
	os.Exit(code)
}

func main() {
	handleArguments()

	files, err := findFiles()
	if err != nil {
		fmt.Fprintf(output, "%s: %s: %s\n", BrightRed("Error"), "Failed reading list of control sockets", err)
		exit(1)
	}

	if len(files) > 1 && pid == 0 && command != "ps" {
		fmt.Fprintf(output, "%s: %s\n", BrightRed("Error"), "You must specify a PID with -p as there is more than one script")
		printStartUp()
		getopt.PrintUsage(os.Stdout)
		printCommandList()
		exit(1)
	}

	if command == "ps" {
		c := make(chan string)
		spawned := 0

		fmt.Fprintf(output, "%10s %8s %8s %s\n", Faint("PID"), "RSS", "TIME", BrightWhite("COMMAND"))

		for scriptPid, file := range files {
			if pid == 0 || scriptPid == pid {
				spawned++

				go func(fpid string, spid int) {
					c <- sendCmd(fpid, spid, "ps")
				}(file, scriptPid)
			}
		}

		for i := 0; i < spawned; i++ {
			result := <-c
			fmt.Fprintf(output, "%s", result)
		}
		exit(0)
	}

	if len(files) == 0 && pid == 0 && command != "ps" {
		fmt.Fprintf(output, "%s: %s\n", BrightRed("Error"), "Could not find any running PHP scripts")
		exit(2)
	}

	for scriptPid, file := range files {
		if (scriptPid == pid) || (len(files) == 1 && pid == 0) {
			if command == "pause" {
				result := sendCmd(file, scriptPid, "pause")
				fmt.Fprintf(output, "%s", result)
				return
			}
			fmt.Fprintf(output, "%s: Unknown command '%s'\n", BrightRed("Error"), command)
			exit(3)
		}
	}

	fmt.Fprintf(output, "%s: No script with PID '%d' active\n", BrightRed("Error"), pid)
	exit(4)
}
