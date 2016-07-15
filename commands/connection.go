package commands

import (
	"strings"

	"github.com/cloudfoundry/cli/plugin"
)

type Connection struct {
	plugin.CliConnection
}

func (c Connection) httpGet(url string) ([]byte, error) {
	output, err := c.CliCommandWithoutTerminalOutput("curl", url, "-X", "GET")

	return []byte(strings.Join(output, "")), err
}
