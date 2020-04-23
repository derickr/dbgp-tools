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
	claimed         bool
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
		claimed:         false,
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

func (list *ConnectionList) ClaimConnection(ideKey string) (*Connection, error) {
	list.Lock()
	defer list.Unlock()

	connection, ok := list.connections[ideKey]

	if !ok {
		return nil, fmt.Errorf("Can not find the connection with key '%s' to claim", ideKey)
	}
	if connection.claimed {
		return nil, fmt.Errorf("The connection with key '%s' is already claimed", ideKey)
	}
	connection.claimed = true

	return connection, nil
}

func (list *ConnectionList) UnclaimConnection(ideKey string) error {
	list.Lock()
	defer list.Unlock()

	connection, ok := list.connections[ideKey]

	if !ok {
		return fmt.Errorf("Can not find the connection with key '%s' to unclaim", ideKey)
	}
	if !connection.claimed {
		return fmt.Errorf("The connection with key '%s' was not claimed", ideKey)
	}
	connection.claimed = false

	return nil
}
