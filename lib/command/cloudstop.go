package command

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/xml"
	"net"
)

type CloudStopCommand struct {
	connectionList *connections.ConnectionList
	connection     *net.Conn
	logger         logger.Logger
	userId         string
	needsRemoving  bool
}

func NewCloudStopCommand(connectionList *connections.ConnectionList, connection *net.Conn, logger logger.Logger) *CloudStopCommand {
	return &CloudStopCommand{connectionList: connectionList, connection: connection, logger: logger, userId: "", needsRemoving: true}
}

func (csCommand *CloudStopCommand) GetName() string {
	return "cloudstop"
}

func (csCommand *CloudStopCommand) GetKey() string {
	return csCommand.userId
}

func (csCommand *CloudStopCommand) ActUponConnection() error {
	err := csCommand.connectionList.RemoveByKey(csCommand.userId)

	if err == nil {
		csCommand.logger.LogUserInfo("cloudstop", csCommand.userId, "Removed connection from %s", (*csCommand.connection).RemoteAddr())
	} else {
		csCommand.logger.LogWarning("cloudstop", "Could not remove connection: %s", err.Error())
	}

	return err
}

func (csCommand *CloudStopCommand) Close() {
}

func (csCommand *CloudStopCommand) Handle() (string, error) {
	var stop *dbgpXml.CloudStop

	err := csCommand.connectionList.RemoveByKey(csCommand.userId)

	if err == nil {
		csCommand.logger.LogUserInfo("cloudstop", csCommand.userId, "CloudStop::Handle: Removed connection for Cloud User from %s", (*csCommand.connection).RemoteAddr())
		stop = dbgpXml.NewCloudStop(true, csCommand.userId, nil)
	} else {
		csCommand.logger.LogUserWarning("cloudstop", csCommand.userId, "Could not remove connection: %s", err.Error())
		stop = dbgpXml.NewCloudStop(false, csCommand.userId, &dbgpXml.CloudStopError{ID: "ERR-10", Message: err.Error()})
	}

	return stop.AsXML()
}

/* cloudstop -u <userid> */
func CreateCloudStop(connectionsList *connections.ConnectionList, connection *net.Conn, arguments []string, logger logger.Logger) (DbgpCloudCommand, error) {
	csCommand := NewCloudStopCommand(connectionsList, connection, logger)

	expectValue := false
	expectValueFor := ""

	for _, value := range arguments {
		if expectValue {
			expectValue = false
			switch expectValueFor {
			case "-i":
				/* ignore */
			case "-u":
				csCommand.userId = value
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
	if csCommand.userId == "" {
		return nil, fmt.Errorf("No username was provided")
	}

	return csCommand, nil
}
