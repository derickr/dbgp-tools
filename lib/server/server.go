package server

import (
	"crypto/tls"
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/protocol"
	"io"
	"net"
	"sync"
	"time"
)

type Handler interface {
	Handle(conn net.Conn) error
}

type Server struct {
	logger     logger.Logger
	address    *net.TCPAddr
	group      *sync.WaitGroup
	stop       bool
	serverType string
}

func NewServer(serverType string, address *net.TCPAddr, group *sync.WaitGroup, logger logger.Logger) *Server {
	return &Server{
		logger,
		address,
		group,
		false,
		serverType,
	}
}

func (server *Server) Listen(handler Handler) {
	server.group.Add(1)
	defer server.group.Done()

	listener, err := net.ListenTCP("tcp", server.address)
	if err != nil {
		panic(err)
	}

	defer server.closeConnection(listener)

	server.logger.LogInfo("server", "Started %s server on %s", server.serverType, server.address)

	for {
		if server.stop {
			break
		}

		_ = listener.SetDeadline(time.Now().Add(time.Second * 2))
		conn, err := listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			fmt.Print(err)
			continue
		}
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second)
		go server.handleConnection(conn, handler)
	}

	server.logger.LogInfo("server", "Shutdown %s server", server.serverType)
}

func (server *Server) ListenSSL(handler Handler) {
	server.group.Add(1)
	defer server.group.Done()

	cert, err := tls.LoadX509KeyPair("certs/fullchain.pem", "certs/privkey.pem")
	if err != nil {
		server.logger.LogError("server", "Can not load SSL keys: %s", err)
		panic(err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}

	listener, err := net.ListenTCP("tcp", server.address)
	if err != nil {
		panic(err)
	}

	defer server.closeConnection(listener)

	server.logger.LogInfo("server", "Started %s SSL server on %s", server.serverType, server.address)

	for {
		if server.stop {
			break
		}

		_ = listener.SetDeadline(time.Now().Add(time.Second * 2))
		conn, err := listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			fmt.Print(err)
			continue
		}
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Millisecond * 3)
		go server.handleConnection(tls.Server(conn, &config), handler)
	}

	server.logger.LogInfo("server", "Shutdown %s SSL server", server.serverType)
}

func (server *Server) CloudConnect(handler Handler, cloudUser string) {
	connToCloud, err := connections.ConnectTo(server.address.String(), true)

	if err != nil {
		server.logger.LogUserError("server", cloudUser, "Can not connect to Xdebug Cloud: %s", err)
		return
	}

	server.logger.LogUserInfo("server", cloudUser, "Connected to Xdebug Cloud on %s", server.address)

	err = protocol.NewDbgpClient(connToCloud, false, server.logger).RunCommand("cloudinit -u " + cloudUser)
	if err != nil {
		server.logger.LogUserError("server", cloudUser, "Not connected to Xdebug Cloud: %s", err)
		return
	}

	go server.handleConnection(connToCloud, handler)
}

func (server *Server) handleConnection(conn net.Conn, handler Handler) {
	defer server.closeConnection(conn)
	server.group.Add(1)
	defer server.group.Done()

	server.logger.LogInfo("server", "Start new %s connection from %s", server.serverType, conn.RemoteAddr())

	err := handler.Handle(conn)

	if err != nil {
		server.logger.LogWarning("server", "Handler response error: %s", err)
	}

	server.logger.LogInfo("server", "Closing %s connection from %s", server.serverType, conn.RemoteAddr())
}

func (server *Server) closeConnection(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		server.logger.LogWarning("server", "Couldn't close connection: %s", err)
	}
}

func (server *Server) Stop() {
	server.stop = true
}
