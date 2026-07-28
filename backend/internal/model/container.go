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

type ContainerStats struct {
	CPUStats    CPUStats    `json:"cpu_stats"`
	MemoryStats MemoryStats `json:"memory_stats"`
	PreCPUStats CPUStats    `json:"precpu_stats"`
	NumProcs    uint32      `json:"num_procs"`
}

type CPUStats struct {
	CPUUsage       CPUUsage `json:"cpu_usage"`
	SystemCPUUsage uint64   `json:"system_cpu_usage"`
	OnlineCPUs     uint32   `json:"online_cpus"`
}

type CPUUsage struct {
	TotalUsage  uint64 `json:"total_usage"`
	PercpuUsage []uint64 `json:"percpu_usage,omitempty"`
}

type MemoryStats struct {
	Usage    uint64 `json:"usage"`
	Limit    uint64 `json:"limit"`
	MaxUsage uint64 `json:"max_usage,omitempty"`
}
