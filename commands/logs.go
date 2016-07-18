package commands

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bradylove/v3-cli-plugin/models"
	"github.com/bradylove/v3-cli-plugin/util"
	"github.com/cloudfoundry/cli/cf/api/logs"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/uihelpers"
	"github.com/cloudfoundry/cli/plugin"
	consumer "github.com/cloudfoundry/loggregator_consumer"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

func Logs(cliConnection plugin.CliConnection, args []string) {
	appName := args[1]
	output, _ := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
	apps := models.V3Apps{}
	json.Unmarshal([]byte(strings.Join(output, "")), &apps)

	if len(apps.Apps) == 0 {
		fmt.Printf("App %s not found\n", appName)
		return
	}
	app := apps.Apps[0]

	messageQueue := logs.NewLoggregatorMessageQueue()

	bufferTime := 25 * time.Millisecond
	ticker := time.NewTicker(bufferTime)

	c := make(chan *logmessage.LogMessage)

	loggregatorEndpoint, err := cliConnection.LoggregatorEndpoint()
	util.ExitIfError(err)

	ssl, err := cliConnection.IsSSLDisabled()
	util.ExitIfError(err)
	tlsConfig := net.NewTLSConfig([]tls.Certificate{}, ssl)

	loggregatorConsumer := consumer.New(loggregatorEndpoint, tlsConfig, http.ProxyFromEnvironment)
	defer func() {
		loggregatorConsumer.Close()
		flushMessageQueue(c, messageQueue)
	}()

	onConnect := func() {
		fmt.Printf("Tailing logs for app %s...\r\n\r\n", appName)
	}
	loggregatorConsumer.SetOnConnectCallback(onConnect)

	accessToken, err := cliConnection.AccessToken()
	util.ExitIfError(err)

	logChan, err := loggregatorConsumer.Tail(app.Guid, accessToken)
	if err != nil {
		util.ExitIfError(err)
	}

	go func() {
		for _ = range ticker.C {
			flushMessageQueue(c, messageQueue)
		}
	}()

	go func() {
		for msg := range logChan {
			messageQueue.PushMessage(msg)
		}

		flushMessageQueue(c, messageQueue)
		close(c)
	}()

	for msg := range c {
		fmt.Printf("%s\r\n", logMessageOutput(msg, time.Local))
	}
}

func flushMessageQueue(c chan *logmessage.LogMessage, messageQueue *logs.LoggregatorMessageQueue) {
	messageQueue.EnumerateAndClear(func(m *logmessage.LogMessage) {
		c <- m
	})
}

func logMessageOutput(msg *logmessage.LogMessage, loc *time.Location) string {
	logHeader, coloredLogHeader := uihelpers.ExtractLogHeader(msg, loc)
	logContent := uihelpers.ExtractLogContent(msg, logHeader)

	return fmt.Sprintf("%s%s", coloredLogHeader, logContent)
}
