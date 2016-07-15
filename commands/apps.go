package commands

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bradylove/v3-cli-plugin/models"
	"github.com/bradylove/v3-cli-plugin/util"
)

func Apps(conn Connection, args []string) {
	mySpace, err := conn.GetCurrentSpace()
	util.ExitIfError(err)

	resp, err := conn.httpGet(fmt.Sprintf("v3/apps?space_guids=%s", mySpace.Guid))
	util.ExitIfError(err)

	var apps models.V3AppsModel
	err = json.Unmarshal(resp, &apps)
	util.ExitIfError(err)

	if len(apps.Apps) == 0 {
		fmt.Println("No v3 apps found.")
		return
	}

	appsTable := util.NewTable([]string{"name", "total_desired_instances"})
	for _, v := range apps.Apps {
		appsTable.Add(v.Name, strconv.Itoa(v.Instances))
	}
	appsTable.Print()
}
