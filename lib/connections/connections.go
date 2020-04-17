package connections

import (
	"fmt"
	"net"
	"sync"
)

type ConnectionControl struct {
	CloseConnection bool
}

func NewCloseConnectionControl() *ConnectionControl {
	return &ConnectionControl{CloseConnection: true}
}

type Connection struct {
	ideKey          string
	ipAddress       string
	port            string
	ssl             bool
	connection      *net.Conn
	DebugRequests   chan int
	ControlRequests chan *ConnectionControl
}

func NewConnection(ideKey string, ipAddress string, port string, ssl bool, connection *net.Conn) *Connection {
	return &Connection{
		ideKey:          ideKey,
		ipAddress:       ipAddress,
		port:            port,
		ssl:             ssl,
		connection:      connection,
		DebugRequests:   make(chan int),
		ControlRequests: make(chan *ConnectionControl),
	}
}

func (connection *Connection) IsSSL() bool {
	return connection.ssl == true
}

func (connection *Connection) FullAddress() string {
	return net.JoinHostPort(connection.ipAddress, connection.port)
}

func (connection *Connection) GetKey() string {
	return connection.ideKey
}

func (connection *Connection) GetConnection() net.Conn {
	return *connection.connection
}

type ConnectionList struct {
	sync.Mutex
	connections map[string]*Connection
	formatError func(existing Connection) error
}

func NewConnectionList(formatError func(existing Connection) error) *ConnectionList {
	return &ConnectionList{connections: map[string]*Connection{}, formatError: formatError}
}

func (list *ConnectionList) Add(connection *Connection) error {
	list.Lock()
	defer list.Unlock()

	existing, ok := list.connections[connection.ideKey]

	if ok {
		return list.formatError(*existing)
	}

	list.connections[connection.ideKey] = connection

	return nil
}

func (list *ConnectionList) RemoveByKey(ideKey string) error {
	list.Lock()
	defer list.Unlock()

	_, ok := list.connections[ideKey]

	if !ok {
		return fmt.Errorf("IDE Key '%s' has not been previously registered", ideKey)
	}

	delete(list.connections, ideKey)

	return nil
}

func (list *ConnectionList) FindByKey(ideKey string) (*Connection, bool) {
	list.Lock()
	defer list.Unlock()

	connection, ok := list.connections[ideKey]

	return connection, ok
}
