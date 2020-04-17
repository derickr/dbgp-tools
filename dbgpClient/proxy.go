// +build !NO_PROXY

package main

import (
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
	"github.com/pborman/getopt/v2"    // BSD-3
	"os"
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
