package commands

import (
	"encoding/json"
	"fmt"

	"github.com/bradylove/v3-cli-plugin/models"
	"github.com/bradylove/v3-cli-plugin/util"
)

func Tasks(conn Connection, args []string) {
	appName := args[1]
	fmt.Printf("Listing tasks for app %s...\n", appName)

	resp, err := conn.httpGet(fmt.Sprintf("/v3/apps?names=%s", appName))
	util.ExitIfError(err)

	var apps models.V3Apps
	err = json.Unmarshal(resp, &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}

	resp, err = conn.httpGet(fmt.Sprintf("/v3/apps/%s/tasks", apps.Apps[0].Guid))
	util.ExitIfError(err)

	var tasks models.V3Tasks
	err = json.Unmarshal(resp, &tasks)
	util.ExitIfError(err)

	if len(tasks.Tasks) == 0 {
		fmt.Println("No v3 tasks found.")
		return
	}

	tasksTable := util.NewTable([]string{("name"), ("command"), ("state")})
	for _, v := range tasks.Tasks {
		tasksTable.Add(v.Name, v.Command, v.State)
	}
	tasksTable.Print()
}
