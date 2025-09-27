package pool

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/wudi/hey/pkg/fastcgi"
	"github.com/wudi/hey/pkg/fpm/handler"
	"github.com/wudi/hey/vmfactory"
)

type WorkerPool struct {
	config      *PoolConfig
	handler     *handler.RequestHandler
	workers     []*Worker
	workersMu   sync.RWMutex
	nextID      int
	stopChan    chan struct{}
	wg          sync.WaitGroup
	stats       *PoolStats
}

type PoolStats struct {
	mu                sync.RWMutex
	AcceptedConn      uint64
	SlowRequests      uint64
	ListenQueue       int
	MaxListenQueue    int
	ActiveProcesses   int
	IdleProcesses     int
	TotalProcesses    int
	StartTime         time.Time
}

func NewWorkerPool(config *PoolConfig, vmFactory *vmfactory.VMFactory) *WorkerPool {
	return &WorkerPool{
		config:   config,
		handler:  handler.NewRequestHandler(vmFactory),
		workers:  make([]*Worker, 0),
		stopChan: make(chan struct{}),
		stats: &PoolStats{
			StartTime: time.Now(),
		},
	}
}

func (p *WorkerPool) Start() error {
	log.Printf("Starting worker pool '%s' with %s process management", p.config.Name, p.config.ProcessManagement)

	switch p.config.ProcessManagement {
	case PMStatic:
		return p.startStatic()
	case PMDynamic:
		return p.startDynamic()
	case PMOndemand:
		return p.startOndemand()
	default:
		return fmt.Errorf("unknown process management mode: %s", p.config.ProcessManagement)
	}
}

func (p *WorkerPool) startStatic() error {
	for i := 0; i < p.config.MaxChildren; i++ {
		p.spawnWorker()
	}
	return nil
}

func (p *WorkerPool) startDynamic() error {
	for i := 0; i < p.config.StartServers; i++ {
		p.spawnWorker()
	}

	p.wg.Add(1)
	go p.dynamicScaler()

	return nil
}

func (p *WorkerPool) startOndemand() error {
	p.wg.Add(1)
	go p.ondemandManager()

	return nil
}

func (p *WorkerPool) spawnWorker() *Worker {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	worker := NewWorker(p.nextID, p.handler, p.config)
	p.nextID++
	p.workers = append(p.workers, worker)
	worker.Start()

	p.updateStats()

	log.Printf("Spawned worker %d (total: %d)", worker.id, len(p.workers))
	return worker
}

func (p *WorkerPool) killWorker(worker *Worker) {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	for i, w := range p.workers {
		if w == worker {
			w.Stop()
			p.workers = append(p.workers[:i], p.workers[i+1:]...)
			p.updateStats()
			log.Printf("Killed worker %d (total: %d)", w.id, len(p.workers))
			return
		}
	}
}

func (p *WorkerPool) dynamicScaler() {
	defer p.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.scaleWorkers()
		}
	}
}

func (p *WorkerPool) scaleWorkers() {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	idle := 0
	active := 0
	var idleWorkers []*Worker

	for _, w := range p.workers {
		if w.IsIdle() {
			idle++
			idleWorkers = append(idleWorkers, w)
		} else {
			active++
		}
	}

	total := len(p.workers)

	if idle < p.config.MinSpareServers && total < p.config.MaxChildren {
		needed := p.config.MinSpareServers - idle
		for i := 0; i < needed && total < p.config.MaxChildren; i++ {
			worker := NewWorker(p.nextID, p.handler, p.config)
			p.nextID++
			p.workers = append(p.workers, worker)
			worker.Start()
			total++
			log.Printf("Scaled up: spawned worker %d (total: %d)", worker.id, total)
		}
	}

	if idle > p.config.MaxSpareServers {
		excess := idle - p.config.MaxSpareServers
		for i := 0; i < excess && i < len(idleWorkers); i++ {
			w := idleWorkers[i]
			w.Stop()
			for j, worker := range p.workers {
				if worker == w {
					p.workers = append(p.workers[:j], p.workers[j+1:]...)
					break
				}
			}
			log.Printf("Scaled down: killed worker %d (total: %d)", w.id, len(p.workers))
		}
	}

	p.updateStats()
}

func (p *WorkerPool) ondemandManager() {
	defer p.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.cleanupIdleWorkers()
		}
	}
}

func (p *WorkerPool) cleanupIdleWorkers() {
	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	var toKill []*Worker

	for _, w := range p.workers {
		if w.IsIdle() && w.IdleTime() > p.config.ProcessIdleTimeout {
			toKill = append(toKill, w)
		}
	}

	for _, w := range toKill {
		w.Stop()
		for i, worker := range p.workers {
			if worker == w {
				p.workers = append(p.workers[:i], p.workers[i+1:]...)
				break
			}
		}
		log.Printf("Ondemand: killed idle worker %d (idle: %v)", w.id, w.IdleTime())
	}

	if len(toKill) > 0 {
		p.updateStats()
	}
}

func (p *WorkerPool) HandleRequest(proto *fastcgi.Protocol, req *fastcgi.Request) error {
	p.stats.mu.Lock()
	p.stats.AcceptedConn++
	p.stats.mu.Unlock()

	worker := p.getAvailableWorker()
	if worker == nil {
		return fmt.Errorf("no available workers")
	}

	if !worker.Submit(proto, req) {
		return fmt.Errorf("failed to submit request to worker")
	}

	return nil
}

func (p *WorkerPool) getAvailableWorker() *Worker {
	p.workersMu.RLock()
	defer p.workersMu.RUnlock()

	for _, w := range p.workers {
		if w.IsIdle() {
			return w
		}
	}

	if p.config.ProcessManagement == PMOndemand && len(p.workers) < p.config.MaxChildren {
		p.workersMu.RUnlock()
		worker := p.spawnWorker()
		p.workersMu.RLock()
		return worker
	}

	return nil
}

func (p *WorkerPool) Stop() {
	log.Printf("Stopping worker pool '%s'", p.config.Name)

	close(p.stopChan)
	p.wg.Wait()

	p.workersMu.Lock()
	defer p.workersMu.Unlock()

	for _, w := range p.workers {
		w.Stop()
	}
	p.workers = nil
}

func (p *WorkerPool) updateStats() {
	idle := 0
	active := 0

	for _, w := range p.workers {
		if w.IsIdle() {
			idle++
		} else {
			active++
		}
	}

	p.stats.mu.Lock()
	p.stats.IdleProcesses = idle
	p.stats.ActiveProcesses = active
	p.stats.TotalProcesses = len(p.workers)
	p.stats.mu.Unlock()
}

func (p *WorkerPool) GetStats() *PoolStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	return &PoolStats{
		AcceptedConn:    p.stats.AcceptedConn,
		SlowRequests:    p.stats.SlowRequests,
		ListenQueue:     p.stats.ListenQueue,
		MaxListenQueue:  p.stats.MaxListenQueue,
		ActiveProcesses: p.stats.ActiveProcesses,
		IdleProcesses:   p.stats.IdleProcesses,
		TotalProcesses:  p.stats.TotalProcesses,
		StartTime:       p.stats.StartTime,
	}
}