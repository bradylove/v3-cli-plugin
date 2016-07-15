package commands

import (
	"encoding/json"
	"fmt"

	. "github.com/bradylove/v3-cli-plugin/models"
	"github.com/cloudfoundry/cli/plugin"
	"strings"
)

func Delete(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	fmt.Printf("Deleting app %s...\n", appName)

	output, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	apps := V3Apps{}
	json.Unmarshal([]byte(strings.Join(output, "")), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}

	if _, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s", apps.Apps[0].Guid), "-X", "DELETE"); err != nil {
		fmt.Printf("Failed to delete app %s\n", appName)
		return
	}

	fmt.Println("OK")
}
