package main

import (
	"fmt"
	"github.com/pborman/getopt/v2" // BSD-3
	"github.com/xdebug/dbgp-tools/lib/connections"
	"github.com/xdebug/dbgp-tools/lib/proxy"
	"github.com/xdebug/dbgp-tools/lib/server"
	"net"
	"os"
	//	"os/user"
	"os/signal"
	"sync"
)

var (
	help          = false
	clientAddress = "localhost:9001"
	serverAddress = "localhost:9000"
	version       = false
)

func printStartUp() {
	fmt.Println("Xdebug DBGp proxy (Go 0.1)")
	fmt.Println("Copyright 2020 by Derick Rethans")
}

func handleArguments() {
	getopt.Flag(&help, 'h', "Show this help")
	getopt.FlagLong(&clientAddress, "client", 'c', "Specify the host:port to listen on for IDE (client) connections", "host:port")
	getopt.FlagLong(&serverAddress, "server", 's', "Specify the host:port to listen on for debugger engine (server) connections", "host:port")
	getopt.Flag(&version, 'v', "Show version number and exit")

	getopt.Parse()

	if help {
		getopt.PrintUsage(os.Stdout)
		os.Exit(1)
	}
	if version {
		os.Exit(0)
	}
}

func main() {
	printStartUp()
	handleArguments()

	ideConnectionList := connections.NewConnectionList()

	syncGroup := &sync.WaitGroup{}
	clientServer := server.NewServer("client", resolveTCP(clientAddress), syncGroup)
	serverServer := server.NewServer("server", resolveTCP(serverAddress), syncGroup)

	go clientServer.Listen(proxy.NewClientHandler(ideConnectionList))
	go serverServer.Listen(proxy.NewServerHandler(ideConnectionList))

	fmt.Println("Proxy started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	fmt.Printf("Signal received: %s", <-signals)

	clientServer.Stop()
	serverServer.Stop()
	syncGroup.Wait()

	fmt.Println("Proxy stopped")
}

func resolveTCP(host string) *net.TCPAddr {
	address, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		panic(err)
	}
	return address
}
