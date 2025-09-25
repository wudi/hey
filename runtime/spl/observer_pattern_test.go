package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
)

func TestObserverPattern(t *testing.T) {
	registry.Initialize()

	// Manually register the SPL interfaces for testing
	for _, iface := range GetSplInterfaces() {
		err := registry.GlobalRegistry.RegisterInterface(iface)
		if err != nil {
			t.Fatalf("Failed to register interface %s: %v", iface.Name, err)
		}
	}

	t.Run("SplObserver Interface Registration", func(t *testing.T) {
		// Test that SplObserver interface is registered
		iface, exists := registry.GlobalRegistry.GetInterface("SplObserver")
		if !exists || iface == nil {
			t.Fatal("SplObserver interface not registered")
		}

		if iface.Name != "SplObserver" {
			t.Errorf("Expected interface name 'SplObserver', got '%s'", iface.Name)
		}

		// Check update method
		updateMethod, exists := iface.Methods["update"]
		if !exists {
			t.Fatal("update method not found in SplObserver interface")
		}

		if updateMethod.Name != "update" {
			t.Errorf("Expected method name 'update', got '%s'", updateMethod.Name)
		}

		if len(updateMethod.Parameters) != 1 {
			t.Errorf("Expected 1 parameter for update method, got %d", len(updateMethod.Parameters))
		}

		if updateMethod.Parameters[0].Name != "subject" {
			t.Errorf("Expected parameter name 'subject', got '%s'", updateMethod.Parameters[0].Name)
		}

		if updateMethod.ReturnType != "void" {
			t.Errorf("Expected return type 'void', got '%s'", updateMethod.ReturnType)
		}
	})

	t.Run("SplSubject Interface Registration", func(t *testing.T) {
		// Test that SplSubject interface is registered
		iface, exists := registry.GlobalRegistry.GetInterface("SplSubject")
		if !exists || iface == nil {
			t.Fatal("SplSubject interface not registered")
		}

		if iface.Name != "SplSubject" {
			t.Errorf("Expected interface name 'SplSubject', got '%s'", iface.Name)
		}

		// Check attach method
		attachMethod, exists := iface.Methods["attach"]
		if !exists {
			t.Fatal("attach method not found in SplSubject interface")
		}

		if attachMethod.Name != "attach" {
			t.Errorf("Expected method name 'attach', got '%s'", attachMethod.Name)
		}

		if len(attachMethod.Parameters) != 1 {
			t.Errorf("Expected 1 parameter for attach method, got %d", len(attachMethod.Parameters))
		}

		if attachMethod.Parameters[0].Type != "SplObserver" {
			t.Errorf("Expected parameter type 'SplObserver', got '%s'", attachMethod.Parameters[0].Type)
		}

		// Check detach method
		detachMethod, exists := iface.Methods["detach"]
		if !exists {
			t.Fatal("detach method not found in SplSubject interface")
		}

		if detachMethod.Name != "detach" {
			t.Errorf("Expected method name 'detach', got '%s'", detachMethod.Name)
		}

		// Check notify method
		notifyMethod, exists := iface.Methods["notify"]
		if !exists {
			t.Fatal("notify method not found in SplSubject interface")
		}

		if notifyMethod.Name != "notify" {
			t.Errorf("Expected method name 'notify', got '%s'", notifyMethod.Name)
		}

		if len(notifyMethod.Parameters) != 0 {
			t.Errorf("Expected 0 parameters for notify method, got %d", len(notifyMethod.Parameters))
		}

		if notifyMethod.ReturnType != "void" {
			t.Errorf("Expected return type 'void', got '%s'", notifyMethod.ReturnType)
		}
	})

	t.Run("Interface Methods Verification", func(t *testing.T) {
		// Verify both interfaces exist in global registry
		observerIface, observerExists := registry.GlobalRegistry.GetInterface("SplObserver")
		subjectIface, subjectExists := registry.GlobalRegistry.GetInterface("SplSubject")

		if !observerExists || !subjectExists || observerIface == nil || subjectIface == nil {
			t.Fatal("Observer or Subject interface not found in global registry")
		}

		// The interfaces should be properly structured for use in Hey-Codex VM
		if len(observerIface.Methods) != 1 {
			t.Errorf("Expected 1 method in SplObserver, got %d", len(observerIface.Methods))
		}

		if len(subjectIface.Methods) != 3 {
			t.Errorf("Expected 3 methods in SplSubject, got %d", len(subjectIface.Methods))
		}
	})
}