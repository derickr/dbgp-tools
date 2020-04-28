package dbgpXml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
	"strconv"
)

type CloudInitError struct {
	XMLName xml.Name `xml:"error"`
	ID      string   `xml:"id,attr"`
	Message string   `xml:"message"`
}

type CloudInitAccountInfo interface {
	AsDbgpXmlType()	*AccountInfo
}

type AccountInfo struct {
	Name                 string `xml:"name,attr"`
	Email                string `xml:"name,omit"`
	Uid                  string `xml:"uid,attr"`
	ConnectionsRemaining int    `xml:"remaining,attr"`
}

type CloudInit struct {
	XMLName     xml.Name        `xml:"cloudinit"`
	XmlNS       string          `xml:"xmlns,attr"`
	XmlNSXdebug string          `xml:"xmlns:xdebug,attr"`
	Success     int             `xml:"success,attr"`
	UserID      string          `xml:"userid,attr"`
	Error       *CloudInitError `xml:"error,omitempty"`
	AccountInfo *AccountInfo    `xml:"accountInfo,omitempty"`
}

func NewCloudInit(success bool, userID string, initError *CloudInitError, accountInfoValid bool, accountInfo CloudInitAccountInfo) *CloudInit {
	var initAccountInfo *AccountInfo = nil

	successStr := 1

	if !success {
		successStr = 0
	}

	if accountInfoValid {
		initAccountInfo = accountInfo.AsDbgpXmlType()
	}

	return &CloudInit{
		XmlNS:       "urn:debugger_protocol_v1",
		XmlNSXdebug: "https://xdebug.org/dbgp/xdebug",
		Success:     successStr,
		UserID:      userID,
		Error:       initError,
		AccountInfo: initAccountInfo,
	}
}

func (cloudInit *CloudInit) AsXML() (string, error) {
	var output bytes.Buffer

	encoder := xml.NewEncoder(&output)

	err := encoder.Encode(cloudInit)

	if err != nil {
		return "", err
	}

	return xml.Header + output.String(), nil
}

func (init CloudInit) IsSuccess() bool {
	return !!(init.Success == 1)
}

func (init CloudInit) GetErrorMessage() string {
	if init.Error != nil {
		return init.Error.Message
	} else {
		return "no error"
	}
}

func (init CloudInit) ExpectMoreResponses() bool {
	return true
}

func (int CloudInit) ShouldCloseConnection() bool {
	return false
}

func (init CloudInit) String() string {
	if init.Success == 0 {
		return fmt.Sprintf("%s | %s: %s\n",
			Yellow(Bold("cloudinit")), Bold(Red("failure")), BrightRed(init.Error.Message))
	} else {
		connectionsLeft := "??"

		if init.AccountInfo != nil {
			connectionsLeft = strconv.Itoa(init.AccountInfo.ConnectionsRemaining)
		}
		return fmt.Sprintf("%s | Connected as %s | %s connections remaining\n\n%s\n",
			Yellow(Bold("cloudinit")),
			Bold(Yellow(init.UserID)),
			BrightGreen(connectionsLeft),
			BrightGreen(Bold("Waiting for incoming connection...")))
	}
}
