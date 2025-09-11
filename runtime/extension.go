package runtime

import (
	"fmt"
	"sort"

	"github.com/wudi/hey/values"
)

// Extension interface for external extensions
type Extension interface {
	GetName() string
	GetVersion() string
	GetDescription() string
	GetDependencies() []string
	GetLoadOrder() int
	Register(registry *RuntimeRegistry) error
	Unregister(registry *RuntimeRegistry) error
}

// BaseExtension provides a foundation for building extensions
type BaseExtension struct {
	name         string
	version      string
	description  string
	dependencies []string
	loadOrder    int
	registered   bool

	// Collected entities during registration
	registeredConstants []string
	registeredVariables []string
	registeredFunctions []string
	registeredClasses   []string
}

// NewBaseExtension creates a new base extension
func NewBaseExtension(name, version, description string) *BaseExtension {
	return &BaseExtension{
		name:                name,
		version:             version,
		description:         description,
		loadOrder:           100, // Default load order
		registeredConstants: make([]string, 0),
		registeredVariables: make([]string, 0),
		registeredFunctions: make([]string, 0),
		registeredClasses:   make([]string, 0),
	}
}

// Interface implementation
func (be *BaseExtension) GetName() string           { return be.name }
func (be *BaseExtension) GetVersion() string        { return be.version }
func (be *BaseExtension) GetDescription() string    { return be.description }
func (be *BaseExtension) GetDependencies() []string { return be.dependencies }
func (be *BaseExtension) GetLoadOrder() int         { return be.loadOrder }

// SetLoadOrder sets the extension load order (lower numbers load first)
func (be *BaseExtension) SetLoadOrder(order int) {
	be.loadOrder = order
}

// SetDependencies sets the extension dependencies
func (be *BaseExtension) SetDependencies(deps []string) {
	be.dependencies = deps
}

// Helper methods for extensions
func (be *BaseExtension) RegisterConstant(registry *RuntimeRegistry, name string, value *values.Value) error {
	if err := registry.RegisterConstant(name, value, false, be.name); err != nil {
		return err
	}
	be.registeredConstants = append(be.registeredConstants, name)
	return nil
}

func (be *BaseExtension) RegisterFunction(registry *RuntimeRegistry, descriptor *FunctionDescriptor) error {
	descriptor.ExtensionName = be.name
	descriptor.IsBuiltin = false

	if err := registry.RegisterFunction(descriptor); err != nil {
		return err
	}
	be.registeredFunctions = append(be.registeredFunctions, descriptor.Name)
	return nil
}

func (be *BaseExtension) RegisterClass(registry *RuntimeRegistry, descriptor *ClassDescriptor) error {
	descriptor.ExtensionName = be.name
	descriptor.IsBuiltin = false

	if err := registry.RegisterClass(descriptor); err != nil {
		return err
	}
	be.registeredClasses = append(be.registeredClasses, descriptor.Name)
	return nil
}

// Default implementations that can be overridden
func (be *BaseExtension) Register(registry *RuntimeRegistry) error {
	be.registered = true
	return nil
}

func (be *BaseExtension) Unregister(registry *RuntimeRegistry) error {
	if !be.registered {
		return nil
	}

	// Remove all registered entities
	for _, name := range be.registeredConstants {
		delete(registry.constants, name)
	}

	for _, name := range be.registeredVariables {
		delete(registry.variables, name)
	}

	for _, name := range be.registeredFunctions {
		delete(registry.functions, name)
	}

	for _, name := range be.registeredClasses {
		delete(registry.classes, name)
	}

	// Clear tracking
	be.registeredConstants = be.registeredConstants[:0]
	be.registeredVariables = be.registeredVariables[:0]
	be.registeredFunctions = be.registeredFunctions[:0]
	be.registeredClasses = be.registeredClasses[:0]

	be.registered = false
	return nil
}

// ExtensionManager manages extension loading and dependency resolution
type ExtensionManager struct {
	registry   *RuntimeRegistry
	extensions map[string]Extension
	loadOrder  []Extension
}

// NewExtensionManager creates a new extension manager
func NewExtensionManager(registry *RuntimeRegistry) *ExtensionManager {
	return &ExtensionManager{
		registry:   registry,
		extensions: make(map[string]Extension),
		loadOrder:  make([]Extension, 0),
	}
}

// RegisterExtension registers an extension
func (em *ExtensionManager) RegisterExtension(ext Extension) error {
	name := ext.GetName()

	// Check if already registered
	if _, exists := em.extensions[name]; exists {
		return fmt.Errorf("extension already registered: %s", name)
	}

	// Validate dependencies
	if err := em.validateDependencies(ext); err != nil {
		return fmt.Errorf("dependency validation failed for %s: %v", name, err)
	}

	// Add to registry
	em.extensions[name] = ext

	// Register extension descriptor
	descriptor := &ExtensionDescriptor{
		Name:         ext.GetName(),
		Version:      ext.GetVersion(),
		Description:  ext.GetDescription(),
		LoadOrder:    ext.GetLoadOrder(),
		Dependencies: ext.GetDependencies(),
	}

	if err := em.registry.RegisterExtension(descriptor); err != nil {
		delete(em.extensions, name)
		return err
	}

	// Rebuild load order
	em.rebuildLoadOrder()

	return nil
}

// LoadExtension loads a registered extension
func (em *ExtensionManager) LoadExtension(name string) error {
	ext, exists := em.extensions[name]
	if !exists {
		return fmt.Errorf("extension not registered: %s", name)
	}

	// Load dependencies first
	for _, dep := range ext.GetDependencies() {
		if err := em.LoadExtension(dep); err != nil {
			return fmt.Errorf("failed to load dependency %s for %s: %v", dep, name, err)
		}
	}

	// Register with runtime
	return ext.Register(em.registry)
}

// UnloadExtension unloads an extension
func (em *ExtensionManager) UnloadExtension(name string) error {
	ext, exists := em.extensions[name]
	if !exists {
		return fmt.Errorf("extension not registered: %s", name)
	}

	// Check for dependents
	for _, other := range em.extensions {
		for _, dep := range other.GetDependencies() {
			if dep == name {
				return fmt.Errorf("cannot unload %s: extension %s depends on it", name, other.GetName())
			}
		}
	}

	return ext.Unregister(em.registry)
}

// LoadAllExtensions loads all registered extensions in dependency order
func (em *ExtensionManager) LoadAllExtensions() error {
	for _, ext := range em.loadOrder {
		if err := ext.Register(em.registry); err != nil {
			return fmt.Errorf("failed to load extension %s: %v", ext.GetName(), err)
		}
	}
	return nil
}

// validateDependencies validates extension dependencies
func (em *ExtensionManager) validateDependencies(ext Extension) error {
	for _, dep := range ext.GetDependencies() {
		if _, exists := em.extensions[dep]; !exists {
			return fmt.Errorf("missing dependency: %s", dep)
		}
	}

	// Check for circular dependencies
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	return em.checkCircularDependencies(ext.GetName(), ext, visited, recursionStack)
}

// checkCircularDependencies checks for circular dependencies
func (em *ExtensionManager) checkCircularDependencies(name string, ext Extension, visited, recursionStack map[string]bool) error {
	visited[name] = true
	recursionStack[name] = true

	for _, dep := range ext.GetDependencies() {
		if !visited[dep] {
			if depExt, exists := em.extensions[dep]; exists {
				if err := em.checkCircularDependencies(dep, depExt, visited, recursionStack); err != nil {
					return err
				}
			}
		} else if recursionStack[dep] {
			return fmt.Errorf("circular dependency detected: %s -> %s", name, dep)
		}
	}

	recursionStack[name] = false
	return nil
}

// rebuildLoadOrder rebuilds the extension load order based on dependencies
func (em *ExtensionManager) rebuildLoadOrder() {
	extensions := make([]Extension, 0, len(em.extensions))
	for _, ext := range em.extensions {
		extensions = append(extensions, ext)
	}

	// Sort by load order, then by dependency topology
	sort.Slice(extensions, func(i, j int) bool {
		orderI := extensions[i].GetLoadOrder()
		orderJ := extensions[j].GetLoadOrder()

		if orderI != orderJ {
			return orderI < orderJ
		}

		// If same load order, sort by dependencies
		return em.hasDependency(extensions[j], extensions[i])
	})

	em.loadOrder = extensions
}

// hasDependency checks if ext1 depends on ext2
func (em *ExtensionManager) hasDependency(ext1, ext2 Extension) bool {
	for _, dep := range ext1.GetDependencies() {
		if dep == ext2.GetName() {
			return true
		}
		if depExt, exists := em.extensions[dep]; exists {
			if em.hasDependency(depExt, ext2) {
				return true
			}
		}
	}
	return false
}

// GetExtensionNames returns all registered extension names
func (em *ExtensionManager) GetExtensionNames() []string {
	names := make([]string, 0, len(em.extensions))
	for name := range em.extensions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetExtension returns a registered extension
func (em *ExtensionManager) GetExtension(name string) (Extension, bool) {
	ext, exists := em.extensions[name]
	return ext, exists
}

// IsExtensionLoaded checks if an extension is registered
func (em *ExtensionManager) IsExtensionLoaded(name string) bool {
	_, exists := em.extensions[name]
	return exists
}
