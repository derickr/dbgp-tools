// +build !NO_CLOUD

package main

import (
	"github.com/pborman/getopt/v2" // BSD-3
)

func handleCloudFlags() {
	getopt.FlagLong(&cloudUser, "cloud", 'c', "Connect to Xdebug Cloud", "cloud-user-id")
	getopt.FlagLong(&disCloudUser, "discloud", 'd', "Disconnect from Xdebug Cloud", "cloud-user-id")
}

func handleCloudArguments() {
}
