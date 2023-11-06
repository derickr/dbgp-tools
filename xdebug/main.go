package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/bitbored/go-ansicon" // BSD-3
	"github.com/derickr/dbgp-tools/lib/dbgpxml"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/protocol"
	"github.com/pborman/getopt/v2"    // BSD-3
	. "github.com/logrusorgru/aurora" // WTFPL
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"time"
)

var clientVersion = "0.0.1"

var re = regexp.MustCompile(`.*\s(@xdebug-ctrl\.(\d+)yx+).*`)

var (
	command = ""
	help    = false
	pid     = 0
	showXML = false
	version      = false
	output       = ansicon.Convert(os.Stdout)
	logOutput    = logger.NewConsoleLogger(output)
)

func printVersion() {
	fmt.Fprintf(output, "Xdebug Controller (%s)\n", Bold(clientVersion))
	fmt.Fprintf(output, "Copyright 2023 by Derick Rethans\n")
}

func printCommandList() {
	fmt.Fprintf(output, "\n")
	fmt.Fprintf(output, "Commands:\n\n")
	fmt.Fprintf(output, " ps        Lists all Xdebug enabled PHP scripts\n")
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
		os.Exit(0)
	}

	if help || getopt.NArgs() != 1 {
		printVersion()
		printStartUp()
		getopt.PrintUsage(os.Stdout)
		printCommandList()
		os.Exit(1)
	}

	command = getopt.Arg(0)
}

func findFiles() map[int]string {
	file, _ := os.Open("/proc/net/unix")

	s := bufio.NewScanner(file)
	v := make(map[int]string)

	for s.Scan() {
		matches := re.FindStringSubmatch(s.Text())
		if len(matches) > 0 {
			pid, _ := strconv.Atoi(matches[2])
			v[pid] = matches[1]
		}
	}

	return v
}

func sendCmd(ctrl_socket string, command string) {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	d.LocalAddr = nil // if you have a local addr, add it here
	raddr := net.UnixAddr{Name: ctrl_socket, Net: "unix"}
	conn, err := d.DialContext(ctx, "unix", raddr.String())
	if err != nil {
		log.Printf("Failed to dial: %v\n", err)
		return
	}
	defer conn.Close()

	bCommand := []byte(command);
	if _, err := conn.Write(bCommand); err != nil {
		log.Fatal(err)
	}

	response, _ := bufio.NewReader(conn).ReadString('\000')
	if showXML {
		fmt.Fprintf(output, "%s\n", Faint(response))
	}
		
	if !dbgpxml.IsValidXml(response) {
		fmt.Errorf("The received XML is not valid, closing connection: %s", response)
	}

	reader := protocol.NewDbgpClient(conn, logOutput)
	formattedResponse := reader.FormatXML(response)
		
	if formattedResponse == nil {
		fmt.Fprintf(output, "%s: Could not interpret XML response\n", BrightRed("Error"))
		return
	} 
	fmt.Fprintln(output, formattedResponse)
}

func main() {
	handleArguments()

	files := findFiles()

	if len(files) > 1 && pid == 0 && command != "ps" {
		fmt.Fprintf(output, "%s\n", BrightRed("You must specify a PID with -p as there is more than one script"))
		return
	}

	if command == "ps" {
		fmt.Fprintf(output, "%10s %8s %8s %s\n", Faint("PID"), "RSS", "TIME", BrightWhite("COMMAND"));
		for scriptPid, file := range files {
			if pid == 0 || scriptPid == pid {
				sendCmd(file, "ps")
			}
		}
		return;
	}

	for scriptPid, file := range files {
		if scriptPid == pid {
			if command == "pause" {
				sendCmd(file, "pause")
				return
			}
			fmt.Fprintf(output, "%s: Unknown command '%s'\n", BrightRed("Error"), command)
			return
		}
	}

	fmt.Fprintf(output, "%s: No script with PID '%d' active\n", BrightRed("Error"), pid)
}
