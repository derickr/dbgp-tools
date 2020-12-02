package dbgpxml

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	. "github.com/logrusorgru/aurora" // WTFPL
)

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
