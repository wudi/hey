package vm

import (
	"strings"
	"sync"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// ClassManager manages class storage and access in the VM
type ClassManager struct {
	// Runtime class information
	ClassTable *sync.Map // map[string]*classRuntime

	// Current class being processed (for class declaration)
	currentClass *classRuntime

	// Mutex for currentClass access
	mu sync.RWMutex
}

// NewClassManager creates a new class manager
func NewClassManager() *ClassManager {
	return &ClassManager{
		ClassTable: &sync.Map{},
	}
}

// GetClass safely retrieves a class runtime
func (cm *ClassManager) GetClass(name string) (*classRuntime, bool) {
	if val, ok := cm.ClassTable.Load(classKey(name)); ok {
		return val.(*classRuntime), true
	}
	return nil, false
}

// EnsureClass ensures a class exists, creating it if necessary
func (cm *ClassManager) EnsureClass(name string, userClasses map[string]*registry.Class) *classRuntime {
	key := classKey(name)

	// Try to load existing class first
	if val, ok := cm.ClassTable.Load(key); ok {
		return val.(*classRuntime)
	}

	// Create new class
	cls := &classRuntime{
		Name:        name,
		Properties:  make(map[string]*propertyRuntime),
		StaticProps: make(map[string]*values.Value),
		Constants:   make(map[string]*values.Value),
	}

	// Get user class definition
	if userClasses != nil {
		if userClassDef, hasUserClass := userClasses[strings.ToLower(name)]; hasUserClass {
			populateRuntimeFromClassDef(cls, userClassDef)
		}
	}

	// Check global registry
	if cls.Descriptor == nil && registry.GlobalRegistry != nil {
		if desc, err := registry.GlobalRegistry.GetClass(name); err == nil {
			if converted := classFromDescriptor(desc); converted != nil {
				populateRuntimeFromClassDef(cls, converted)
			}
		}
	}

	// Check global class cache
	if cls.Descriptor == nil {
		if cached := getGlobalClass(name); cached != nil {
			populateRuntimeFromClassDef(cls, cached)
		}
	}

	// Use LoadOrStore to handle race conditions
	if actual, loaded := cm.ClassTable.LoadOrStore(key, cls); loaded {
		// Another goroutine created it first, use theirs
		cls = actual.(*classRuntime)
	}

	// Handle parent class inheritance (recursive call is safe now)
	if cls.Parent != "" {
		if parent := cm.EnsureClass(cls.Parent, userClasses); parent != nil {
			inheritClassMetadata(cls, parent)
		}
	}

	return cls
}

// SetCurrentClass sets the current class being processed
func (cm *ClassManager) SetCurrentClass(cls *classRuntime) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.currentClass = cls
}

// GetCurrentClass gets the current class being processed
func (cm *ClassManager) GetCurrentClass() *classRuntime {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.currentClass
}

// ClearCurrentClass clears the current class
func (cm *ClassManager) ClearCurrentClass() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.currentClass = nil
}

// StoreClass stores a class runtime in the table
func (cm *ClassManager) StoreClass(name string, cls *classRuntime) {
	cm.ClassTable.Store(classKey(name), cls)
}

// Copy creates a deep copy of the class manager state
func (cm *ClassManager) Copy() *ClassManager {
	copy := NewClassManager()

	// Copy class table
	cm.ClassTable.Range(func(key, value interface{}) bool {
		copy.ClassTable.Store(key, copyClassRuntime(value.(*classRuntime)))
		return true
	})

	// Copy current class
	cm.mu.RLock()
	if cm.currentClass != nil {
		copy.currentClass = copyClassRuntime(cm.currentClass)
	}
	cm.mu.RUnlock()

	return copy
}

// Clear resets all class storage (useful for testing)
func (cm *ClassManager) Clear() {
	cm.ClassTable = &sync.Map{}
	cm.ClearCurrentClass()
}