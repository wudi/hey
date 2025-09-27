package pool

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/wudi/hey/pkg/fastcgi"
	"github.com/wudi/hey/pkg/fpm/handler"
)

type WorkerState int32

const (
	WorkerIdle WorkerState = iota
	WorkerBusy
	WorkerStopping
)

type Worker struct {
	id            int
	handler       *handler.RequestHandler
	config        *PoolConfig
	requestCount  uint64
	state         atomic.Int32
	lastUsed      time.Time
	stopChan      chan struct{}
	requestChan   chan *workerRequest
}

type workerRequest struct {
	proto *fastcgi.Protocol
	req   *fastcgi.Request
}

func NewWorker(id int, handler *handler.RequestHandler, config *PoolConfig) *Worker {
	w := &Worker{
		id:          id,
		handler:     handler,
		config:      config,
		lastUsed:    time.Now(),
		stopChan:    make(chan struct{}),
		requestChan: make(chan *workerRequest, 1),
	}
	w.state.Store(int32(WorkerIdle))
	return w
}

func (w *Worker) Start() {
	go w.run()
}

func (w *Worker) run() {
	log.Printf("Worker %d started", w.id)

	for {
		select {
		case <-w.stopChan:
			log.Printf("Worker %d stopping", w.id)
			return

		case req := <-w.requestChan:
			w.handleRequest(req)
		}
	}
}

func (w *Worker) handleRequest(req *workerRequest) {
	w.setState(WorkerBusy)
	defer w.setState(WorkerIdle)

	w.lastUsed = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), w.config.RequestTerminateTimeout)
	defer cancel()

	if err := w.handler.HandleRequest(ctx, req.proto, req.req); err != nil {
		log.Printf("Worker %d: error handling request: %v", w.id, err)
	}

	atomic.AddUint64(&w.requestCount, 1)

	if w.config.MaxRequests > 0 && atomic.LoadUint64(&w.requestCount) >= uint64(w.config.MaxRequests) {
		log.Printf("Worker %d reached max requests (%d), will restart", w.id, w.config.MaxRequests)
		w.Stop()
	}
}

func (w *Worker) Submit(proto *fastcgi.Protocol, req *fastcgi.Request) bool {
	select {
	case w.requestChan <- &workerRequest{proto: proto, req: req}:
		return true
	default:
		return false
	}
}

func (w *Worker) Stop() {
	if w.state.Load() == int32(WorkerStopping) {
		return
	}
	w.setState(WorkerStopping)
	close(w.stopChan)
}

func (w *Worker) GetState() WorkerState {
	return WorkerState(w.state.Load())
}

func (w *Worker) setState(state WorkerState) {
	w.state.Store(int32(state))
}

func (w *Worker) IsIdle() bool {
	return w.GetState() == WorkerIdle
}

func (w *Worker) IdleTime() time.Duration {
	if !w.IsIdle() {
		return 0
	}
	return time.Since(w.lastUsed)
}

func (w *Worker) RequestCount() uint64 {
	return atomic.LoadUint64(&w.requestCount)
}