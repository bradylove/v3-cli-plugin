package commands

import (
	"strings"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/bradylove/v3-cli-plugin/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudfoundry/gofileutils/fileutils"
	"os"
	"os/exec"
	"time"
	"github.com/bradylove/v3-cli-plugin/util"
	"github.com/cloudfoundry/cli/cf/appfiles"
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