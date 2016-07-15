package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/bradylove/v3-cli-plugin/models"
	"github.com/bradylove/v3-cli-plugin/util"
	"github.com/cloudfoundry/cli/plugin"
	"strings"
)

type runningTask struct {
	guid    string
	command string
	state   string
	time    time.Time
}

func CancelTask(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	taskName := args[2]

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	util.ExitIfError(err)

	apps := models.V3Apps{}
	json.Unmarshal([]byte(strings.Join(output, "")), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}

	appGuid := apps.Apps[0].Guid

	tasksJson, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/tasks", appGuid))
	util.ExitIfError(err)

	tasks := models.V3Tasks{}
	err = json.Unmarshal([]byte(tasksJson[0]), &tasks)
	util.ExitIfError(err)

	var runningTasks []runningTask
	for _, task := range tasks.Tasks {
		if taskName == task.Name && task.State == "RUNNING" {
			runningTasks = append(runningTasks, runningTask{task.Guid, task.Command, task.State, task.UpdatedAt})
		}
	}

	if len(runningTasks) == 0 {
		fmt.Println("No running task found. Task name:", taskName)
		return
	} else if len(runningTasks) == 1 {
		output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/tasks/%s/cancel", runningTasks[0].guid), "-X", "PUT", "-d", "{}")
		util.ExitIfError(err)
		fmt.Println(output)
		return
	} else {
		fmt.Printf("Please select which task to cancel: \n\n")
		tasksTable := util.NewTable([]string{"#", "Task Name", "Command", "State", "Time"})
		for i, task := range runningTasks {
			tasksTable.Add(
				strconv.Itoa(i+1),
				taskName,
				task.command,
				task.state,
				fmt.Sprintf("%s", task.time.Format("Jan 2, 15:04:05 MST")),
			)
		}
		tasksTable.Print()

		var i int64 = -1
		var str string

		for i <= 0 || i > int64(len(runningTasks)) {
			fmt.Printf("\nSelect from above > ")
			fmt.Scanf("%s", &str)
			i, _ = strconv.ParseInt(str, 10, 32)
		}

		output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/tasks/%s/cancel", runningTasks[i-1].guid), "-X", "PUT", "-d", "{}")
		util.ExitIfError(err)
		fmt.Println(output)
	}
}
