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
	userId         string
	needsRemoving  bool
}

func NewCloudInitCommand(connectionList *connections.ConnectionList, connection *net.Conn) *CloudInitCommand {
	return &CloudInitCommand{connectionList: connectionList, connection: connection, userId: "", needsRemoving: true}
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
		fmt.Printf("  - Added connection for Cloud User '%s' from %s\n", ciCommand.GetName(), (*ciCommand.connection).RemoteAddr())
	} else {
		ciCommand.needsRemoving = false
		fmt.Printf("  - Could not add connection: %s\n", err.Error())
	}

	return err
}

func (ciCommand *CloudInitCommand) Close() {
	if ciCommand.needsRemoving {
		fmt.Printf("  - Removed connection for Cloud User '%s' from %s\n", ciCommand.userId, (*ciCommand.connection).RemoteAddr())
		ciCommand.connectionList.RemoveByKey(ciCommand.userId)
	}
}

func (ciCommand *CloudInitCommand) Handle() (string, error) {
	var init *dbgpXml.CloudInit

	conn := connections.NewConnection(ciCommand.userId, "", "", true, ciCommand.connection)
	err := ciCommand.connectionList.Add(conn)

	if err == nil {
		fmt.Printf("  - Added connection for Cloud User '%s' from %s\n", ciCommand.userId, (*ciCommand.connection).RemoteAddr())
		init = dbgpXml.NewCloudInit(true, ciCommand.userId, nil)
	} else {
		ciCommand.needsRemoving = false
		fmt.Printf("  - Could not add connection: %s\n", err.Error())
		init = dbgpXml.NewCloudInit(false, ciCommand.userId, &dbgpXml.CloudInitError{ID: "ERR-01", Message: err.Error()})
	}

	return init.AsXML()
}

/* cloudinit -u <userid> */
func CreateCloudInit(connectionsList *connections.ConnectionList, connection *net.Conn, arguments []string) (DbgpCloudInitCommand, error) {
	ciCommand := NewCloudInitCommand(connectionsList, connection)

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
