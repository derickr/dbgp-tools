package command

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/server"
	"github.com/derickr/dbgp-tools/lib/xml"
)

type ProxyStopCommand struct {
	connectionList *connections.ConnectionList
	ideKey         string
	logger         server.Logger
}

func NewProxyStopCommand(connectionList *connections.ConnectionList, logger server.Logger) *ProxyStopCommand {
	return &ProxyStopCommand{connectionList: connectionList, ideKey: "", logger: logger}
}

func (piCommand *ProxyStopCommand) GetName() string {
	return "proxystop"
}

func (piCommand *ProxyStopCommand) Handle() (string, error) {
	var stop *dbgpXml.ProxyStop

	err := piCommand.connectionList.RemoveByKey(piCommand.ideKey)

	if err == nil {
		piCommand.logger.LogUserInfo("proxystop", piCommand.ideKey, "Removed connection for IDE Key '%s'", piCommand.ideKey)
		stop = dbgpXml.NewProxyStop(true, piCommand.ideKey, nil)
	} else {
		piCommand.logger.LogUserWarning("proxystop", piCommand.ideKey, "Could not remove connection: %s", err.Error())
		stop = dbgpXml.NewProxyStop(false, piCommand.ideKey, &dbgpXml.ProxyInitError{ID: "ERR-02", Message: err.Error()})
	}

	return stop.AsXML()
}

/* proxystop -k PHPSTORM */
func CreateProxyStop(connectionList *connections.ConnectionList, arguments []string, logger server.Logger) (DbgpCommand, error) {
	piCommand := NewProxyStopCommand(connectionList, logger)

	expectValue := false
	expectValueFor := ""

	for _, value := range arguments {
		if expectValue {
			expectValue = false
			switch expectValueFor {
			case "-i":
				/* ignore */
			case "-k":
				piCommand.ideKey = value
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
	if piCommand.ideKey == "" {
		return nil, fmt.Errorf("No IDE key was provided")
	}

	return piCommand, nil
}
