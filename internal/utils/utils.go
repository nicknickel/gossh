package utils

import (
	"strings"
)

func NewReceiveTemplate(remoteSrc string, destName string) []string {
	return []string{"scp", "-rp", "{{.FinalAddr}}:" + remoteSrc, destName + "_{{.CleanTitle}}"}
}

func NewConnectTemplate() []string {
	return []string{"ssh", "{{.FinalAddr}}"}
}

func NewSendTemplate(src string, remoteDest string) []string {
	return []string{"scp", "-rp", src, "{{.FinalAddr}}:" + remoteDest}
}

func NewCommandTemplate(cmdToRun string) (osCommand []string) {
	osCommand = append(osCommand, "ssh", "{{.FinalAddr}}")
	osCommand = append(osCommand, strings.Split(cmdToRun, " ")...)
	return
}
