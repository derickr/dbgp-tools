package command

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/server"
	"github.com/derickr/dbgp-tools/lib/xml"
	"strconv"
)

type ProxyInitCommand struct {
	connectionList    *connections.ConnectionList
	ideKey            string
	ipAddress         string
	port              int
	multipleSupported bool
	ssl               bool
}

func NewProxyInitCommand(ipAddress string, connectionList *connections.ConnectionList) *ProxyInitCommand {
	return &ProxyInitCommand{connectionList: connectionList, ideKey: "", ipAddress: ipAddress, port: 9000, multipleSupported: false}
}

func (piCommand *ProxyInitCommand) GetName() string {
	return "proxyinit"
}

func (piCommand *ProxyInitCommand) Handle() (string, error) {
	var init *dbgpXml.ProxyInit

	conn := connections.NewConnection(piCommand.ideKey, piCommand.ipAddress, strconv.Itoa(piCommand.port), piCommand.ssl, nil)
	err := piCommand.connectionList.Add(conn)

	if err == nil {
		if piCommand.ssl {
			fmt.Printf("  - Added SSL connection for IDE Key '%s': %s:%d\n", piCommand.ideKey, piCommand.ipAddress, piCommand.port)
		} else {
			fmt.Printf("  - Added connection for IDE Key '%s': %s:%d\n", piCommand.ideKey, piCommand.ipAddress, piCommand.port)
		}
		init = dbgpXml.NewProxyInit(true, piCommand.ideKey, piCommand.ipAddress, piCommand.port, piCommand.ssl, nil)
	} else {
		fmt.Printf("  - Could not add connection: %s\n", err.Error())
		init = dbgpXml.NewProxyInit(false, piCommand.ideKey, piCommand.ipAddress, piCommand.port, piCommand.ssl, &dbgpXml.ProxyInitError{ID: "ERR-01", Message: err.Error()})
	}

	return init.AsXML()
}

/* proxyinit -p 9000 -k PHPSTORM -m 1 -s ? */
func CreateProxyInit(ipAddress string, connectionList *connections.ConnectionList, arguments []string) (DbgpCommand, error) {
	piCommand := NewProxyInitCommand(ipAddress, connectionList)

	expectValue := false
	expectValueFor := ""

	for _, value := range arguments {
		if expectValue {
			expectValue = false
			switch expectValueFor {
			case "-i":
				/* ignore */
			case "-p":
				port, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("Port number (%s) given for '-p' is not a valid number: %s", value, err.Error())
				}
				piCommand.port = port
			case "-k":
				piCommand.ideKey = value
			case "-m":
				if value == "1" {
					piCommand.multipleSupported = true
				}
			case "-s":
				ssl, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("SSL value (%s) given for '-s' is not a valid value: %s", value, err.Error())
				}
				piCommand.ssl = ssl == 1
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
