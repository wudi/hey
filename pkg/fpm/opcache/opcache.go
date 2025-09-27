package opcache

import (
	"crypto/sha256"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wudi/hey/registry"
)

type CompiledScript struct {
	Bytecode   []byte
	Constants  []*registry.Constant
	Functions  map[string]*registry.Function
	Classes    map[string]*registry.Class
	Interfaces map[string]*registry.Interface
	Traits     map[string]*registry.Trait
	Timestamp  time.Time
	FileHash   string
}

type OpcacheManager struct {
	cache      map[string]*CompiledScript
	mu         sync.RWMutex
	enabled    bool
	maxEntries int
	validateTimestamps bool
}

type OpcacheConfig struct {
	Enabled            bool
	MaxEntries         int
	ValidateTimestamps bool
}

func NewOpcacheManager(config *OpcacheConfig) *OpcacheManager {
	return &OpcacheManager{
		cache:              make(map[string]*CompiledScript),
		enabled:            config.Enabled,
		maxEntries:         config.MaxEntries,
		validateTimestamps: config.ValidateTimestamps,
	}
}

func (o *OpcacheManager) Get(file string) (*CompiledScript, bool) {
	if !o.enabled {
		return nil, false
	}

	o.mu.RLock()
	defer o.mu.RUnlock()

	cached, ok := o.cache[file]
	if !ok {
		return nil, false
	}

	if o.validateTimestamps {
		stat, err := os.Stat(file)
		if err != nil {
			delete(o.cache, file)
			return nil, false
		}

		if stat.ModTime().After(cached.Timestamp) {
			delete(o.cache, file)
			return nil, false
		}

		hash, err := computeFileHash(file)
		if err != nil || hash != cached.FileHash {
			delete(o.cache, file)
			return nil, false
		}
	}

	return cached, true
}

func (o *OpcacheManager) Set(file string, compiled *CompiledScript) {
	if !o.enabled {
		return
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if len(o.cache) >= o.maxEntries {
		o.evictOldest()
	}

	stat, err := os.Stat(file)
	if err != nil {
		return
	}

	hash, err := computeFileHash(file)
	if err != nil {
		return
	}

	compiled.Timestamp = stat.ModTime()
	compiled.FileHash = hash

	o.cache[file] = compiled
}

func (o *OpcacheManager) Invalidate(file string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	delete(o.cache, file)
}

func (o *OpcacheManager) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.cache = make(map[string]*CompiledScript)
}

func (o *OpcacheManager) Stats() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return map[string]interface{}{
		"enabled":             o.enabled,
		"cached_scripts":      len(o.cache),
		"max_cached_scripts":  o.maxEntries,
		"validate_timestamps": o.validateTimestamps,
	}
}

func (o *OpcacheManager) evictOldest() {
	var oldestFile string
	var oldestTime time.Time

	for file, compiled := range o.cache {
		if oldestFile == "" || compiled.Timestamp.Before(oldestTime) {
			oldestFile = file
			oldestTime = compiled.Timestamp
		}
	}

	if oldestFile != "" {
		delete(o.cache, oldestFile)
	}
}

func computeFileHash(file string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

var GlobalOpcache *OpcacheManager

func InitGlobalOpcache(config *OpcacheConfig) {
	GlobalOpcache = NewOpcacheManager(config)
}