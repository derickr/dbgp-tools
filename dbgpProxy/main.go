package main

import (
	"fmt"
	"github.com/bitbored/go-ansicon" // BSD-3
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/protocol"
	"github.com/derickr/dbgp-tools/lib/proxy"
	"github.com/derickr/dbgp-tools/lib/server"
	"github.com/derickr/dbgp-tools/lib/xml"
	"github.com/pborman/getopt/v2" // BSD-3
	"net"
	"os"
	"os/signal"
	"sync"
	// "time"
)

var clientVersion = "0.2"

var (
	cloudUser        = ""
	disCloudUser     = ""
	enableSSLServers = false
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

func checkEnableSSLServers(logger logger.Logger) {
	if _, err := os.Stat("certs/fullchain.pem"); err != nil {
		logger.LogWarning("SSL", "The 'certs/fullchain.pem' file could not be found, not enabling SSL listeners")
		return
	}
	if _, err := os.Stat("certs/privkey.pem"); err != nil {
		logger.LogWarning("SSL", "The 'certs/privkey.pem' file could not be found, not enabling SSL listeners")
		return
	}
	enableSSLServers = true
}

func handleArguments() {
	getopt.Flag(&help, 'h', "Show this help")
	getopt.FlagLong(&clientAddress, "client", 'i', "Specify the host:port to listen on for IDE (client) connections", "host:port")
	getopt.FlagLong(&serverAddress, "server", 's', "Specify the host:port to listen on for debugger engine (server) connections", "host:port")
	if enableSSLServers {
		getopt.FlagLong(&clientSSLAddress, "client-ssl", 0, "Specify the host:port to listen on for IDE (client) SSL connections", "host:port")
		getopt.FlagLong(&serverSSLAddress, "server-ssl", 0, "Specify the host:port to listen on for debugger engine (server) SSL connections", "host:port")
	}
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

func handleConnection(c net.Conn, logger logger.Logger) error {
	reader := protocol.NewDbgpClient(c, false, logger)

	response, err := reader.ReadResponse()

	if err != nil { // reading failed
		return err
	}

	if !dbgpXml.IsValidXml(response) {
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
	var err error
	var cloudClient *server.Server
	var serverServer *server.Server
	var clientSSLServer *server.Server
	var serverSSLServer *server.Server

	logger := logger.NewConsoleLogger(output)

	printStartUp()
	checkEnableSSLServers(logger)
	handleArguments()

	ideConnectionList := connections.NewConnectionList()

	syncGroup := &sync.WaitGroup{}
	signalShutdown := make(chan int, 1)

	if cloudUser != "" {
		if disCloudUser != "" {
			protocol.UnregisterCloudClient(CloudDomain, CloudPort, disCloudUser, output, logger)
		}
		cloudClient = server.NewServer(
			"cloud-client-ssl",
			resolveTCP(connections.CloudHostFromUserId(CloudDomain, CloudPort, cloudUser)),
			syncGroup,
			logger)
		err = cloudClient.CloudConnect(proxy.NewServerHandler(ideConnectionList, logger), cloudUser, signalShutdown)
	} else {
		serverServer = server.NewServer("server", resolveTCP(serverAddress), syncGroup, logger)
		go serverServer.Listen(proxy.NewServerHandler(ideConnectionList, logger))

		if enableSSLServers {
			serverSSLServer = server.NewServer("server-ssl", resolveTCP(serverSSLAddress), syncGroup, logger)
			go serverSSLServer.ListenSSL(proxy.NewServerHandler(ideConnectionList, logger))
		}
	}

	if err != nil {
		logger.LogError("dbgpProxy", "Proxy could not be started: %s", err)
		return
	}

	clientServer := server.NewServer("client", resolveTCP(clientAddress), syncGroup, logger)
	go clientServer.Listen(proxy.NewClientHandler(ideConnectionList, logger))
	if enableSSLServers {
		clientSSLServer = server.NewServer("client-ssl", resolveTCP(clientSSLAddress), syncGroup, logger)
		go clientSSLServer.ListenSSL(proxy.NewClientHandler(ideConnectionList, logger))
	}

	logger.LogInfo("dbgpProxy", "Proxy started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	select {
	case s := <-signals:
		logger.LogWarning("dbgpProxy", "Signal received: %s", s)
	case <-signalShutdown:
		logger.LogWarning("dbgpProxy", "Shutdown requested")
	}

	clientServer.Stop()
	if enableSSLServers {
		clientSSLServer.Stop()
	}

	if cloudUser == "" {
		serverServer.Stop()
		if enableSSLServers {
			serverSSLServer.Stop()
		}
	} else {
		cloudClient.Stop()
	}

	syncGroup.Wait()

	logger.LogInfo("dbgpProxy", "Proxy stopped")
}

func resolveTCP(host string) *net.TCPAddr {
	address, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		panic(err)
	}
	return address
}
