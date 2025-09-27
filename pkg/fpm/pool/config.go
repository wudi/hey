package pool

import "time"

type ProcessManagement string

const (
	PMStatic   ProcessManagement = "static"
	PMDynamic  ProcessManagement = "dynamic"
	PMOndemand ProcessManagement = "ondemand"
)

type PoolConfig struct {
	Name              string
	ProcessManagement ProcessManagement

	MaxChildren      int
	StartServers     int
	MinSpareServers  int
	MaxSpareServers  int
	MaxRequests      int
	ProcessIdleTimeout time.Duration

	RequestTerminateTimeout time.Duration
}

func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		Name:              "www",
		ProcessManagement: PMDynamic,
		MaxChildren:       50,
		StartServers:      5,
		MinSpareServers:   5,
		MaxSpareServers:   35,
		MaxRequests:       500,
		ProcessIdleTimeout: 10 * time.Second,
		RequestTerminateTimeout: 30 * time.Second,
	}
}