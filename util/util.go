package util

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/plugin"
)

func Poll(cliConnection plugin.CliConnection, endpoint string, desired string, timeout time.Duration, timeoutMessage string) {
	timeElapsed := 0 * time.Second
	for timeElapsed < timeout {
		output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", endpoint, "-X", "GET")
		ExitIfError(err)

		if strings.Contains(strings.Join(output, ""), desired) {
			return
		}

		timeElapsed = timeElapsed + 1*time.Second
		time.Sleep(1 * time.Second)
	}

	ExitIfError(errors.New(timeoutMessage))
}

func ExitIfError(err error) {
	if err != nil {
		fmt.Println("Error Will Robinson!: ", err.Error())
		os.Exit(1)
	}
}
