package connections

import (
	"fmt"
	"net"
	"sync"
)

type Connection struct {
	ideKey    string
	ipAddress string
	port      string
}

func NewConnection(ideKey string, ipAddress string, port string) *Connection {
	return &Connection{ideKey: ideKey, ipAddress: ipAddress, port: port}
}

func (connection *Connection) FullAddress() string {
	return net.JoinHostPort(connection.ipAddress, connection.port)
}

type ConnectionList struct {
	sync.Mutex
	connections map[string]*Connection
}

func NewConnectionList() *ConnectionList {
	return &ConnectionList{connections: map[string]*Connection{}}
}

func (list *ConnectionList) Add(connection *Connection) error {
	list.Lock()
	defer list.Unlock()

	existing, ok := list.connections[connection.ideKey]

	if ok {
		return fmt.Errorf("IDE Key '%s' is already registered for connection %s:%s", existing.ideKey, existing.ipAddress, existing.port)
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
