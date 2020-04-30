package dbgpXml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
)

type CloudStopError struct {
	XMLName xml.Name `xml:"error"`
	ID      string   `xml:"id,attr"`
	Message string   `xml:"message"`
}

type CloudStop struct {
	XMLName     xml.Name        `xml:"cloudstop"`
	XmlNS       string          `xml:"xmlns,attr"`
	XmlNSXdebug string          `xml:"xmlns:xdebug,attr"`
	Success     int             `xml:"success,attr"`
	UserID      string          `xml:"userid,attr"`
	Error       *CloudStopError `xml:"error,omitempty"`
}

func NewCloudStop(success bool, userID string, initError *CloudStopError) *CloudStop {
	successStr := 1

	if !success {
		successStr = 0
	}

	return &CloudStop{
		XmlNS:       "urn:debugger_protocol_v1",
		XmlNSXdebug: "https://xdebug.org/dbgp/xdebug",
		Success:     successStr,
		UserID:      userID,
		Error:       initError,
	}
}

func (cloudInit *CloudStop) AsXML() (string, error) {
	var output bytes.Buffer

	encoder := xml.NewEncoder(&output)

	err := encoder.Encode(cloudInit)

	if err != nil {
		return "", err
	}

	return xml.Header + output.String(), nil
}

func (init CloudStop) IsSuccess() bool {
	return !!(init.Success == 1)
}

func (init CloudStop) GetErrorMessage() string {
	if init.Error != nil {
		return init.Error.Message
	} else {
		return "no error"
	}
}

func (init CloudStop) ExpectMoreResponses() bool {
	return false
}

func (int CloudStop) ShouldCloseConnection() bool {
	return true
}

func (init CloudStop) String() string {
	if init.Success == 0 {
		return fmt.Sprintf("%s | %s: %s\n",
			Yellow(Bold("cloudstop")), Bold(Red("failure")), BrightRed(init.Error.Message))
	} else {
		return fmt.Sprintf("%s | Disconnected as %s\n\n",
			Yellow(Bold("cloudstop")),
			Bold(Yellow(init.UserID)))
	}
}
