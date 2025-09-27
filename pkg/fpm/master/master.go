package master

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/wudi/hey/compiler"
	"github.com/wudi/hey/pkg/fastcgi"
	"github.com/wudi/hey/pkg/fpm/pool"
	"github.com/wudi/hey/runtime"
	"github.com/wudi/hey/vmfactory"
)

type MasterConfig struct {
	Listen    string
	PIDFile   string
	ErrorLog  string
	LogLevel  string
	PoolConfig *pool.PoolConfig
}

type Master struct {
	config       *MasterConfig
	pool         *pool.WorkerPool
	listener     net.Listener
	sigChan      chan os.Signal
	stopChan     chan struct{}
	wg           sync.WaitGroup
	shutdownOnce sync.Once
}

func NewMaster(config *MasterConfig) *Master {
	return &Master{
		config:   config,
		sigChan:  make(chan os.Signal, 1),
		stopChan: make(chan struct{}),
	}
}

func (m *Master) Start() error {
	log.Printf("Starting Hey-FPM master process")

	if err := runtime.Bootstrap(); err != nil {
		return fmt.Errorf("failed to bootstrap runtime: %v", err)
	}

	if err := runtime.InitializeVMIntegration(); err != nil {
		return fmt.Errorf("failed to initialize VM integration: %v", err)
	}

	listener, err := net.Listen("tcp", m.config.Listen)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", m.config.Listen, err)
	}
	m.listener = listener
	log.Printf("Listening on %s", m.config.Listen)

	if m.config.PIDFile != "" {
		if err := m.writePIDFile(); err != nil {
			return fmt.Errorf("failed to write PID file: %v", err)
		}
	}

	vmFactory := vmfactory.NewVMFactory(func() vmfactory.Compiler {
		return compiler.NewCompiler()
	})

	m.pool = pool.NewWorkerPool(m.config.PoolConfig, vmFactory)
	if err := m.pool.Start(); err != nil {
		return fmt.Errorf("failed to start worker pool: %v", err)
	}

	signal.Notify(m.sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGQUIT)

	m.wg.Add(1)
	go m.handleSignals()

	m.wg.Add(1)
	go m.acceptConnections()

	return nil
}

func (m *Master) acceptConnections() {
	defer m.wg.Done()

	for {
		conn, err := m.listener.Accept()
		if err != nil {
			select {
			case <-m.stopChan:
				return
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		go m.handleConnection(conn)
	}
}

func (m *Master) handleConnection(conn net.Conn) {
	defer conn.Close()

	proto := fastcgi.NewProtocol(conn)

	for {
		req, err := proto.ReadRequest()
		if err != nil {
			return
		}

		if err := m.pool.HandleRequest(proto, req); err != nil {
			log.Printf("Error handling request: %v", err)
			proto.SendResponse(req.ID, nil, []byte(err.Error()), 1)
		}
	}
}

func (m *Master) handleSignals() {
	defer m.wg.Done()

	for {
		select {
		case <-m.stopChan:
			return
		case sig := <-m.sigChan:
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
				log.Printf("Received %v, initiating graceful shutdown", sig)
				m.GracefulShutdown()
				return

			case syscall.SIGUSR1:
				log.Printf("Received SIGUSR1, reopening log files")
				m.reopenLogs()

			case syscall.SIGUSR2:
				log.Printf("Received SIGUSR2, reloading configuration")
				m.reloadConfig()
			}
		}
	}
}

func (m *Master) GracefulShutdown() {
	m.shutdownOnce.Do(func() {
		log.Printf("Graceful shutdown initiated")

		close(m.stopChan)

		if m.listener != nil {
			m.listener.Close()
		}

		if m.pool != nil {
			m.pool.Stop()
		}

		if m.config.PIDFile != "" {
			os.Remove(m.config.PIDFile)
		}
	})
}

func (m *Master) Wait() {
	m.wg.Wait()
	log.Printf("Master process shutdown complete")
}

func (m *Master) reopenLogs() {
	log.Printf("Log rotation not yet implemented")
}

func (m *Master) reloadConfig() {
	log.Printf("Configuration reload not yet implemented")
}

func (m *Master) writePIDFile() error {
	pid := os.Getpid()
	return os.WriteFile(m.config.PIDFile, []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

func (m *Master) GetStats() *pool.PoolStats {
	if m.pool == nil {
		return nil
	}
	return m.pool.GetStats()
}