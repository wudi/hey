package compiler

import (
	"fmt"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/runtime"
	"github.com/wudi/hey/stdlib"
)

// InitializeUnifiedSystem initializes the unified class registry system
func InitializeUnifiedSystem() error {
	// 1. Initialize the unified registry
	registry.Initialize()

	// 2. Initialize runtime with unified system
	if err := runtime.UnifiedBootstrap(); err != nil {
		return fmt.Errorf("failed to initialize unified runtime: %v", err)
	}

	// 3. Initialize stdlib classes with unified system
	if err := stdlib.InitializeUnifiedClasses(); err != nil {
		return fmt.Errorf("failed to initialize unified stdlib classes: %v", err)
	}

	return nil
}

// InitializeLegacySystem initializes the legacy system for backward compatibility
func InitializeLegacySystem() error {
	// Initialize legacy runtime
	if err := runtime.Bootstrap(); err != nil {
		return fmt.Errorf("failed to initialize legacy runtime: %v", err)
	}

	return nil
}

// InitializeBothSystems initializes both unified and legacy systems for compatibility
func InitializeBothSystems() error {
	// Initialize unified system first
	if err := InitializeUnifiedSystem(); err != nil {
		return err
	}

	// Initialize legacy system for backward compatibility
	// This will also call UnifiedBootstrap() internally
	return nil
}
