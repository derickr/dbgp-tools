package protocol

import (
	"fmt"
	"github.com/derickr/dbgp-tools/lib/connections"
	"github.com/derickr/dbgp-tools/lib/logger"
	"github.com/derickr/dbgp-tools/lib/xml"
	. "github.com/logrusorgru/aurora" // WTFPL
	"io"
)

func UnregisterCloudClient(cloudDomain string, cloudPort string, cloudUser string, output io.Writer, logger logger.Logger) {
	var formattedResponse Response

	conn, err := connections.ConnectToCloud(cloudDomain, cloudPort, cloudUser, logger)

	if err != nil {
		fmt.Fprintf(output, "%s '%s': %s\n", BrightRed("Can not connect to Xdebug cloud at"), BrightYellow(cloudDomain), BrightRed(err))
		return
	}
	defer conn.Close()

	reader := NewDbgpClient(conn, false, logger)

	command := "cloudstop -u " + cloudUser
	reader.SendCommand(command)

	response, err, timedOut := reader.ReadResponse()

	if timedOut {
		fmt.Fprintf(output, "Time out: %s", err);
		return
	}

	if err != nil { // reading failed
		fmt.Fprintf(output, "Reading failed: %s", err);
		return
	}

	if !dbgpXml.IsValidXml(response) {
		fmt.Fprintf(output, "The received XML is not valid, closing connection: %s", response)
		return
	}

	formattedResponse = reader.FormatXML(response)

	if formattedResponse == nil {
		fmt.Fprintf(output, "Could not interpret XML, closing connection")
		return
	}
	fmt.Fprintln(output, formattedResponse)
}

