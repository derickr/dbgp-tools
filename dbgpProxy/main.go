package main

import (
	"fmt"
	"github.com/bitbored/go-ansicon" // BSD-3
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/protocol"
	"github.com/derickr/dbgp-tools/lib/proxy"
	"github.com/derickr/dbgp-tools/lib/server"
	"github.com/pborman/getopt/v2" // BSD-3
	"net"
	"os"
	"os/signal"
	"sync"
)

var clientVersion = "0.4.2-dev"

var (
	cloudUser        = ""
	disCloudUser     = ""
	enableSSLServers = false
	CloudDomain      = "cloud.xdebug.com"
	CloudPort        = "9021"
	help             = false
	clientAddress    = "localhost:9001"
	clientSSLAddress = "localhost:9011"
	serverAddress    = "localhost:9003"
	serverSSLAddress = "localhost:9013"
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

func main() {
	var cloudClient *server.Server
	var serverServer *server.Server
	var clientSSLServer *server.Server
	var serverSSLServer *server.Server

	log := logger.NewConsoleLogger(output)

	printStartUp()
	checkEnableSSLServers(log)
	handleArguments()

	ideConnectionList := connections.NewConnectionList()

	syncGroup := &sync.WaitGroup{}
	signalShutdown := make(chan int, 1)

	if cloudUser != "" {
		if disCloudUser != "" {
			protocol.UnregisterCloudClient(CloudDomain, CloudPort, disCloudUser, output, log)
		}
		cloudClient = server.NewServer(
			"cloud-client-ssl",
			resolveTCP(connections.CloudHostFromUserId(CloudDomain, CloudPort, cloudUser)),
			syncGroup,
			log)
		err := cloudClient.CloudConnect(proxy.NewServerHandler(ideConnectionList, log), cloudUser, signalShutdown)
		if err != nil {
			log.LogError("dbgpProxy", "Proxy could not be started: %s", err)
			return
		}
	} else {
		serverServer = server.NewServer("server", resolveTCP(serverAddress), syncGroup, log)
		go serverServer.Listen(proxy.NewServerHandler(ideConnectionList, log))

		if enableSSLServers {
			serverSSLServer = server.NewServer("server-ssl", resolveTCP(serverSSLAddress), syncGroup, log)
			go serverSSLServer.ListenSSL(proxy.NewServerHandler(ideConnectionList, log))
		}
	}

	clientServer := server.NewServer("client", resolveTCP(clientAddress), syncGroup, log)
	go clientServer.Listen(proxy.NewClientHandler(ideConnectionList, log))
	if enableSSLServers {
		clientSSLServer = server.NewServer("client-ssl", resolveTCP(clientSSLAddress), syncGroup, log)
		go clientSSLServer.ListenSSL(proxy.NewClientHandler(ideConnectionList, log))
	}

	log.LogInfo("dbgpProxy", "Proxy started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	select {
	case s := <-signals:
		log.LogWarning("dbgpProxy", "Signal received: %s", s)
	case <-signalShutdown:
		log.LogWarning("dbgpProxy", "Shutdown requested")
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

	log.LogInfo("dbgpProxy", "Proxy stopped")
}

func resolveTCP(host string) *net.TCPAddr {
	address, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		panic(err)
	}
	return address
}
