package dbgpxml

import (
	"strings"
)

func IsValidXml(xml string) bool {
	return strings.HasPrefix(xml, "<?xml")
}
