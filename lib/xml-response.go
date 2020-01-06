package dbgp

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
	"golang.org/x/net/html/charset"
	"strings"
)

/*
<response xmlns="urn:debugger_protocol_v1"
xmlns:xdebug="https://xdebug.org/dbgp/xdebug" command="feature_set"
transaction_id="1" feature="resolved_breakpoints" success="1"></response>
*/

type Property struct {
	XMLName     xml.Name   `xml:"property"`
	Name        string     `xml:"name,attr"`
	Fullname    string     `xml:"fullname,attr"`
	Type        string     `xml:"type,attr"`
	Classname   string     `xml:"classname,attr,omitempty"`
	HasChildren bool       `xml:"children,attr,omitempty"`
	NumChildren int        `xml:"numchildren,attr,omitempty"`
	Page        int        `xml:"page,attr,omitempty"`
	PageSize    int        `xml:"pagesize,attr,omitempty"`
	Children    []Property `xml:"property"`
	Encoding    string     `xml:"encoding,attr,omitempty"`
	Value       string     `xml:",chardata"`
}

type Message struct {
	XMLName  xml.Name `xml:"message"`
	Filename string   `xml:"filename,attr"`
	LineNo   int      `xml:"lineno,attr"`
}

type Stack struct {
	XMLName  xml.Name `xml:"stack"`
	Where    string   `xml:"where,attr"`
	Level    int      `xml:"level,attr"`
	Type     string   `xml:"type,attr"`
	Filename string   `xml:"filename,attr"`
	LineNo   int      `xml:"lineno,attr"`
}

type ErrorMessage struct {
	XMLName xml.Name `xml:"message"`
	Text    string   `xml:",cdata"`
}

type Error struct {
	XMLName xml.Name     `xml:"error"`
	Code    int          `xml:"code,attr"`
	Message ErrorMessage `xml:"message"`
}

type Context struct {
	XMLName xml.Name `xml:"context"`
	ID      int      `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
}

type Expression struct {
	XMLName  xml.Name `xml:"expression"`
	Encoding string   `xml:"encoding,attr,omitempty"`
	Value    string   `xml:",chardata"`
}

type Breakpoint struct {
	XMLName      xml.Name   `xml:"breakpoint"`
	ID           int        `xml:"id,attr"`
	Type         string     `xml:"type,attr"`
	State        string     `xml:"state,attr,omitempty"`
	Resolved     string     `xml:"resolved,attr,omitempty"`
	Filename     string     `xml:"filename,attr,omitempty"`
	LineNo       int        `xml:"lineno,attr,omitempty"`
	Classname    string     `xml:"class,attr,omitempty"`
	Function     string     `xml:"function,attr,omitempty"`
	Exception    string     `xml:"exception,attr,omitempty"`
	HitValue     int        `xml:"hit_value,attr,omitempty"`
	HitCount     int        `xml:"hit_count,attr,omitempty"`
	HitCondition string     `xml:"hit_condition,attr,omitempty"`
	Expression   Expression `xml:"expression"`
}

type Response struct {
	XMLName     xml.Name  `xml:"response"`
	XmlNS       string    `xml:"xmlns,attr"`
	XmlNSXdebug string    `xml:"xmlns:xdebug,attr"`
	TID         string    `xml:"transaction_id,attr"`
	ID          string    `xml:"id,attr"`
	Command     string    `xml:"command,attr,omitempty"`
	Status      string    `xml:"status,attr,omitempty"`
	Success     int       `xml:"success,attr,omitempty"`
	Supported   int       `xml:"supported,attr,omitempty"`
	Feature     string    `xml:"feature,attr,omitempty"`
	Reason      string    `xml:"reason,attr,omitempty"`
	Stack       []Stack   `xml:"stack,omitempty"`
	Contexts    []Context `xml:"context,omitempty"`

	Message Message `xml:"message,omitempty"`

	Error Error `xml:"error,omitempty"`

	Breakpoints []Breakpoint `xml:"breakpoint,omitempty"`
	Property    []Property   `xml:"property,omitempty"`

	Value string `xml:",cdata"`
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

func formatFeatureGet(response Response) string {
	if response.Supported == 1 {
		return fmt.Sprintf("%s | %s > %s\n", Black(response.TID), Bold(Green("supported")), Bold(Green(response.Value)))
	} else {
		return fmt.Sprintf("%s | %s\n", Black(response.TID), Bold(Red("not supported")))
	}
}

func formatFeatureSet(response Response) string {
	if response.Success == 1 {
		return fmt.Sprintf("%s | %s\n", Black(response.TID), Bold(Green("OK")))
	} else {
		return fmt.Sprintf("%s | %s\n", Black(response.TID), Bold(Red("-")))
	}
}

func formatLocation(filename string, lineno int) string {
	return fmt.Sprintf("%s:%d", Bold(Green(filename)), Bold(Green(lineno)))
}

func formatFunction(class string, method string) string {
	if class != "" {
		return fmt.Sprintf("%s->%s", Bold(BrightBlue(class)), Bold(BrightBlue(method)))
	} else {
		return fmt.Sprintf("%s", Bold(BrightBlue(method)))
	}
}

func formatStackFrame(tid string, frame Stack) string {
	return fmt.Sprintf("%s | %d: %s: %s\n", Black(tid), Yellow(frame.Level), formatLocation(frame.Filename, frame.LineNo), Bold(Yellow(frame.Where)))
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
			content += leader + formatProperty(tid, leader+"  ", child)
		}
		content += header + leader + "}"
	} else if prop.Type != "uninitialized" {
		value := []byte(prop.Value)
		if prop.Encoding == "base64" {
			value, _ = base64.StdEncoding.DecodeString(string(value))
		}
		content += fmt.Sprintf(": %s", Bold(Yellow(value)))
	}

	return header + content + "\n"
}

func formatBreakpointSet(response Response) string {
	return fmt.Sprintf("%s | Breakpoint set with ID %s\n", Black(response.TID), Bold(Green(response.ID)))
}

func (brkpoint Breakpoint) String() string {
	content := ""

	switch brkpoint.State {
	case "enabled":
		content += fmt.Sprintf("%s ", Bold(BrightGreen("●")))
	case "disabled":
		content += fmt.Sprintf("%s ", Bold(BrightRed("○")))
	case "temporary":
		if brkpoint.HitCount == 0 {
			content += fmt.Sprintf("%s ", Bold(BrightYellow("◐")))
		} else {
			content += fmt.Sprintf("%s ", Bold(BrightRed("◐")))
		}
	}

	content += fmt.Sprintf("(%d", Yellow(brkpoint.HitCount))
	if brkpoint.HitCondition != "" {
		content += fmt.Sprintf(" %s%d", Green(brkpoint.HitCondition), Yellow(brkpoint.HitValue))
	}
	content += ") "

	content += fmt.Sprintf("%d %s: ", Bold(Green(brkpoint.ID)), Yellow(brkpoint.Type))

	switch brkpoint.Type {
	case "condition", "line":
		content += formatLocation(brkpoint.Filename, brkpoint.LineNo)

	case "call", "return":
		content += formatFunction(brkpoint.Classname, brkpoint.Function)

	case "exception":
		content += fmt.Sprintf("%s", Bold(Green(brkpoint.Exception)))
	}

	if brkpoint.Expression.Value != "" {
		value := []byte(brkpoint.Expression.Value)
		if brkpoint.Expression.Encoding == "base64" {
			value, _ = base64.StdEncoding.DecodeString(string(value))
		}
		content += fmt.Sprintf(" cond: %s", Yellow(value))
	}

	return content
}

func formatBreakpoint(TID string, brkpoint Breakpoint) string {
	return fmt.Sprintf("%s | ", Black(TID)) + brkpoint.String() + "\n"
}

func formatError(response Response) string {
	return fmt.Sprintf("%s | %s(%d): %s\n", Black(response.TID), Bold(Red("Error")), Red(response.Error.Code), BrightRed(Bold(response.Error.Message.Text)))
}

func (response Response) String() string {
	output := fmt.Sprintf("%s | %s", Yellow(response.TID), Green(response.Command))

	if response.Status != "" && response.Reason != "" {
		output += fmt.Sprintf(" > %s/%s", Green(response.Status), Green(response.Reason))
	}

	output += "\n"

	if response.Error.Code != 0 {
		return output + formatError(response)
	}

	switch response.Command {
	case "breakpoint_get", "breakpoint_list", "breakpoint_remove", "breakpoint_update":
		for _, brkpoint := range response.Breakpoints {
			output += formatBreakpoint(response.TID, brkpoint)
		}

	case "breakpoint_set":
		output += formatBreakpointSet(response)

	case "context_get", "property_get":
		for _, prop := range response.Property {
			output += formatProperty(response.TID, "", prop)
		}

	case "context_names":
		for _, context := range response.Contexts {
			output += formatContext(response.TID, context)
		}

	case "feature_get":
		output += formatFeatureGet(response)

	case "feature_set":
		output += formatFeatureSet(response)

	case "stack_get":
		for _, frame := range response.Stack {
			output += formatStackFrame(response.TID, frame)
		}

	case "run", "step_into", "step_over":
		if response.Status != "stopping" {
			output += fmt.Sprintf("%s | %s:%d\n", Black(response.TID), Bold(Green(response.Message.Filename)), Bold(Green(response.Message.LineNo)))
		}
	}

	return output
}
