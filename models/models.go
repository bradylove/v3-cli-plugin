package models

import "time"

type V3App struct {
	Name       string
	Guid       string
	Error_Code string
	Processes  string
	Instances  int `json:"total_desired_instances"`
}

type V3Process struct {
	Type      string
	Instances int
	Memory    int        `json:"memory_in_mb"`
	Disk      int        `json:"disk_in_mb"`
	Links     Links `json:"links"`
}

type V3Task struct {
	Name      string    `json:"name"`
	Guid      string    `json:"guid"`
	Command   string    `json:"command"`
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
}

type V3Apps struct {
	Apps []V3App `json:"resources"`
}

type V3Processes struct {
	Processes []V3Process `json:"resources"`
}

type V3Tasks struct {
	Tasks []V3Task `json:"resources"`
}

type Link struct {
	Href string
}

type Links struct {
	App   Link
	Space Link
}

type V3Package struct {
	Guid      string
	ErrorCode string
}

type V3Droplet struct {
	Guid string
}

type Metadata struct {
	Guid string `json:"guid"`
}

type Entity struct {
	Name string `json:"name"`
}
type RouteEntity struct {
	Host string `json:"host"`
}

type Domains struct {
	NextUrl   string        `json:"next_url,omitempty"`
	Resources []Domain `json:"resources"`
}
type Domain struct {
	Metadata Metadata `json:"metadata"`
	Entity   Entity   `json:"entity"`
}
type Route struct {
	Metadata Metadata    `json:"metadata"`
	Entity   RouteEntity `json:"entity"`
}
type RoutesModel struct {
	Routes []Route `json:"resources"`
}
