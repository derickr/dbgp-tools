package main

import (
	"fmt"
	"github.com/bitbored/go-ansicon" // BSD-3
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/proxy"
	"github.com/derickr/dbgp-tools/lib/server"
	"github.com/pborman/getopt/v2" // BSD-3
	"net"
	"os"
	"os/signal"
	"sync"
)

var clientVersion = "0.2"

var (
	help             = false
	clientAddress    = "localhost:9001"
	clientSSLAddress = "localhost:9011"
	serverAddress    = "localhost:9000"
	serverSSLAddress = "localhost:9010"
	output           = ansicon.Convert(os.Stdout)
	version          = false
)

func printStartUp() {
	fmt.Printf("Xdebug DBGp proxy (%s)\n", clientVersion)
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
		getopt.PrintUsage(output)
		os.Exit(1)
	}
	if version {
		os.Exit(0)
	}
}

func formatError(connection connections.Connection) error {
	return fmt.Errorf("IDE Key '%s' is already registered for connection %s", connection.GetKey(), connection.FullAddress())
}

func main() {
	printStartUp()
	handleArguments()

	logger := server.NewConsoleLogger(output)

	ideConnectionList := connections.NewConnectionList(formatError)

	syncGroup := &sync.WaitGroup{}
	clientServer := server.NewServer("client", resolveTCP(clientAddress), syncGroup, logger)
	serverServer := server.NewServer("server", resolveTCP(serverAddress), syncGroup, logger)
	clientSSLServer := server.NewServer("client-ssl", resolveTCP(clientSSLAddress), syncGroup, logger)
	serverSSLServer := server.NewServer("server-ssl", resolveTCP(serverSSLAddress), syncGroup, logger)

	go clientServer.Listen(proxy.NewClientHandler(ideConnectionList, logger))
	go serverServer.Listen(proxy.NewServerHandler(ideConnectionList, logger))
	go clientSSLServer.ListenSSL(proxy.NewClientHandler(ideConnectionList, logger))
	go serverSSLServer.ListenSSL(proxy.NewServerHandler(ideConnectionList, logger))

	logger.LogInfo("server", "Proxy started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	logger.LogWarning("server", "Signal received: %s", <-signals)

	clientServer.Stop()
	serverServer.Stop()
	clientSSLServer.Stop()
	serverSSLServer.Stop()
	syncGroup.Wait()

	logger.LogInfo("server", "Proxy stopped")
}

func resolveTCP(host string) *net.TCPAddr {
	address, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		panic(err)
	}
	return address
}
