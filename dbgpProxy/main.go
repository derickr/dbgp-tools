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
	cloudUser        = ""
	CloudDomain      = "cloud.xdebug.com"
	CloudPort        = "9021"
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
	getopt.FlagLong(&clientAddress, "client", 'i', "Specify the host:port to listen on for IDE (client) connections", "host:port")
	getopt.FlagLong(&clientSSLAddress, "client-ssl", 0, "Specify the host:port to listen on for IDE (client) SSL connections", "host:port")
	getopt.FlagLong(&serverAddress, "server", 's', "Specify the host:port to listen on for debugger engine (server) connections", "host:port")
	getopt.FlagLong(&serverSSLAddress, "server-ssl", 0, "Specify the host:port to listen on for debugger engine (server) SSL connections", "host:port")
	getopt.Flag(&version, 'v', "Show version number and exit")

	handleCloudFlags()

	getopt.Parse()

	if help {
		getopt.PrintUsage(output)
		os.Exit(1)
	}
	if version {
		os.Exit(0)
	}
}

func isValidXml(xml string) bool {
	return strings.HasPrefix(xml, "<?xml")
}

func handleConnection(c net.Conn, logger server.Logger) error {
	reader := protocol.NewDbgpClient(c, false, logger)

	response, err, timedOut := reader.ReadResponse()

	if timedOut {
		return nil
	}

	if err != nil { // reading failed
		return err
	}

	if !isValidXml(response) {
		return fmt.Errorf("The received XML is not valid, closing connection: %s", response)
	}

	formattedResponse := reader.FormatXML(response)
	if formattedResponse.IsSuccess() == false {
		return fmt.Errorf("%s", formattedResponse.GetErrorMessage())
	}

	if formattedResponse == nil {
		return fmt.Errorf("Could not interpret XML, closing connection.")
	}

	return nil
}

func main() {
	var serverServer *server.Server
	var serverSSLServer *server.Server

	printStartUp()
	handleArguments()

	logger := server.NewConsoleLogger(output)

	ideConnectionList := connections.NewConnectionList()

	syncGroup := &sync.WaitGroup{}

	clientServer := server.NewServer("client", resolveTCP(clientAddress), syncGroup, logger)
	clientSSLServer := server.NewServer("client-ssl", resolveTCP(clientSSLAddress), syncGroup, logger)
	go clientServer.Listen(proxy.NewClientHandler(ideConnectionList, logger))
	go clientSSLServer.ListenSSL(proxy.NewClientHandler(ideConnectionList, logger))

	if cloudUser == "" {
		serverServer = server.NewServer("server", resolveTCP(serverAddress), syncGroup, logger)
		serverSSLServer = server.NewServer("server-ssl", resolveTCP(serverSSLAddress), syncGroup, logger)
		go serverServer.Listen(proxy.NewServerHandler(ideConnectionList, logger))
		go serverSSLServer.ListenSSL(proxy.NewServerHandler(ideConnectionList, logger))
	} else {
		connections.ConnectToCloud(CloudDomain, CloudPort, cloudUser, logger)
	}

	logger.LogInfo("server", "Proxy started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	logger.LogWarning("server", "Signal received: %s", <-signals)

	clientServer.Stop()
	clientSSLServer.Stop()

	if cloudUser == "" {
		serverServer.Stop()
		serverSSLServer.Stop()
	}

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
