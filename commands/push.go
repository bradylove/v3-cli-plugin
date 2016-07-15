package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bradylove/v3-cli-plugin/models"
	"github.com/bradylove/v3-cli-plugin/util"
	"github.com/simonleung8/flags"
)

func Push(conn Connection, args []string) {
	appDir := "."
	buildpack := "null"
	dockerImage := ""

	fc := flags.New()
	fc.NewStringFlag("filepath", "p", "path to app dir or zip to upload")
	fc.NewStringFlag("buildpack", "b", "the buildpack to use")
	fc.NewStringFlag("docker-image", "di", "the docker image to use")
	fc.Parse(args...)
	if fc.IsSet("p") {
		appDir = fc.String("p")
	}
	if fc.IsSet("b") {
		buildpack = fmt.Sprintf(`"%s"`, fc.String("b"))
	}
	if fc.IsSet("di") {
		dockerImage = fmt.Sprintf(`"%s"`, fc.String("di"))
	}

	mySpace, err := conn.GetCurrentSpace()
	util.ExitIfError(err)

	var lifecycle string
	if dockerImage != "" {
		lifecycle = `"lifecycle": { "type": "docker", "data": {} }`
	} else {
		lifecycle = fmt.Sprintf(`"lifecycle": { "type": "buildpack", "data": { "buildpack": %s } }`, buildpack)
	}

	app, err := conn.createV3App(fmt.Sprintf(`{"name":"%s", "relationships": { "space": {"guid":"%s"}}, %s}`, fc.Args()[1], mySpace.Guid, lifecycle))
	util.ExitIfError(err)

	// go Logs(cliConnection, args)
	// time.Sleep(2 * time.Second) // b/c sharing the cliConnection makes things break

	//create package
	var pack models.V3Package
	if dockerImage != "" {
		pack, err = conn.createDockerPackage(app, dockerImage)
		util.ExitIfError(err)
	} else {
		pack, err = conn.createSourcePackage(app, appDir)
		util.ExitIfError(err)
	}

	droplet, err := conn.waitForDroplet(pack)
	util.ExitIfError(err)

	//assign droplet to the app
	output, err := conn.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/droplets/current", app.Guid), "-X", "PUT", "-d", fmt.Sprintf("{\"droplet_guid\":\"%s\"}", droplet.Guid))
	util.ExitIfError(err)

	//pick the first available shared domain, get the guid
	space, _ := conn.GetCurrentSpace()
	nextUrl := "/v2/shared_domains"
	var allDomains models.Domains
	for nextUrl != "" {
		output, err = conn.CliCommandWithoutTerminalOutput("curl", nextUrl)
		util.ExitIfError(err)
		var tmp models.Domains
		err = json.Unmarshal([]byte(strings.Join(output, "")), &tmp)
		util.ExitIfError(err)
		allDomains.Resources = append(allDomains.Resources, tmp.Resources...)

		if tmp.NextUrl != "" {
			nextUrl = tmp.NextUrl
		} else {
			nextUrl = ""
		}
	}
	domainGuid := allDomains.Resources[0].Metadata.Guid
	output, err = conn.CliCommandWithoutTerminalOutput("curl", "v2/routes", "-X", "POST", "-d", fmt.Sprintf(`{"host":"%s","domain_guid":"%s","space_guid":"%s"}`, fc.Args()[1], domainGuid, space.Guid))

	var routeGuid string
	if strings.Contains(output[0], "CF-RouteHostTaken") {
		output, err = conn.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v2/routes?q=host:%s;domain_guid:%s", fc.Args()[1], domainGuid))
		var routes models.RoutesModel
		err = json.Unmarshal([]byte(strings.Join(output, "")), &routes)
		routeGuid = routes.Routes[0].Metadata.Guid
	} else {
		var route models.Route
		err = json.Unmarshal([]byte(strings.Join(output, "")), &route)
		if err != nil {
			util.ExitIfError(errors.New("error unmarshaling the route: " + err.Error()))
		}
		routeGuid = route.Metadata.Guid
	}

	util.ExitIfError(err)
	var route models.Route

	err = json.Unmarshal([]byte(strings.Join(output, "")), &route)
	if err != nil {
		util.ExitIfError(errors.New("error unmarshaling the route: " + err.Error()))
	}

	//map the route to the app
	output, err = conn.CliCommandWithoutTerminalOutput("curl", "/v3/route_mappings", "-X", "POST", "-d", fmt.Sprintf(`{"relationships": { "route": { "guid": "%s" }, "app": { "guid": "%s" } }`, routeGuid, app.Guid))
	util.ExitIfError(err)

	//start the app
	output, err = conn.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("/v3/apps/%s/start", app.Guid), "-X", "PUT")
	util.ExitIfError(err)

	fmt.Println("Done pushing! Checkout your processes using 'cf apps'")
}
