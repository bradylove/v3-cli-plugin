package main

import (
	"fmt"

	"github.com/bradylove/v3-cli-plugin/commands"
	"github.com/cloudfoundry/cli/plugin"
)

const (
	commandPush        = "v3-push"
	commandApps        = "v3-apps"
	commandProcesses   = "v3-processes"
	commandDelete      = "v3-delete"
	commandLogs        = "v3-logs"
	commandTasks       = "v3-tasks"
	commandRunTask     = "v3-run-task"
	commandCancelTask  = "v3-cancel-task"
	commandBindService = "v3-bind-service"
)

type V3Plugin struct{}

func main() {
	plugin.Start(new(V3Plugin))
}

func (v3plugin *V3Plugin) Run(cliConnection plugin.CliConnection, args []string) {
	conn := commands.Connection{cliConnection}

	switch args[0] {
	case commandPush:
		commands.Push(conn, args)
	case commandApps:
		if len(args) != 1 {
			fmt.Printf("Wrong number of argument, type `cf %s -h` for help\n", args[0])
			return
		}
		commands.Apps(conn, args)
	case commandProcesses:
		if len(args) != 1 {
			fmt.Printf("Wrong number of argument, type `cf %s -h` for help\n", args[0])
			return
		}
		commands.Processes(cliConnection, args)
	case commandDelete:
		if len(args) != 2 {
			fmt.Printf("Wrong number of argument, type `cf %s -h` for help\n", args[0])
			return
		}
		commands.Delete(cliConnection, args)
	case commandLogs:
		fmt.Println(commandLogs, "is temporarily disabled.")
		// if len(args) == 2 {
		// 	commands.Logs(cliConnection, args)
		// } else {
		// 	fmt.Printf("Wrong number of argument, type `cf %s -h` for help\n", args[0])
		// }
	case commandTasks:
		if len(args) != 2 {
			fmt.Printf("Wrong number of argument, type `cf %s -h` for help\n", args[0])
			return
		}
		commands.Tasks(commands.Connection{cliConnection}, args)
	case commandRunTask:
		if len(args) != 4 {
			fmt.Printf("Wrong number of argument, type `cf %s -h` for help\n", args[0])
			return
		}
		commands.RunTask(cliConnection, args)
	case commandCancelTask:
		if len(args) != 3 {
			fmt.Printf("Wrong number of argument, type `cf %s -h` for help\n", args[0])
			return
		}
		commands.CancelTask(cliConnection, args)
	case commandBindService:
		if len(args) < 3 {
			fmt.Printf("Wrong number of argument, type `cf %s -h` for help\n", args[0])
			return
		}
		commands.BindService(cliConnection, args)
	}
}

func (v3plugin *V3Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "v3_beta",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 4,
			Build: 20,
		},
		Commands: []plugin.Command{
			{
				Name:     commandApps,
				HelpText: "displays all v3 apps",
				UsageDetails: plugin.Usage{
					Usage:   commandApps,
					Options: map[string]string{},
				},
			},
			{
				Name:     commandPush,
				HelpText: "pushes current dir as a v3 process",
				UsageDetails: plugin.Usage{
					Usage: fmt.Sprintf("%s APPNAME", commandPush),
					Options: map[string]string{
						"p":  "path to dir or zip to push",
						"b":  "custom buildpack by name or Git URL",
						"di": "path to docker image to push",
					},
				},
			},
			{
				Name:     commandDelete,
				HelpText: "delete a v3 app",
				UsageDetails: plugin.Usage{
					Usage:   fmt.Sprintf("%s APPNAME", commandDelete),
					Options: map[string]string{},
				},
			},
			{
				Name:     commandProcesses,
				HelpText: "displays all v3 processes",
				UsageDetails: plugin.Usage{
					Usage:   fmt.Sprintf("%s ", commandProcesses),
					Options: map[string]string{},
				},
			},
			{
				Name:     commandLogs,
				HelpText: "tail logs for a v3 app",
				UsageDetails: plugin.Usage{
					Usage:   fmt.Sprintf("%s APPNAME", commandLogs),
					Options: map[string]string{},
				},
			},
			{
				Name:     commandBindService,
				HelpText: "bind a service instance to a v3 app",
				UsageDetails: plugin.Usage{
					Usage: fmt.Sprintf("%s APPNAME SERVICEINSTANCE", commandBindService),
					Options: map[string]string{
						"c": "parameters as json",
					},
				},
			},
			{
				Name:     commandTasks,
				HelpText: "list tasks for a v3 app",
				UsageDetails: plugin.Usage{
					Usage:   fmt.Sprintf("%s APPNAME", commandTasks),
					Options: map[string]string{},
				},
			},
			{
				Name:     commandRunTask,
				HelpText: "run a task on a v3 app",
				UsageDetails: plugin.Usage{
					Usage:   fmt.Sprintf("%s APPNAME TASKNAME COMMAND", commandRunTask),
					Options: map[string]string{},
				},
			},
			{
				Name:     commandCancelTask,
				HelpText: "cancel a task on a v3 app",
				UsageDetails: plugin.Usage{
					Usage: fmt.Sprintf("%s APPNAME TASKNAME", commandCancelTask),
				},
			},
		},
	}
}
