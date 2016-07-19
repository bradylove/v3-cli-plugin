package commands

import (
	"fmt"

	"github.com/bradylove/v3-cli-plugin/util"
)

func RunTask(conn Connection, args []string) {
	appName := args[1]
	taskName := args[2]
	taskCommand := args[3]

	fmt.Printf("Running task %s on app %s...\n", taskName, appName)

	app, err := conn.findAppByName(appName)
	util.ExitIfError(err)

	_, err = conn.createTask(app, taskName, taskCommand)
	util.ExitIfError(err)

	fmt.Println("OK")
}
