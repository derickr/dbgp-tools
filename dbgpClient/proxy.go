// +build !NO_PROXY

package main

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/protocol"
	. "github.com/logrusorgru/aurora" // WTFPL
	"github.com/pborman/getopt/v2"    // BSD-3
	"os"
	"strconv"
)

func handleProxyFlags() {
	getopt.FlagLong(&proxy, "proxy", 'y', "Register with a DBGp proxy", "host:port")
	getopt.FlagLong(&register, "register", 'r', "Register client with DBGp proxy", "idekey")
	getopt.FlagLong(&unregister, "unregister", 'u', "Unregister client with DBGp proxy", "idekey")
	getopt.FlagLong(&ssl, "ssl", 's', "Enable SSL")
}

func handleProxyArguments() {
	if register != "" {
		if cloudUser != "" {
			fmt.Fprintf(output, "%s\n", BrightRed(Bold("Refusing to register to proxy because we're connecting to Xdebug Cloud")))
			os.Exit(2)
		}

		err := registerWithProxy(proxy, register)
		if err != nil {
			fmt.Fprintf(output, "%s: %s\n", BrightRed(Bold("Error registering with proxy")), BrightRed(err.Error()))
			os.Exit(2)
		}
	}

	if unregister != "" {
		if cloudUser != "" {
			fmt.Fprintf(output, "%s\n", BrightRed(Bold("Refusing to register to proxy because we're connecting to Xdebug Cloud")))
			os.Exit(2)
		}
		err := unregisterWithProxy(proxy, unregister)
		if err != nil {
			fmt.Fprintf(output, "%s: %s\n", BrightRed(Bold("Error unregistering with proxy")), BrightRed(err.Error()))
			os.Exit(2)
		}
		os.Exit(0)
	}
}

func registerWithProxy(address string, idekey string) error {
	conn, err := connections.ConnectTo(address, ssl)
	if err != nil {
		return err
	}

	protocol := protocol.NewDbgpClient(conn, false, logger)

	command := "proxyinit -m 1 -k " + idekey + " -p " + strconv.Itoa(port)
	if ssl {
		command = command + " -s 1"
	}

	protocol.SendCommand(command)

	response, err, _ := protocol.ReadResponse()
	if err != nil {
		return fmt.Errorf("proxyinit failed: %s", err)
	}

	if showXML {
		fmt.Fprintf(output, "%s\n", Faint(response))
	}

	formatted := protocol.FormatXML(response)

	fmt.Fprintln(output, formatted)

	if !formatted.IsSuccess() {
		return fmt.Errorf("proxyinit failed")
	}

	return nil
}

func unregisterWithProxy(address string, idekey string) error {
	conn, err := connections.ConnectTo(address, ssl)
	if err != nil {
		return err
	}

	protocol := protocol.NewDbgpClient(conn, false, logger)

	command := "proxystop -k " + idekey

	protocol.SendCommand(command)

	response, err, _ := protocol.ReadResponse()
	if err != nil {
		return fmt.Errorf("proxystop failed: %s", err)
	}

	if showXML {
		fmt.Fprintf(output, "%s\n", Faint(response))
	}

	formatted := protocol.FormatXML(response)
	fmt.Fprintln(output, formatted)

	if !formatted.IsSuccess() {
		return fmt.Errorf("proxystop failed")
	}

	return nil
}
