package model

import "time"

type Container struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	State   string            `json:"state"`
	Status  string            `json:"status"`
	Created time.Time         `json:"created"`
	Ports   []ContainerPort   `json:"ports,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}

type ContainerPort struct {
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port,omitempty"`
	Type        string `json:"type"`
	IP          string `json:"ip,omitempty"`
}

type ContainerAction struct {
	Action string `json:"action" binding:"required,oneof=start stop restart remove"`
}
