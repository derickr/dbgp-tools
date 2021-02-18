package protocol

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	. "github.com/logrusorgru/aurora" // WTFPL
	"io"
)

func UnregisterCloudClient(cloudDomain string, cloudPort string, cloudUser string, output io.Writer, logger logger.Logger) {
	conn, err := connections.ConnectToCloud(cloudDomain, cloudPort, cloudUser, logger)
	if err != nil {
		fmt.Fprintf(output, "%s '%s': %s\n", BrightRed("Can not connect to Xdebug cloud at"), BrightYellow(cloudDomain), BrightRed(err))
		return
	}
	defer conn.Close()

	command := "cloudstop -u " + cloudUser

	_ = RunAndQuit(conn, command, output, logger, false)
}
