package spl

import (
	"github.com/wudi/hey/registry"
)

// GetSplClasses returns all SPL class definitions
func GetSplClasses() []*registry.ClassDescriptor {
	return []*registry.ClassDescriptor{
		GetArrayIteratorClass(),
		GetArrayObjectClass(),
		GetSplDoublyLinkedListClass(),
		GetSplStackClass(),
		GetSplQueueClass(),
		GetSplFixedArrayClass(),
		GetSplObjectStorageClass(),
		GetEmptyIteratorClass(),
		GetSplFileInfoClass(),
		GetIteratorIteratorClass(),
		GetLimitIteratorClass(),
		GetAppendIteratorClass(),
		GetFilterIteratorClass(),
		GetCallbackFilterIteratorClass(),
		GetRecursiveArrayIteratorClass(),
		GetRecursiveIteratorIteratorClass(),
		GetNoRewindIteratorClass(),
		GetInfiniteIteratorClass(),
		GetMultipleIteratorClass(),
		GetCachingIteratorClass(),
		GetRegexIteratorClass(),
		GetSplHeapClass(),
		GetSplMaxHeapClass(),
		GetSplMinHeapClass(),
		GetSplPriorityQueueClass(),
		GetDirectoryIteratorClass(),
		GetFilesystemIteratorClass(),
		GetSplFileObjectClass(),
		GetSplTempFileObjectClass(),
		GetGlobIteratorClass(),
		GetRecursiveDirectoryIteratorClass(),
		GetRecursiveFilterIteratorClass(),
		GetParentIteratorClass(),
		GetRecursiveCachingIteratorClass(),
		GetRecursiveCallbackFilterIteratorClass(),
		GetRecursiveRegexIteratorClass(),
		GetRecursiveTreeIteratorClass(),
	}
}

// GetSplInterfaces returns all SPL interface definitions
func GetSplInterfaces() []*registry.Interface {
	return []*registry.Interface{
		// SPL extends existing interfaces, so we return additional ones
		getArrayAccessInterface(),
		getCountableInterface(),
		getIteratorAggregateInterface(),
		getSeekableIteratorInterface(),
		getRecursiveIteratorInterface(),
		getOuterIteratorInterface(),
		getSplObserverInterface(),
		getSplSubjectInterface(),
	}
}

// getArrayAccessInterface returns the ArrayAccess interface
func getArrayAccessInterface() *registry.Interface {
	methods := map[string]*registry.InterfaceMethod{
		"offsetExists": {
			Name:       "offsetExists",
			Visibility: "public",
			Parameters: []*registry.Parameter{
				{Name: "offset", Type: "mixed"},
			},
			ReturnType: "bool",
		},
		"offsetGet": {
			Name:       "offsetGet",
			Visibility: "public",
			Parameters: []*registry.Parameter{
				{Name: "offset", Type: "mixed"},
			},
			ReturnType: "mixed",
		},
		"offsetSet": {
			Name:       "offsetSet",
			Visibility: "public",
			Parameters: []*registry.Parameter{
				{Name: "offset", Type: "mixed"},
				{Name: "value", Type: "mixed"},
			},
			ReturnType: "void",
		},
		"offsetUnset": {
			Name:       "offsetUnset",
			Visibility: "public",
			Parameters: []*registry.Parameter{
				{Name: "offset", Type: "mixed"},
			},
			ReturnType: "void",
		},
	}

	return &registry.Interface{
		Name:    "ArrayAccess",
		Methods: methods,
		Extends: []string{},
	}
}

// getCountableInterface returns the Countable interface
func getCountableInterface() *registry.Interface {
	methods := map[string]*registry.InterfaceMethod{
		"count": {
			Name:       "count",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "int",
		},
	}

	return &registry.Interface{
		Name:    "Countable",
		Methods: methods,
		Extends: []string{},
	}
}

// getIteratorAggregateInterface returns the IteratorAggregate interface
func getIteratorAggregateInterface() *registry.Interface {
	methods := map[string]*registry.InterfaceMethod{
		"getIterator": {
			Name:       "getIterator",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "Traversable",
		},
	}

	return &registry.Interface{
		Name:    "IteratorAggregate",
		Methods: methods,
		Extends: []string{"Traversable"},
	}
}

// getSeekableIteratorInterface returns the SeekableIterator interface
func getSeekableIteratorInterface() *registry.Interface {
	methods := map[string]*registry.InterfaceMethod{
		"seek": {
			Name:       "seek",
			Visibility: "public",
			Parameters: []*registry.Parameter{
				{Name: "position", Type: "int"},
			},
			ReturnType: "void",
		},
	}

	return &registry.Interface{
		Name:    "SeekableIterator",
		Methods: methods,
		Extends: []string{"Iterator"},
	}
}

// getRecursiveIteratorInterface returns the RecursiveIterator interface
func getRecursiveIteratorInterface() *registry.Interface {
	methods := map[string]*registry.InterfaceMethod{
		"hasChildren": {
			Name:       "hasChildren",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
		},
		"getChildren": {
			Name:       "getChildren",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "RecursiveIterator",
		},
	}

	return &registry.Interface{
		Name:    "RecursiveIterator",
		Methods: methods,
		Extends: []string{"Iterator"},
	}
}

// getOuterIteratorInterface returns the OuterIterator interface
func getOuterIteratorInterface() *registry.Interface {
	methods := map[string]*registry.InterfaceMethod{
		"getInnerIterator": {
			Name:       "getInnerIterator",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "Iterator",
		},
	}

	return &registry.Interface{
		Name:    "OuterIterator",
		Methods: methods,
		Extends: []string{"Iterator"},
	}
}

// getSplObserverInterface returns the SplObserver interface
func getSplObserverInterface() *registry.Interface {
	methods := map[string]*registry.InterfaceMethod{
		"update": {
			Name:       "update",
			Visibility: "public",
			Parameters: []*registry.Parameter{
				{Name: "subject", Type: "SplSubject"},
			},
			ReturnType: "void",
		},
	}

	return &registry.Interface{
		Name:    "SplObserver",
		Methods: methods,
		Extends: []string{},
	}
}

// getSplSubjectInterface returns the SplSubject interface
func getSplSubjectInterface() *registry.Interface {
	methods := map[string]*registry.InterfaceMethod{
		"attach": {
			Name:       "attach",
			Visibility: "public",
			Parameters: []*registry.Parameter{
				{Name: "observer", Type: "SplObserver"},
			},
			ReturnType: "void",
		},
		"detach": {
			Name:       "detach",
			Visibility: "public",
			Parameters: []*registry.Parameter{
				{Name: "observer", Type: "SplObserver"},
			},
			ReturnType: "void",
		},
		"notify": {
			Name:       "notify",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "void",
		},
	}

	return &registry.Interface{
		Name:    "SplSubject",
		Methods: methods,
		Extends: []string{},
	}
}