package connections

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"sync"
)

type ConnectionControl struct {
	CloseConnection     bool
	TryReadForCloudStop bool
}

func NewCloseConnectionControl() *ConnectionControl {
	return &ConnectionControl{CloseConnection: true}
}

func NewTryReadForCloudStopControl() *ConnectionControl {
	return &ConnectionControl{TryReadForCloudStop: true}
}

type Connection struct {
	ideKey          string
	ipAddress       string
	port            string
	ssl             bool
	sid             string
	connection      *net.Conn
	claimed         bool
	DebugRequests   chan int
	ControlRequests chan *ConnectionControl
}

func NewConnection(ideKey string, ipAddress string, port string, ssl bool, connection *net.Conn) *Connection {
	sid, _ := uuid.NewRandom()

	return &Connection{
		ideKey:          ideKey,
		ipAddress:       ipAddress,
		port:            port,
		sid:             sid.String(),
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

func (connection *Connection) GetSID() string {
	return connection.sid
}

func (connection *Connection) GetConnection() net.Conn {
	return *connection.connection
}

type ConnectionList struct {
	sync.Mutex
	forceAdd bool
	connections map[string]*Connection
}

func NewConnectionList(forceAdd bool) *ConnectionList {
	return &ConnectionList{connections: map[string]*Connection{}, forceAdd: forceAdd}
}

func (list *ConnectionList) Add(connection *Connection) error {
	list.Lock()
	defer list.Unlock()

	_, ok := list.connections[connection.ideKey]

	if ok {
		if (list.forceAdd) {
			delete(list.connections, connection.ideKey)

			return nil
		} else {

			return fmt.Errorf("A client for '%s' is already connected", connection.ideKey)
		}
	}

	list.connections[connection.ideKey] = connection

	return nil
}

func (list *ConnectionList) RemoveByKey(ideKey string) error {
	list.Lock()

	connection, ok := list.connections[ideKey]

	if !ok {
		list.Unlock()
		return fmt.Errorf("A client for '%s' has not been previously registered", ideKey)
	}

	delete(list.connections, ideKey)
	list.Unlock()

	if connection.claimed {
		connection.ControlRequests <- NewCloseConnectionControl()
	}

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
		return nil, fmt.Errorf("There is no IDE connected to Xdebug Cloud for UserID '%s'", ideKey)
	}
	if connection.claimed {
		return nil, fmt.Errorf("A Xdebug connection for UserID '%s' is already active", ideKey)
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
