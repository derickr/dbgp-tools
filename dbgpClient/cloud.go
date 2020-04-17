// +build !NO_CLOUD

package main

import (
	"github.com/pborman/getopt/v2" // BSD-3
)

func handleCloudFlags() {
	getopt.FlagLong(&cloudUser, "cloud", 'c', "Connect to Xdebug Cloud", "email-address")
}

func handleCloudArguments() {
	if cloudUser != "" {
		ssl = true
	}
}
