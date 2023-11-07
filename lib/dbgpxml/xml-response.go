package dbgpxml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
	"strings"
)

/*
<response xmlns="urn:debugger_protocol_v1"
xmlns:xdebug="https://xdebug.org/dbgp/xdebug" command="feature_set"
transaction_id="1" feature="resolved_breakpoints" success="1"></response>
*/

type Property struct {
	XMLName      xml.Name   `xml:"property"`
	Name         string     `xml:"name,attr"`
	Fullname     string     `xml:"fullname,attr"`
	Type         string     `xml:"type,attr"`
	Classname    string     `xml:"classname,attr,omitempty"`
	HasChildren  bool       `xml:"children,attr,omitempty"`
	NumChildren  int        `xml:"numchildren,attr,omitempty"`
	Page         int        `xml:"page,attr,omitempty"`
	PageSize     int        `xml:"pagesize,attr,omitempty"`
	Encoding     string     `xml:"encoding,attr,omitempty"`
	Value        string     `xml:",chardata"`
	ExtName      string     `xml:"name,omitempty"`
	ExtFullName  string     `xml:"fullname,omitempty"`
	ExtClassname string     `xml:"classname,omitempty"`
	Children     []Property `xml:"property"`
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

func (s *Typemap) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	nameAttr := ""
	typeAttr := ""
	xsiTypeAttr := ""

	for _, attr := range start.Attr {
		if attr.Name.Space == "" && attr.Name.Local == "name" {
			nameAttr = attr.Value
		}
		if attr.Name.Space == "" && attr.Name.Local == "type" {
			typeAttr = attr.Value
		}
		if attr.Name.Space == "http://www.w3.org/2001/XMLSchema-instance" && attr.Name.Local == "type" {
			xsiTypeAttr = attr.Value
		}
	}

	d.Skip()
	*s = Typemap{Name: nameAttr, Type: typeAttr, XsiType: xsiTypeAttr}

	return nil
}

type Typemap struct {
	XMLName xml.Name `xml:"map"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
	XsiType string   `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
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
	FeatureName string    `xml:"feature_name,attr,omitempty"`
	Reason      string    `xml:"reason,attr,omitempty"`
	Encoding    string    `xml:"encoding,attr,omitempty"`
	Stack       []Stack   `xml:"stack,omitempty"`
	Contexts    []Context `xml:"context,omitempty"`
	Typemap     []Typemap `xml:"map,omitempty"`

	Message Message `xml:"message,omitempty"`

	Error *Error `xml:"error,omitempty"`

	Breakpoints []Breakpoint `xml:"breakpoint,omitempty"`
	Property    []Property   `xml:"property,omitempty"`

	Value string `xml:",cdata"`

	LastSourceBegin int
}

func formatContext(tid string, context Context) string {
	return fmt.Sprintf("%s | %d: %s\n", Black(tid), Yellow(context.ID), Bold(Green(context.Name)))
}

func formatFeatureGet(response Response) string {
	if response.Supported == 1 {
		return fmt.Sprintf("%s | %s: %s > %s\n", Black(response.TID), Bold(Yellow(response.FeatureName)), Bold(Green("supported")), Bold(Green(response.Value)))
	} else {
		return fmt.Sprintf("%s | %s: %s\n", Black(response.TID), Bold(Yellow(response.FeatureName)), Bold(Red("not supported")))
	}
}

func formatFeatureSet(response Response) string {
	if response.Success == 1 {
		return fmt.Sprintf("%s | %s: %s\n", Black(response.TID), Bold(Yellow(response.Feature)), Bold(Green("OK")))
	} else {
		return fmt.Sprintf("%s | %s: %s\n", Black(response.TID), Bold(Yellow(response.Feature)), Bold(Red("FAIL")))
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

func formatTypemap(tid string, typemap Typemap) string {
	if typemap.XsiType != "" {
		return fmt.Sprintf("%s | %s: %s (%s)\n", Black(tid), Yellow(typemap.Name), Bold(Green(typemap.Type)), Bold(Green(typemap.XsiType)))
	} else {
		return fmt.Sprintf("%s | %s: %s\n", Black(tid), Yellow(typemap.Name), Bold(Green(typemap.Type)))
	}
}

func formatProperty(tid string, leader string, prop Property) string {
	header := fmt.Sprintf("%s | ", Black(tid))

	if prop.Name == "" && prop.ExtName != "" {
		tmpName, _ := base64.StdEncoding.DecodeString(string(prop.ExtName))
		prop.Name = string(tmpName)
	}

	if prop.Classname == "" && prop.ExtClassname != "" {
		tmpClassname, _ := base64.StdEncoding.DecodeString(string(prop.ExtClassname))
		prop.Classname = string(tmpClassname)
	}

	content := leader + fmt.Sprintf("%s %s", prop.Type, Bold(Green(prop.Name)))

	switch prop.Type {
	case "object":
		content += fmt.Sprintf("(%s)", Green(prop.Classname))
	}

	if prop.HasChildren && prop.NumChildren > 0 {
		if prop.Type == "array" {
			content += ": [ \n"
		} else {
			content += " { \n"
		}
		for _, child := range prop.Children {
			content += leader + formatProperty(tid, leader+"  ", child)
		}
		content += header + leader
		if prop.Type == "array" {
			content += "]"
		} else {
			content += "}"
		}
	} else if prop.Type != "uninitialized" {
		value := []byte(prop.Value)
		if prop.Encoding == "base64" {
			value, _ = base64.StdEncoding.DecodeString(string(value))
		}
		switch prop.Type {
		case "null":
			/* do nothing */
		case "bool":
			if string(value) == "1" {
				content += fmt.Sprintf(": %s", Bold(Yellow("true")))
			} else {
				content += fmt.Sprintf(": %s", Bold(Yellow("false")))
			}
		case "array":
			content += fmt.Sprintf(": []")
		case "object":
			content += fmt.Sprintf(": {}")
		default:
			content += fmt.Sprintf(": %s", Bold(Yellow(value)))
		}
	}

	return header + content + "\n"
}

func formatSource(response Response) string {
	var content string

	value := []byte(response.Value)
	if response.Encoding == "base64" {
		value, _ = base64.StdEncoding.DecodeString(string(value))
	}

	if len(value) == 0 {
		return fmt.Sprintf("%s | %s\n", Black(response.TID), Bold(Red("The result was empty")))
	}

	lines := strings.Split(strings.TrimRight(string(value), " \n"), "\n")

	for i, line := range lines {
		content += fmt.Sprintf("%4d", Bold(Green((i+response.LastSourceBegin)))) + " " + line + "\n"
	}

	return content
}

func formatBreakpointSet(response Response) string {
	return fmt.Sprintf("%s | Breakpoint set with ID %s\n", Black(response.TID), Bold(Green(response.ID)))
}

func formatBreakpoint(TID string, brkpoint Breakpoint) string {
	return fmt.Sprintf("%s | ", Black(TID)) + brkpoint.String() + "\n"
}

func formatError(response Response) string {
	return fmt.Sprintf("%s | %s(%d): %s\n", Black(response.TID), Bold(Red("Error")), Red(response.Error.Code), BrightRed(Bold(response.Error.Message.Text)))
}

func (response Response) IsSuccess() bool {
	return (!!(response.Success == 1) ||
		response.Command == "detach")
}

func (response Response) GetErrorMessage() string {
	if response.Error != nil {
		return response.Error.Message.Text
	} else {
		return "no error"
	}
}

func (response Response) ExpectMoreResponses() bool {
	if response.Command == "break" {
		return true
	}
	return false
}

func (response Response) ShouldCloseConnection() bool {
	if response.Command == "detach" || response.Command == "stop" {
		return true
	}
	return false
}

func (response Response) String() string {
	output := fmt.Sprintf("%s | %s", Yellow(response.TID), Green(response.Command))

	if response.Status != "" && response.Reason != "" {
		output += fmt.Sprintf(" > %s/%s", Green(response.Status), Green(response.Reason))
	}

	output += "\n"

	if response.Error != nil && response.Error.Code != 0 {
		return output + formatError(response)
	}

	switch response.Command {
	case "breakpoint_get", "breakpoint_list", "breakpoint_remove", "breakpoint_update":
		for _, brkpoint := range response.Breakpoints {
			output += formatBreakpoint(response.TID, brkpoint)
		}

	case "breakpoint_set":
		output += formatBreakpointSet(response)

	case "context_get", "eval", "property_get":
		for _, prop := range response.Property {
			output += formatProperty(response.TID, "", prop)
		}

	case "context_names":
		for _, context := range response.Contexts {
			output += formatContext(response.TID, context)
		}

	case "feature_get":
		output += formatFeatureGet(response)

	case "feature_set", "stdout":
		output += formatFeatureSet(response)

	case "source":
		output += formatSource(response)

	case "stack_get":
		for _, frame := range response.Stack {
			output += formatStackFrame(response.TID, frame)
		}

	case "typemap_get":
		for _, typemap := range response.Typemap {
			output += formatTypemap(response.TID, typemap)
		}

	case "run", "step_into", "step_over", "step_out":
		if response.Status != "stopping" {
			output += fmt.Sprintf("%s | %s:%d\n", Black(response.TID), Bold(Green(response.Message.Filename)), Bold(Green(response.Message.LineNo)))
		}
	}

	return output
}
