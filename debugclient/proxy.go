// +build !NO_PROXY

package main

import (
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
	"github.com/pborman/getopt/v2"    // BSD-3
	"os"
)

func handleProxyFlags() {
	getopt.FlagLong(&key, "proxy-key", 'k', "The IDE Key to use with the DBGp proxy")
	getopt.FlagLong(&proxy, "proxy-init", 'i', "Register with a DBGp proxy")
	getopt.FlagLong(&unproxy, "proxy-stop", 'u', "Unregister with a DBGp proxy")
}

func handleProxyArguments() {
	if proxy != "none" {
		if key == "" {
			getopt.PrintUsage(os.Stdout)
			os.Exit(1)
		}

		if unproxy {
			err := unregisterWithProxy(proxy, key)
			if err != nil {
				fmt.Fprintf(output, "%s: %s\n", BrightRed(Bold("Error unregistering with proxy")), BrightRed(err.Error()))
				os.Exit(2)
			}
			os.Exit(0)
		}

		err := registerWithProxy(proxy, key)
		if err != nil {
			fmt.Fprintf(output, "%s: %s\n", BrightRed(Bold("Error registering with proxy")), BrightRed(err.Error()))
			os.Exit(2)
		}
	}
}
