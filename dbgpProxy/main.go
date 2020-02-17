package main

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/proxy"
	"github.com/derickr/dbgp-tools/lib/server"
	"github.com/pborman/getopt/v2" // BSD-3
	"net"
	"os"
	"os/signal"
	"sync"
)

var (
	help             = false
	clientAddress    = "localhost:9001"
	clientSSLAddress = "localhost:9031"
	serverAddress    = "localhost:9000"
	serverSSLAddress = "localhost:9030"
	version          = false
)

func printStartUp() {
	fmt.Println("Xdebug DBGp proxy (Go 0.1)")
	fmt.Println("Copyright 2020 by Derick Rethans")
}

func handleArguments() {
	getopt.Flag(&help, 'h', "Show this help")
	getopt.FlagLong(&clientAddress, "client", 'c', "Specify the host:port to listen on for IDE (client) connections", "host:port")
	getopt.FlagLong(&clientSSLAddress, "client-ssl", 0, "Specify the host:port to listen on for IDE (client) SSL connections", "host:port")
	getopt.FlagLong(&serverAddress, "server", 's', "Specify the host:port to listen on for debugger engine (server) connections", "host:port")
	getopt.FlagLong(&serverSSLAddress, "server-ssl", 0, "Specify the host:port to listen on for debugger engine (server) SSL connections", "host:port")
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
	clientSSLServer := server.NewServer("client-ssl", resolveTCP(clientSSLAddress), syncGroup)
	serverSSLServer := server.NewServer("server-ssl", resolveTCP(serverSSLAddress), syncGroup)

	go clientServer.Listen(proxy.NewClientHandler(ideConnectionList))
	go serverServer.Listen(proxy.NewServerHandler(ideConnectionList))
	go clientSSLServer.ListenSSL(proxy.NewClientHandler(ideConnectionList))
	go serverSSLServer.ListenSSL(proxy.NewServerHandler(ideConnectionList))

	fmt.Println("Proxy started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	fmt.Printf("- Signal received: %s\n", <-signals)

	clientServer.Stop()
	serverServer.Stop()
	clientSSLServer.Stop()
	serverSSLServer.Stop()
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
