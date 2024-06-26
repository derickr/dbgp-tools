package command

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/dbgpxml"
	"github.com/derickr/dbgp-tools/lib/logger"
	"net"
)

type CloudInitCommand struct {
	connectionList *connections.ConnectionList
	connection     *net.Conn
	logger         logger.Logger
	userId         string
	needsRemoving  bool
}

func NewCloudInitCommand(connectionList *connections.ConnectionList, connection *net.Conn, logger logger.Logger) *CloudInitCommand {
	return &CloudInitCommand{connectionList: connectionList, connection: connection, logger: logger, userId: "", needsRemoving: true}
}

func (ciCommand *CloudInitCommand) GetName() string {
	return "cloudinit"
}

func (ciCommand *CloudInitCommand) GetKey() string {
	return ciCommand.userId
}

func (ciCommand *CloudInitCommand) ActUponConnection() error {
	conn := connections.NewConnection(ciCommand.userId, "", "", true, ciCommand.connection)
	err := ciCommand.connectionList.Add(conn)

	if err == nil {
		ciCommand.logger.LogUserInfo("cloudinit", ciCommand.userId, "Added connection from %s", (*ciCommand.connection).RemoteAddr())
	} else {
		ciCommand.needsRemoving = false
		ciCommand.logger.LogWarning("cloudinit", "Could not add connection: %s", err.Error())
	}

	return err
}

func (ciCommand *CloudInitCommand) Close() {
	if ciCommand.needsRemoving {
		ciCommand.logger.LogUserInfo("cloudinit", ciCommand.userId, "CloudInit::Close: Removed connection for Cloud User from %s", (*ciCommand.connection).RemoteAddr())
		ciCommand.connectionList.RemoveByKey(ciCommand.userId)
	}
}

func (ciCommand *CloudInitCommand) Handle() (string, error) {
	var init *dbgpxml.CloudInit

	conn := connections.NewConnection(ciCommand.userId, "", "", true, ciCommand.connection)
	err := ciCommand.connectionList.Add(conn)

	if err == nil {
		ciCommand.logger.LogUserInfo("cloudinit", ciCommand.userId, "Added connection for Cloud User from %s", (*ciCommand.connection).RemoteAddr())
		init = dbgpxml.NewCloudInit(true, ciCommand.userId, nil, false, nil)
	} else {
		ciCommand.needsRemoving = false
		ciCommand.logger.LogUserWarning("cloudinit", ciCommand.userId, "Could not add connection: %s", err.Error())
		init = dbgpxml.NewCloudInit(false, ciCommand.userId, &dbgpxml.CloudInitError{ID: "CLOUD-ERR-11", Message: err.Error()}, false, nil)
	}

	return init.AsXML()
}

/* cloudinit -u <userid> */
func CreateCloudInit(connectionsList *connections.ConnectionList, connection *net.Conn, arguments []string, logger logger.Logger) (DbgpCloudCommand, error) {
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
