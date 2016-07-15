package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bradylove/v3-cli-plugin/models"
	"github.com/bradylove/v3-cli-plugin/util"
	"github.com/cloudfoundry/cli/plugin"
)

func Processes(cliConnection plugin.CliConnection, args []string) {
	mySpace, err := cliConnection.GetCurrentSpace()
	util.ExitIfError(err)

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", "v3/processes?per_page=5000", "-X", "GET")
	util.ExitIfError(err)

	var processes models.V3ProcessesModel
	err = json.Unmarshal([]byte(strings.Join(output, "")), &processes)
	util.ExitIfError(err)

	if len(processes.Processes) > 0 {
		processesTable := util.NewTable([]string{("app"), ("type"), ("instances"), ("memory in MB"), ("disk in MB")})
		for _, v := range processes.Processes {
			if strings.Contains(v.Links.Space.Href, mySpace.Guid) {
				appName := "N/A"
				if v.Links.App.Href != "/v3/apps/" {
					appName = strings.Split(v.Links.App.Href, "/v3/apps/")[1]
				}
				processesTable.Add(
					appName,
					v.Type,
					strconv.Itoa(v.Instances),
					strconv.Itoa(v.Memory)+"MB",
					strconv.Itoa(v.Disk)+"MB",
				)
			}
		}
		fmt.Println("print table?")
		processesTable.Print()
	} else {
		fmt.Println("No v3 processes found.")
	}
}
