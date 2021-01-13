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
	stopGuard  *sync.RWMutex
	stop       bool
	serverType string
}

func NewServer(serverType string, address *net.TCPAddr, group *sync.WaitGroup, logger logger.Logger) *Server {
	return &Server{
		logger,
		address,
		group,
		&sync.RWMutex{},
		false,
		serverType,
	}
}

func (server *Server) Listen(handler Handler) {
	// TODO(florin): Is the idea behind this to count how many active connections there are
	// 	and only when the last one is closed, stop the server? Or is this correct and it's
	// 	only available here to count how many listeners are alive?
	server.group.Add(1)
	defer server.group.Done()

	listener, err := net.ListenTCP("tcp", server.address)
	if err != nil {
		panic(err)
	}

	defer server.closeConnection(listener)

	server.logger.LogInfo("server", "Started %s server on %s", server.serverType, server.address)

	for {
		// TODO(florin): Here there's a race condition between this for loop
		// 	and the *Server.Stop() method.
		// 	I surrounded it with a rwmutex to cover for that.
		server.stopGuard.RLock()
		if server.stop {
			server.stopGuard.RUnlock()
			break
		}
		server.stopGuard.RUnlock()

		_ = listener.SetDeadline(time.Now().Add(time.Second * 2))
		conn, err := listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			// TODO(florin): I think this should call a server.logger.Log* method
			fmt.Print(err)
			continue
		}
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second)
		go server.handleConnection(conn, handler, nil)
	}

	server.logger.LogInfo("server", "Shutdown %s server", server.serverType)
}

// TODO(florin): Given this is pretty much similar to the above method,
// 	I would introduce an unexported method to handle both of these based on the config being nil or not
func (server *Server) ListenSSL(handler Handler) {
	// TODO(florin): Same question as with Listen() method above
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
		server.stopGuard.RLock()
		if server.stop {
			server.stopGuard.RUnlock()
			break
		}
		server.stopGuard.RUnlock()

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
		go server.handleConnection(tls.Server(conn, &config), handler, nil)
	}

	server.logger.LogInfo("server", "Shutdown %s SSL server", server.serverType)
}

func (server *Server) CloudConnect(handler Handler, cloudUser string, shutdownSignal chan int) error {
	connToCloud, err := connections.ConnectTo(server.address.String(), true)

	if err != nil {
		server.logger.LogUserError("server", cloudUser, "Can not connect to Xdebug Cloud: %s", err)
		return err
	}

	server.logger.LogUserInfo("server", cloudUser, "Connected to Xdebug Cloud on %s", server.address)

	err = protocol.NewDbgpClient(connToCloud, false, server.logger).RunCommand("cloudinit -u " + cloudUser)
	if err != nil {
		server.logger.LogUserError("server", cloudUser, "Not connected to Xdebug Cloud: %s", err)
		return err
	}

	// TODO(florin): Given this is the only place that uses the "shutdownSignal" property,
	// 	I would move it to the *Server struct and then have it passed via NewServer() so that
	// 	Listen and ListenSSL can also receive the channel.
	go server.handleConnection(connToCloud, handler, shutdownSignal)

	return nil
}

func (server *Server) handleConnection(conn net.Conn, handler Handler, shutdownSignal chan int) {
	// TODO(florin): Just to double-check here, the server.closeConnection() will be called
	// 	only after group.Done() will be, since the defer execution order is in reverse
	// 	of their declaration. Is this intended?
	defer server.closeConnection(conn)
	server.group.Add(1)
	defer server.group.Done()

	server.logger.LogInfo("server", "Start new %s connection from %s", server.serverType, conn.RemoteAddr())

	err := handler.Handle(conn)

	// TODO(florin): This would shutdown the whole server, not just this connection, if a single error
	// 	is received while processing data in a single connection.
	// 	Would it be better to return an error and let the client retry/handle termination?
	if err != nil {
		server.logger.LogWarning("server", "Handler response error: %s", err)
		if shutdownSignal != nil {
			shutdownSignal <- 1
		}
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
	server.stopGuard.Lock()
	server.stop = true
	server.stopGuard.Unlock()
}
