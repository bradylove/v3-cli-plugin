package commands

import (
	"strings"

	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/bradylove/v3-cli-plugin/models"
	"github.com/bradylove/v3-cli-plugin/util"
	"github.com/cloudfoundry/cli/cf/appfiles"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/gofileutils/fileutils"
)

type Connection struct {
	plugin.CliConnection
}

func (c Connection) httpGet(url string) ([]byte, error) {
	output, err := c.CliCommandWithoutTerminalOutput("curl", url, "-X", "GET")

	return []byte(strings.Join(output, "")), err
}

func (c Connection) httpPost(url, body string) ([]byte, error) {
	output, err := c.CliCommandWithoutTerminalOutput("curl", url, "-X", "POST", "-d", body)

	return []byte(strings.Join(output, "")), err
}

func (c Connection) findAppByName(name string) (models.V3App, error) {
	resp, err := c.httpGet(fmt.Sprintf("/v3/apps?names=%s", name))
	if err != nil {
		return models.V3App{}, err
	}

	apps := models.V3Apps{}
	err = json.Unmarshal(resp, &apps)
	if err != nil {
		return models.V3App{}, err
	}

	if len(apps.Apps) == 0 {
		return models.V3App{}, fmt.Errorf("App %s not found\n", name)
	}

	return apps.Apps[0], nil
}

func (c Connection) createV3App(body string) (models.V3App, error) {
	var app models.V3App

	resp, err := c.httpPost("/v3/apps", body)
	if err != nil {
		return app, err
	}

	err = json.Unmarshal(resp, &app)
	if err != nil {
		return app, err
	}

	if app.Error_Code != "" {
		return app, errors.New("Error creating v3 app: " + app.Error_Code)
	}

	return app, nil
}

func (c Connection) createDockerPackage(app models.V3App, dockerImage string) (models.V3Package, error) {
	var pack models.V3Package

	body := fmt.Sprintf(`{"type": "docker", "data": {"image": %s}}`, dockerImage)
	resp, err := c.httpPost(fmt.Sprintf("/v3/apps/%s/packages", app.Guid), body)
	if err != nil {
		return pack, err
	}

	err = json.Unmarshal(resp, &pack)
	if err != nil {
		return pack, err
	}

	if pack.ErrorCode != "" {
		return pack, errors.New("Error creating v3 app package: " + pack.ErrorCode)
	}

	return pack, nil
}

func (c Connection) createSourcePackage(app models.V3App, appDir string) (models.V3Package, error) {
	var pack models.V3Package

	resp, err := c.httpPost(fmt.Sprintf("/v3/apps/%s/packages", app.Guid), "{\"type\": \"bits\"}")
	if err != nil {
		return pack, nil
	}

	err = json.Unmarshal(resp, &pack)
	if err != nil {
		return pack, nil
	}

	if pack.ErrorCode != "" {
		return pack, errors.New("Error creating v3 app package: " + pack.ErrorCode)
	}

	token, err := c.AccessToken()
	if err != nil {
		return pack, err
	}

	api, err := c.ApiEndpoint()
	if err != nil {
		return pack, err
	}

	if strings.Index(api, "s") == 4 {
		api = api[:4] + api[5:]
	}

	//gather files
	var zipper appfiles.ApplicationZipper
	fileutils.TempFile("uploads", func(zipFile *os.File, err error) {
		zipper.Zip(appDir, zipFile)
		data, err := exec.Command("curl",
			fmt.Sprintf("%s/v3/packages/%s/upload", api, pack.Guid),
			"-F", fmt.Sprintf("bits=@%s", zipFile.Name()),
			"-H", fmt.Sprintf("Authorization: %s", token)).Output()

		fmt.Printf("%s\n", data)
		fmt.Println(err)

		util.ExitIfError(err)

	})

	//waiting for cc to pour bits into blobstore
	util.Poll(c, fmt.Sprintf("/v3/packages/%s", pack.Guid), "READY", time.Minute, "Package failed to upload")

	return pack, nil
}

func (c Connection) createTask(app models.V3App, name, command string) (models.V3Task, error) {
	var task models.V3Task

	body := map[string]string{
		"name":    name,
		"command": command,
	}

	data, err := json.Marshal(&body)
	if err != nil {
		return task, err
	}

	resp, err := c.httpPost(fmt.Sprintf("/v3/apps/%s/tasks", app.Guid), string(data))
	if err != nil {
		return task, err
	}

	err = json.Unmarshal(resp, &task)
	if err != nil {
		return task, err
	}

	if task.Guid == "" {
		return task, fmt.Errorf("Failed to run task %s:\n%s\n", name, resp)
	}

	return task, nil
}

func (c Connection) waitForDroplet(pack models.V3Package) (models.V3Droplet, error) {
	var droplet models.V3Droplet

	resp, err := c.httpPost(fmt.Sprintf("/v3/packages/%s/droplets", pack.Guid), "{}")
	util.ExitIfError(err)

	err = json.Unmarshal(resp, &droplet)
	if err != nil {
		return droplet, err
	}

	//wait for the droplet to be ready
	util.Poll(c, fmt.Sprintf("/v3/droplets/%s", droplet.Guid), "STAGED", time.Minute, "Droplet failed to stage")

	return droplet, nil
}
