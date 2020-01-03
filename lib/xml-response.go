package dbgp

import (
    "encoding/xml"
	"golang.org/x/net/html/charset"
	. "github.com/logrusorgru/aurora" // WTFPL
    "fmt"
	"strings"
	"encoding/base64"
)

/*
<response xmlns="urn:debugger_protocol_v1"
xmlns:xdebug="https://xdebug.org/dbgp/xdebug" command="feature_set"
transaction_id="1" feature="resolved_breakpoints" success="1"></response>
*/

type Property struct {
	XMLName xml.Name `xml:"property"`
	Name string            `xml:"name,attr"`
	Fullname string  `xml:"fullname,attr"`
	Type string      `xml:"type,attr"`
	Classname string `xml:"classname,attr,omitempty"`
	HasChildren bool `xml:"children,attr,omitempty"`
	NumChildren int  `xml:"numchildren,attr,omitempty"`
	Page int         `xml:"page,attr,omitempty"`
	PageSize int     `xml:"pagesize,attr,omitempty"`
	Children []Property `xml:"property"`
	Encoding string  `xml:"encoding,attr,omitempty"`
	Value string     `xml:",chardata"`
}

type Message struct {
	XMLName xml.Name `xml:"message"`
	Filename string `xml:"filename,attr"`
	LineNo   int    `xml:"lineno,attr"`
}

type Stack struct {
	XMLName xml.Name `xml:"stack"`
	Where   string   `xml:"where,attr"`
	Level   int      `xml:"level,attr"`
	Type   string    `xml:"type,attr"`
	Filename string  `xml:"filename,attr"`
	LineNo   int     `xml:"lineno,attr"`
}

type ErrorMessage struct {
	XMLName xml.Name `xml:"message"`
	Text string `xml:",cdata"`
}

type Error struct {
	XMLName xml.Name `xml:"error"`
	Code  int `xml:"code,attr"`
	Message ErrorMessage `xml:"message"`
}

type Context struct {
	XMLName xml.Name `xml:"context"`
	ID  int `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type Response struct {
    XMLName xml.Name `xml:"response"`
	XmlNS   string   `xml:"xmlns,attr"`
	XmlNSXdebug string `xml:"xmlns:xdebug,attr"`
    TID  string   `xml:"transaction_id,attr"`
    Command string   `xml:"command,attr,omitempty"`
	Status  string   `xml:"status,attr,omitempty"`
    Success string   `xml:"success,attr,omitempty"`
	Feature string   `xml:"feature,attr,omitempty"`
	Reason  string   `xml:"reason,attr,omitempty"`
	Stack   []Stack  `xml:"stack,omitempty"`
	Contexts []Context `xml:"context,omitempty"`

	Message Message  `xml:"message,omitempty"`

	Error Error    `xml:"error,omitempty"`

	Property []Property `xml:"property,omitempty"`
}

func ParseResponseXML(rawXmlData string) (Response, error) {
	response := Response{}

	reader := strings.NewReader(rawXmlData)

	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(&response)

	if err != nil {
		return response, err
	}

	return response, nil
}

func formatContext(tid string, context Context) string {
	return fmt.Sprintf("%s | %d: %s\n", Black(tid), Yellow(context.ID), Bold(Green(context.Name)))
}

func formatStackFrame(tid string, frame Stack) string {
	return fmt.Sprintf("%s | %d: %s:%d: %s\n", Black(tid), Yellow(frame.Level), Bold(Green(frame.Filename)), Bold(Green(frame.LineNo)), Bold(Yellow(frame.Where)))
}

func formatProperty(tid string, leader string, prop Property) string {
	header := fmt.Sprintf("%s | ", Black(tid))

	content := leader + fmt.Sprintf("%s %s", prop.Type, Bold(Green(prop.Name)))

	switch prop.Type {
		case "object":
			content += fmt.Sprintf("(%s)", Green(prop.Classname))
	}

	if prop.HasChildren {
		content += " { \n"
		for _, child := range prop.Children {
			content += leader + formatProperty(tid, leader + "  ", child)
		}
		content += header + leader + "}"
	} else if prop.Type != "uninitialized" {
		value := []byte(prop.Value)
		if (prop.Encoding == "base64") {
			value, _ = base64.StdEncoding.DecodeString(string(value))
		}
		content += fmt.Sprintf(": %s", Bold(Yellow(value)))
	}

	return header + content + "\n"
}

func formatError(response Response) string {
	return fmt.Sprintf("%s | %s", Yellow(response.TID), Green(response.Command)) +
		fmt.Sprintf("\n%s | %s(%d): %s\n", Black(response.TID), Bold(Red("Error")), Red(response.Error.Code), BrightRed(Bold(response.Error.Message.Text)))
}

func (response Response) String() string {
	output := fmt.Sprintf("%s | %s", Yellow(response.TID), Green(response.Command))

	if response.Error.Code != 0 {
		return formatError(response)
	}

	switch response.Command {
		case "status", "step_into":
			output += fmt.Sprintf(" > %s/%s", Green(response.Status), Green(response.Reason))
	}

	output += "\n"

	switch response.Command {
		case "context_get", "property_get":
			for _, prop := range response.Property {
				output += formatProperty(response.TID, "", prop)
			}

		case "context_names":
			for _, context := range response.Contexts {
				output += formatContext(response.TID, context)
			}

		case "stack_get":
			for _, frame := range response.Stack {
				output += formatStackFrame(response.TID, frame)
			}

		case "step_into":
			if response.Status != "stopping" {
				output += fmt.Sprintf("%s | %s:%d\n", Black(response.TID), Bold(Green(response.Message.Filename)), Bold(Green(response.Message.LineNo)))
			}
	}

	return output
}
