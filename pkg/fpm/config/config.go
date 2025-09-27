package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wudi/hey/pkg/fpm/pool"
)

type Config struct {
	Global GlobalConfig
	Pools  []PoolConfig
}

type GlobalConfig struct {
	PIDFile      string
	ErrorLog     string
	LogLevel     string
	EmergencyRestartThreshold int
	EmergencyRestartInterval  time.Duration
}

type PoolConfig struct {
	Name              string
	Listen            string
	ListenBacklog     int
	ListenOwner       string
	ListenGroup       string
	ListenMode        string

	ProcessManagement pool.ProcessManagement

	MaxChildren      int
	StartServers     int
	MinSpareServers  int
	MaxSpareServers  int
	MaxRequests      int
	ProcessIdleTimeout time.Duration

	RequestTerminateTimeout time.Duration
	SlowLogFile             string
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{
		Global: GlobalConfig{
			PIDFile:  "/var/run/hey-fpm.pid",
			ErrorLog: "/var/log/hey-fpm.log",
			LogLevel: "notice",
			EmergencyRestartThreshold: 10,
			EmergencyRestartInterval:  1 * time.Minute,
		},
		Pools: []PoolConfig{},
	}

	scanner := bufio.NewScanner(file)
	var currentSection string
	var currentPool *PoolConfig

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			if currentSection != "global" {
				currentPool = &PoolConfig{
					Name:              currentSection,
					Listen:            "127.0.0.1:9000",
					ListenBacklog:     511,
					ProcessManagement: pool.PMDynamic,
					MaxChildren:       50,
					StartServers:      5,
					MinSpareServers:   5,
					MaxSpareServers:   35,
					MaxRequests:       500,
					ProcessIdleTimeout: 10 * time.Second,
					RequestTerminateTimeout: 30 * time.Second,
				}
				config.Pools = append(config.Pools, *currentPool)
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if currentSection == "global" {
			if err := parseGlobalConfig(&config.Global, key, value); err != nil {
				return nil, fmt.Errorf("error parsing global config: %v", err)
			}
		} else if currentPool != nil {
			if err := parsePoolConfig(currentPool, key, value); err != nil {
				return nil, fmt.Errorf("error parsing pool config: %v", err)
			}
			config.Pools[len(config.Pools)-1] = *currentPool
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func parseGlobalConfig(config *GlobalConfig, key, value string) error {
	switch key {
	case "pid":
		config.PIDFile = value
	case "error_log":
		config.ErrorLog = value
	case "log_level":
		config.LogLevel = value
	case "emergency_restart_threshold":
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		config.EmergencyRestartThreshold = val
	case "emergency_restart_interval":
		dur, err := parseDuration(value)
		if err != nil {
			return err
		}
		config.EmergencyRestartInterval = dur
	}
	return nil
}

func parsePoolConfig(config *PoolConfig, key, value string) error {
	switch key {
	case "listen":
		config.Listen = value
	case "listen.backlog":
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		config.ListenBacklog = val
	case "listen.owner":
		config.ListenOwner = value
	case "listen.group":
		config.ListenGroup = value
	case "listen.mode":
		config.ListenMode = value
	case "pm":
		config.ProcessManagement = pool.ProcessManagement(value)
	case "pm.max_children":
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		config.MaxChildren = val
	case "pm.start_servers":
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		config.StartServers = val
	case "pm.min_spare_servers":
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		config.MinSpareServers = val
	case "pm.max_spare_servers":
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		config.MaxSpareServers = val
	case "pm.max_requests":
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		config.MaxRequests = val
	case "pm.process_idle_timeout":
		dur, err := parseDuration(value)
		if err != nil {
			return err
		}
		config.ProcessIdleTimeout = dur
	case "request_terminate_timeout":
		dur, err := parseDuration(value)
		if err != nil {
			return err
		}
		config.RequestTerminateTimeout = dur
	case "slowlog":
		config.SlowLogFile = value
	}
	return nil
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "s") {
		seconds, err := strconv.Atoi(strings.TrimSuffix(s, "s"))
		if err != nil {
			return 0, err
		}
		return time.Duration(seconds) * time.Second, nil
	}
	if strings.HasSuffix(s, "m") {
		minutes, err := strconv.Atoi(strings.TrimSuffix(s, "m"))
		if err != nil {
			return 0, err
		}
		return time.Duration(minutes) * time.Minute, nil
	}
	return time.ParseDuration(s)
}