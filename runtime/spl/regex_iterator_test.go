package spl

import (
	"testing"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

func TestRegexIterator(t *testing.T) {
	// Initialize registry
	registry.Initialize()
	ctx := &mockContext{registry: registry.GlobalRegistry}

	// Register all SPL classes for testing
	for _, class := range GetSplClasses() {
		if err := registry.GlobalRegistry.RegisterClass(class); err != nil {
			t.Fatalf("Failed to register class %s: %v", class.Name, err)
		}
	}

	// Get the RegexIterator class
	class := GetRegexIteratorClass()
	if class == nil {
		t.Fatal("RegexIterator class is nil")
	}

	// Create test ArrayIterator with string data
	arrayIterClass := GetArrayIteratorClass()
	arrayIterObj := &values.Object{
		ClassName:  "ArrayIterator",
		Properties: make(map[string]*values.Value),
	}
	arrayIterValue := &values.Value{
		Type: values.TypeObject,
		Data: arrayIterObj,
	}

	testArray := values.NewArray()
	testArray.ArraySet(values.NewInt(0), values.NewString("hello"))
	testArray.ArraySet(values.NewInt(1), values.NewString("world"))
	testArray.ArraySet(values.NewInt(2), values.NewString("test123"))
	testArray.ArraySet(values.NewInt(3), values.NewString("apple"))
	testArray.ArraySet(values.NewInt(4), values.NewString("banana"))
	testArray.ArraySet(values.NewInt(5), values.NewString("cherry"))
	testArray.ArraySet(values.NewInt(6), values.NewString("123number"))
	testArray.ArraySet(values.NewInt(7), values.NewString("regex_match"))
	testArray.ArraySet(values.NewInt(8), values.NewString("no_match"))

	constructorArgs := []*values.Value{arrayIterValue, testArray}
	arrayIterConstructor := arrayIterClass.Methods["__construct"]
	impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
	_, err := impl.GetFunction().Builtin(ctx, constructorArgs)
	if err != nil {
		t.Fatalf("Failed to initialize ArrayIterator: %v", err)
	}

	t.Run("Constructor", func(t *testing.T) {
		// Create RegexIterator instance
		obj := &values.Object{
			ClassName:  "RegexIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructor := class.Methods["__construct"]
		if constructor == nil {
			t.Fatal("Constructor not found")
		}

		// Test with basic regex pattern
		args := []*values.Value{thisObj, arrayIterValue, values.NewString("/\\d+/"), values.NewInt(REGEX_ITERATOR_MATCH)}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Check that parameters were stored
		if iterator, ok := obj.Properties["__iterator"]; !ok || iterator != arrayIterValue {
			t.Fatal("Iterator not stored correctly")
		}
		if regex, ok := obj.Properties["__regex"]; !ok || regex.ToString() != "/\\d+/" {
			t.Fatal("Regex not stored correctly")
		}
		if mode, ok := obj.Properties["__mode"]; !ok || mode.ToInt() != REGEX_ITERATOR_MATCH {
			t.Fatal("Mode not stored correctly")
		}
	})

	t.Run("BasicFiltering", func(t *testing.T) {
		// Create fresh RegexIterator for this test
		obj := &values.Object{
			ClassName:  "RegexIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		// Create fresh ArrayIterator
		freshArrayIterObj := &values.Object{
			ClassName:  "ArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		freshArrayIterValue := &values.Value{
			Type: values.TypeObject,
			Data: freshArrayIterObj,
		}

		freshTestArray := values.NewArray()
		freshTestArray.ArraySet(values.NewInt(0), values.NewString("hello"))
		freshTestArray.ArraySet(values.NewInt(1), values.NewString("world"))
		freshTestArray.ArraySet(values.NewInt(2), values.NewString("test123"))
		freshTestArray.ArraySet(values.NewInt(3), values.NewString("apple"))
		freshTestArray.ArraySet(values.NewInt(4), values.NewString("123number"))

		constructorArgs := []*values.Value{freshArrayIterValue, freshTestArray}
		arrayIterConstructor := arrayIterClass.Methods["__construct"]
		impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, constructorArgs)
		if err != nil {
			t.Fatalf("Failed to initialize fresh ArrayIterator: %v", err)
		}

		// Initialize RegexIterator to match strings with numbers
		constructor := class.Methods["__construct"]
		args := []*values.Value{thisObj, freshArrayIterValue, values.NewString("/\\d+/"), values.NewInt(REGEX_ITERATOR_MATCH)}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Get iteration methods
		rewindMethod := class.Methods["rewind"]
		validMethod := class.Methods["valid"]
		currentMethod := class.Methods["current"]
		keyMethod := class.Methods["key"]
		nextMethod := class.Methods["next"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		// Rewind to start
		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// Should match "test123" (index 2)
		validResult, _ := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if !validResult.ToBool() {
			t.Fatal("Should be valid for first matching element")
		}

		currentResult, _ := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if currentResult.ToString() != "test123" {
			t.Fatalf("Expected 'test123', got '%s'", currentResult.ToString())
		}

		keyResult, _ := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if keyResult.ToInt() != 2 {
			t.Fatalf("Expected key 2, got %d", keyResult.ToInt())
		}

		// Move to next match
		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("next failed: %v", err)
		}

		// Should match "123number" (index 4)
		validResult, _ = validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if !validResult.ToBool() {
			t.Fatal("Should be valid for second matching element")
		}

		currentResult, _ = currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if currentResult.ToString() != "123number" {
			t.Fatalf("Expected '123number', got '%s'", currentResult.ToString())
		}

		keyResult, _ = keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if keyResult.ToInt() != 4 {
			t.Fatalf("Expected key 4, got %d", keyResult.ToInt())
		}

		// Move to next - should be invalid (no more matches)
		_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("next failed: %v", err)
		}

		validResult, _ = validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if validResult.ToBool() {
			t.Fatal("Should be invalid after all matches consumed")
		}
	})

	t.Run("ModeGetMatch", func(t *testing.T) {
		// Create fresh RegexIterator for GET_MATCH mode test
		obj := &values.Object{
			ClassName:  "RegexIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		// Create fresh ArrayIterator with data that has capture groups
		freshArrayIterObj := &values.Object{
			ClassName:  "ArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		freshArrayIterValue := &values.Value{
			Type: values.TypeObject,
			Data: freshArrayIterObj,
		}

		freshTestArray := values.NewArray()
		freshTestArray.ArraySet(values.NewInt(0), values.NewString("regex_match"))
		freshTestArray.ArraySet(values.NewInt(1), values.NewString("no_match"))
		freshTestArray.ArraySet(values.NewInt(2), values.NewString("test_data"))

		constructorArgs := []*values.Value{freshArrayIterValue, freshTestArray}
		arrayIterConstructor := arrayIterClass.Methods["__construct"]
		impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, constructorArgs)
		if err != nil {
			t.Fatalf("Failed to initialize fresh ArrayIterator: %v", err)
		}

		// Initialize RegexIterator with GET_MATCH mode and capture groups
		constructor := class.Methods["__construct"]
		args := []*values.Value{thisObj, freshArrayIterValue, values.NewString("/(\\w+)_(\\w+)/"), values.NewInt(REGEX_ITERATOR_GET_MATCH)}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Test iteration
		rewindMethod := class.Methods["rewind"]
		currentMethod := class.Methods["current"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)

		// Rewind to start
		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		// Should return array of matches for "regex_match"
		currentResult, _ := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if !currentResult.IsArray() {
			t.Fatal("Expected array result for GET_MATCH mode")
		}

		matchArray := currentResult.Data.(*values.Array)
		if len(matchArray.Elements) != 3 {
			t.Fatalf("Expected 3 captures, got %d", len(matchArray.Elements))
		}

		// Check capture groups: [0] = full match, [1] = "regex", [2] = "match"
		fullMatch := matchArray.Elements[int64(0)]
		if fullMatch.ToString() != "regex_match" {
			t.Fatalf("Expected full match 'regex_match', got '%s'", fullMatch.ToString())
		}

		group1 := matchArray.Elements[int64(1)]
		if group1.ToString() != "regex" {
			t.Fatalf("Expected capture group 1 'regex', got '%s'", group1.ToString())
		}

		group2 := matchArray.Elements[int64(2)]
		if group2.ToString() != "match" {
			t.Fatalf("Expected capture group 2 'match', got '%s'", group2.ToString())
		}
	})

	t.Run("UseKeyFlag", func(t *testing.T) {
		// Create fresh RegexIterator to test USE_KEY flag
		obj := &values.Object{
			ClassName:  "RegexIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		// Create fresh ArrayIterator
		freshArrayIterObj := &values.Object{
			ClassName:  "ArrayIterator",
			Properties: make(map[string]*values.Value),
		}
		freshArrayIterValue := &values.Value{
			Type: values.TypeObject,
			Data: freshArrayIterObj,
		}

		freshTestArray := values.NewArray()
		freshTestArray.ArraySet(values.NewInt(0), values.NewString("value0"))
		freshTestArray.ArraySet(values.NewInt(1), values.NewString("value1"))
		freshTestArray.ArraySet(values.NewInt(2), values.NewString("value2"))
		freshTestArray.ArraySet(values.NewInt(10), values.NewString("value10"))

		constructorArgs := []*values.Value{freshArrayIterValue, freshTestArray}
		arrayIterConstructor := arrayIterClass.Methods["__construct"]
		impl := arrayIterConstructor.Implementation.(*BuiltinMethodImpl)
		_, err := impl.GetFunction().Builtin(ctx, constructorArgs)
		if err != nil {
			t.Fatalf("Failed to initialize fresh ArrayIterator: %v", err)
		}

		// Initialize RegexIterator with USE_KEY flag to match keys with pattern
		constructor := class.Methods["__construct"]
		args := []*values.Value{thisObj, freshArrayIterValue, values.NewString("/^[0-2]$/"), values.NewInt(REGEX_ITERATOR_MATCH), values.NewInt(REGEX_ITERATOR_USE_KEY)}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err = constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		// Test that it filters by keys, not values
		rewindMethod := class.Methods["rewind"]
		validMethod := class.Methods["valid"]
		currentMethod := class.Methods["current"]
		keyMethod := class.Methods["key"]
		nextMethod := class.Methods["next"]

		rewindImpl := rewindMethod.Implementation.(*BuiltinMethodImpl)
		validImpl := validMethod.Implementation.(*BuiltinMethodImpl)
		currentImpl := currentMethod.Implementation.(*BuiltinMethodImpl)
		keyImpl := keyMethod.Implementation.(*BuiltinMethodImpl)
		nextImpl := nextMethod.Implementation.(*BuiltinMethodImpl)

		// Rewind and collect all matches
		_, err = rewindImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("rewind failed: %v", err)
		}

		matchedKeys := []int64{}
		matchedValues := []string{}

		for {
			validResult, _ := validImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if !validResult.ToBool() {
				break
			}

			keyResult, _ := keyImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			currentResult, _ := currentImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})

			matchedKeys = append(matchedKeys, keyResult.ToInt())
			matchedValues = append(matchedValues, currentResult.ToString())

			_, err = nextImpl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
			if err != nil {
				break
			}
		}

		// Should match keys 0, 1, 2 (but not 10)
		expectedKeys := []int64{0, 1, 2}
		if len(matchedKeys) != len(expectedKeys) {
			t.Fatalf("Expected %d matches, got %d", len(expectedKeys), len(matchedKeys))
		}

		for i, expectedKey := range expectedKeys {
			if matchedKeys[i] != expectedKey {
				t.Fatalf("Expected key %d at position %d, got %d", expectedKey, i, matchedKeys[i])
			}
		}
	})

	t.Run("getInnerIterator", func(t *testing.T) {
		// Create RegexIterator instance
		obj := &values.Object{
			ClassName:  "RegexIterator",
			Properties: make(map[string]*values.Value),
		}
		thisObj := &values.Value{
			Type: values.TypeObject,
			Data: obj,
		}

		constructor := class.Methods["__construct"]
		args := []*values.Value{thisObj, arrayIterValue, values.NewString("/test/"), values.NewInt(REGEX_ITERATOR_MATCH)}
		constructorImpl := constructor.Implementation.(*BuiltinMethodImpl)
		_, err := constructorImpl.GetFunction().Builtin(ctx, args)
		if err != nil {
			t.Fatalf("Constructor failed: %v", err)
		}

		method := class.Methods["getInnerIterator"]
		if method == nil {
			t.Fatal("getInnerIterator method not found")
		}

		impl := method.Implementation.(*BuiltinMethodImpl)
		result, err := impl.GetFunction().Builtin(ctx, []*values.Value{thisObj})
		if err != nil {
			t.Fatalf("getInnerIterator failed: %v", err)
		}

		if result != arrayIterValue {
			t.Fatal("getInnerIterator should return the wrapped iterator")
		}
	})
}