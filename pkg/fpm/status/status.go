package status

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/wudi/hey/pkg/fpm/pool"
)

type StatusHandler struct {
	pool *pool.WorkerPool
}

func NewStatusHandler(p *pool.WorkerPool) *StatusHandler {
	return &StatusHandler{pool: p}
}

type Status struct {
	Pool            string    `json:"pool"`
	ProcessManager  string    `json:"process-manager"`
	StartTime       time.Time `json:"start-time"`
	StartSince      int64     `json:"start-since"`
	AcceptedConn    uint64    `json:"accepted-conn"`
	ListenQueue     int       `json:"listen-queue"`
	MaxListenQueue  int       `json:"max-listen-queue"`
	ListenQueueLen  int       `json:"listen-queue-len"`
	IdleProcesses   int       `json:"idle-processes"`
	ActiveProcesses int       `json:"active-processes"`
	TotalProcesses  int       `json:"total-processes"`
	MaxActiveProcesses int    `json:"max-active-processes"`
	MaxChildrenReached int    `json:"max-children-reached"`
	SlowRequests    uint64    `json:"slow-requests"`
}

func (h *StatusHandler) GetStatus() *Status {
	stats := h.pool.GetStats()

	return &Status{
		Pool:               "www",
		ProcessManager:     "dynamic",
		StartTime:          stats.StartTime,
		StartSince:         int64(time.Since(stats.StartTime).Seconds()),
		AcceptedConn:       stats.AcceptedConn,
		ListenQueue:        stats.ListenQueue,
		MaxListenQueue:     stats.MaxListenQueue,
		ListenQueueLen:     0,
		IdleProcesses:      stats.IdleProcesses,
		ActiveProcesses:    stats.ActiveProcesses,
		TotalProcesses:     stats.TotalProcesses,
		MaxActiveProcesses: stats.ActiveProcesses,
		MaxChildrenReached: 0,
		SlowRequests:       stats.SlowRequests,
	}
}

func (h *StatusHandler) GetStatusJSON() ([]byte, error) {
	status := h.GetStatus()
	return json.MarshalIndent(status, "", "  ")
}

func (h *StatusHandler) GetStatusText() string {
	status := h.GetStatus()
	return fmt.Sprintf(`pool:                 %s
process manager:      %s
start time:           %s
start since:          %d
accepted conn:        %d
listen queue:         %d
max listen queue:     %d
listen queue len:     %d
idle processes:       %d
active processes:     %d
total processes:      %d
max active processes: %d
max children reached: %d
slow requests:        %d`,
		status.Pool,
		status.ProcessManager,
		status.StartTime.Format(time.RFC3339),
		status.StartSince,
		status.AcceptedConn,
		status.ListenQueue,
		status.MaxListenQueue,
		status.ListenQueueLen,
		status.IdleProcesses,
		status.ActiveProcesses,
		status.TotalProcesses,
		status.MaxActiveProcesses,
		status.MaxChildrenReached,
		status.SlowRequests,
	)
}