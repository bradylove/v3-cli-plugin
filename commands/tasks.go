package commands

import (
	"encoding/json"
	"fmt"

	. "github.com/bradylove/v3-cli-plugin/models"
	. "github.com/bradylove/v3-cli-plugin/util"
	"github.com/cloudfoundry/cli/plugin"
)

func Tasks(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	fmt.Printf("Listing tasks for app %s...\n", appName)

	output, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	apps := V3AppsModel{}
	json.Unmarshal([]byte(output[0]), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}

	appGuid := apps.Apps[0].Guid

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/tasks", appGuid), "-X", "GET")
	ExitIfError(err)
	tasks := V3TasksModel{}
	err = json.Unmarshal([]byte(output[0]), &tasks)
	ExitIfError(err)

	if len(tasks.Tasks) > 0 {
		tasksTable := NewTable([]string{("name"), ("command"), ("state")})
		for _, v := range tasks.Tasks {
			tasksTable.Add(
				v.Name,
				v.Command,
				v.State,
			)
		}
		tasksTable.Print()
	} else {
		fmt.Println("No v3 tasks found.")
	}
}
