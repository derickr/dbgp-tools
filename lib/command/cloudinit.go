package command

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/server"
	"github.com/derickr/dbgp-tools/lib/xml"
	"net"
)

type CloudInitCommand struct {
	connectionList *connections.ConnectionList
	connection     *net.Conn
	logger         server.Logger
	userId         string
	needsRemoving  bool
}

func NewCloudInitCommand(connectionList *connections.ConnectionList, connection *net.Conn, logger server.Logger) *CloudInitCommand {
	return &CloudInitCommand{connectionList: connectionList, connection: connection, logger: logger, userId: "", needsRemoving: true}
}

func (ciCommand *CloudInitCommand) GetName() string {
	return "cloudinit"
}

func (ciCommand *CloudInitCommand) GetKey() string {
	return ciCommand.userId
}

func (ciCommand *CloudInitCommand) AddConnection() error {
	conn := connections.NewConnection(ciCommand.userId, "", "", true, ciCommand.connection)
	err := ciCommand.connectionList.Add(conn)

	if err == nil {
		ciCommand.logger.LogUserInfo("conn", ciCommand.userId, "Added connection for Cloud User '%s' from %s", ciCommand.userId, (*ciCommand.connection).RemoteAddr())
	} else {
		ciCommand.needsRemoving = false
		ciCommand.logger.LogWarning("conn", "Could not add connection: %s", err.Error())
	}

	return err
}

func (ciCommand *CloudInitCommand) Close() {
	if ciCommand.needsRemoving {
		ciCommand.logger.LogUserInfo("conn", ciCommand.userId, "Removed connection for Cloud User '%s' from %s", ciCommand.userId, (*ciCommand.connection).RemoteAddr())
		ciCommand.connectionList.RemoveByKey(ciCommand.userId)
	}
}

func (ciCommand *CloudInitCommand) Handle() (string, error) {
	var init *dbgpXml.CloudInit

	conn := connections.NewConnection(ciCommand.userId, "", "", true, ciCommand.connection)
	err := ciCommand.connectionList.Add(conn)

	if err == nil {
		ciCommand.logger.LogUserInfo("conn", ciCommand.userId, "Added connection for Cloud User '%s' from %s", ciCommand.userId, (*ciCommand.connection).RemoteAddr())
		init = dbgpXml.NewCloudInit(true, ciCommand.userId, nil, nil)
	} else {
		ciCommand.needsRemoving = false
		ciCommand.logger.LogUserWarning("conn", ciCommand.userId, "Could not add connection: %s", err.Error())
		init = dbgpXml.NewCloudInit(false, ciCommand.userId, &dbgpXml.CloudInitError{ID: "ERR-01", Message: err.Error()}, nil)
	}

	return init.AsXML()
}

/* cloudinit -u <userid> */
func CreateCloudInit(connectionsList *connections.ConnectionList, connection *net.Conn, arguments []string, logger server.Logger) (DbgpCloudInitCommand, error) {
	ciCommand := NewCloudInitCommand(connectionsList, connection, logger)

	expectValue := false
	expectValueFor := ""

	for _, value := range arguments {
		if expectValue {
			expectValue = false
			switch expectValueFor {
			case "-i":
				/* ignore */
			case "-u":
				ciCommand.userId = value
			default:
				return nil, fmt.Errorf("Unknown argument '%s' (with value '%s')", expectValueFor, value)
			}
		} else {
			expectValueFor = value
			expectValue = true
		}
	}

	if expectValue {
		return nil, fmt.Errorf("No argument given for '%s'", expectValueFor)
	}
	if ciCommand.userId == "" {
		return nil, fmt.Errorf("No username was provided")
	}

	return ciCommand, nil
}
