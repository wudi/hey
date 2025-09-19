package runtime

import (
	"fmt"
	"strings"

	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/values"
)

// GeneratorExecutor is used to execute generator functions with yield support
type GeneratorExecutor struct {
	generator *ChannelGenerator
}

// ChannelGenerator implements PHP generators using Go channels
type ChannelGenerator struct {
	valueChan    chan *values.Value  // Channel for yielded values
	keyChan      chan *values.Value  // Channel for yielded keys
	closeChan    chan struct{}       // Channel to signal generator completion
	errorChan    chan error          // Channel for error propagation
	started      bool
	finished     bool
	currentKey   *values.Value
	currentValue *values.Value
	function     interface{} // *registry.Function
	args         []*values.Value
	vm           interface{} // *vm.VirtualMachine (avoid import cycle)
}

// NewChannelGenerator creates a new channel-based generator
func NewChannelGenerator(function interface{}, args []*values.Value, vm interface{}) *ChannelGenerator {
	return &ChannelGenerator{
		valueChan:    make(chan *values.Value),
		keyChan:      make(chan *values.Value),
		closeChan:    make(chan struct{}),
		errorChan:    make(chan error, 1), // Buffered to prevent blocking
		started:      false,
		finished:     false,
		currentKey:   values.NewNull(),
		currentValue: values.NewNull(),
		function:     function,
		args:         args,
		vm:           vm,
	}
}

// Start begins the generator execution in a separate goroutine
func (g *ChannelGenerator) Start() {
	if g.started {
		return
	}
	g.started = true

	// Launch generator function in goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.errorChan <- fmt.Errorf("generator panic: %v", r)
			}
			close(g.valueChan)
			close(g.keyChan)
			close(g.closeChan)
			g.finished = true
		}()

		// Execute generator function
		// This will be implemented to actually run the function
		// For now, simulate some yields for testing
		g.executeGeneratorFunction()
	}()
}

// executeGeneratorFunction executes the actual PHP generator function
func (g *ChannelGenerator) executeGeneratorFunction() {
	// Cast the function and VM interfaces back to their concrete types
	fn, ok := g.function.(*registry.Function)
	if !ok {
		g.errorChan <- fmt.Errorf("invalid function type")
		return
	}

	vm, ok := g.vm.(interface {
		CreateExecutionContext() interface{}
		CreateCallFrame(*registry.Function, []*values.Value) interface{}
		ExecuteFunction(interface{}, interface{}) error
	})
	if !ok {
		g.errorChan <- fmt.Errorf("invalid VM type")
		return
	}

	// Create execution context and call frame for generator
	ctx := vm.CreateExecutionContext()
	frame := vm.CreateCallFrame(fn, g.args)

	// Set generator reference in call frame
	if frameTyped, ok := frame.(interface{ SetGenerator(interface{}) }); ok {
		frameTyped.SetGenerator(g)
	}

	// Execute the generator function
	// The function will yield values through execYield calls
	err := vm.ExecuteFunction(ctx, frame)
	if err != nil {
		g.errorChan <- fmt.Errorf("generator execution error: %v", err)
		return
	}

	// Generator finished normally
	g.finished = true
}

// Next advances the generator to the next value
func (g *ChannelGenerator) Next() bool {
	if !g.started {
		g.Start()
	}

	if g.finished {
		return false
	}

	select {
	case key, ok := <-g.keyChan:
		if !ok {
			g.finished = true
			return false
		}
		g.currentKey = key

		value, ok := <-g.valueChan
		if !ok {
			g.finished = true
			return false
		}
		g.currentValue = value
		return true

	case err := <-g.errorChan:
		// Handle generator error
		fmt.Printf("Generator error: %v\n", err)
		g.finished = true
		return false

	case <-g.closeChan:
		g.finished = true
		return false
	}
}

// Current returns the current value
func (g *ChannelGenerator) Current() *values.Value {
	return g.currentValue
}

// Key returns the current key
func (g *ChannelGenerator) Key() *values.Value {
	return g.currentKey
}

// Valid returns whether the generator has more values
func (g *ChannelGenerator) Valid() bool {
	return !g.finished && g.started
}

// Rewind resets the generator (not supported for channel generators)
func (g *ChannelGenerator) Rewind() error {
	if g.started {
		return fmt.Errorf("Cannot rewind a generator that was already run")
	}
	return nil
}

// Yield is called from within the generator function to yield a value
func (g *ChannelGenerator) Yield(key, value *values.Value) {
	if key == nil {
		// Auto-increment key
		key = values.NewInt(0) // This should be tracked properly
	}

	select {
	case g.keyChan <- key:
	case <-g.closeChan:
		return
	}

	select {
	case g.valueChan <- value:
	case <-g.closeChan:
		return
	}
}


type builtinSpec struct {
	Name       string
	Parameters []*registry.Parameter
	ReturnType string
	IsVariadic bool
	MinArgs    int
	MaxArgs    int
	Impl       registry.BuiltinImplementation
}

var builtinClassStubs = map[string]map[string]struct{}{
	"stdclass": {},
	"exception": {
		"getmessage":       {},
		"getcode":          {},
		"getfile":          {},
		"getline":          {},
		"gettrace":         {},
		"gettraceasstring": {},
	},
}

var builtinFunctionSpecs = []builtinSpec{
	{
		Name:       "count",
		Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
		ReturnType: "int",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 {
				return values.NewInt(0), nil
			}
			val := args[0]
			if val == nil {
				return values.NewInt(0), nil
			}
			switch val.Type {
			case values.TypeArray:
				return values.NewInt(int64(val.ArrayCount())), nil
			case values.TypeObject:
				obj := val.Data.(*values.Object)
				return values.NewInt(int64(len(obj.Properties))), nil
			default:
				return values.NewInt(1), nil
			}
		},
	},
	{
		Name: "function_exists",
		Parameters: []*registry.Parameter{
			{Name: "function_name", Type: "string"},
		},
		ReturnType: "bool",
		MinArgs:    1,
		MaxArgs:    2,
		Impl: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil {
				return values.NewBool(false), nil
			}
			name := args[0].ToString()
			if name == "" {
				return values.NewBool(false), nil
			}
			if fn, ok := ctx.LookupUserFunction(name); ok && fn != nil {
				return values.NewBool(true), nil
			}
			if reg := ctx.SymbolRegistry(); reg != nil {
				if _, ok := reg.GetFunction(name); ok {
					return values.NewBool(true), nil
				}
			}
			return values.NewBool(false), nil
		},
	},
	{
		Name: "class_exists",
		Parameters: []*registry.Parameter{
			{Name: "class_name", Type: "string"},
		},
		ReturnType: "bool",
		MinArgs:    1,
		MaxArgs:    2,
		Impl: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil {
				return values.NewBool(false), nil
			}
			name := args[0].ToString()
			if name == "" {
				return values.NewBool(false), nil
			}
			nameLower := strings.ToLower(name)
			if _, ok := ctx.LookupUserClass(name); ok {
				return values.NewBool(true), nil
			}
			if reg := ctx.SymbolRegistry(); reg != nil {
				if _, err := reg.GetClass(name); err == nil {
					return values.NewBool(true), nil
				}
			}
			if _, ok := builtinClassStubs[nameLower]; ok {
				return values.NewBool(true), nil
			}
			return values.NewBool(false), nil
		},
	},
	{
		Name: "method_exists",
		Parameters: []*registry.Parameter{
			{Name: "object_or_class", Type: "mixed"},
			{Name: "method_name", Type: "string"},
		},
		ReturnType: "bool",
		MinArgs:    2,
		MaxArgs:    2,
		Impl: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 || args[1] == nil {
				return values.NewBool(false), nil
			}
			methodName := strings.ToLower(args[1].ToString())
			if methodName == "" {
				return values.NewBool(false), nil
			}
			var className string
			if args[0] != nil && args[0].IsObject() {
				className = args[0].Data.(*values.Object).ClassName
			} else if args[0] != nil {
				className = args[0].ToString()
			}
			if className == "" {
				return values.NewBool(false), nil
			}
			lowerClass := strings.ToLower(className)
			if classInfo, ok := ctx.LookupUserClass(className); ok && classInfo != nil {
				for name := range classInfo.Methods {
					if strings.ToLower(name) == methodName {
						return values.NewBool(true), nil
					}
				}
			}
			if reg := ctx.SymbolRegistry(); reg != nil {
				if desc, err := reg.GetClass(className); err == nil && desc != nil {
					for name := range desc.Methods {
						if strings.ToLower(name) == methodName {
							return values.NewBool(true), nil
						}
					}
				}
			}
			if methods, ok := builtinClassStubs[lowerClass]; ok {
				if _, exists := methods[methodName]; exists {
					return values.NewBool(true), nil
				}
			}
			return values.NewBool(false), nil
		},
	},
	{
		Name:       "array_keys",
		Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
		ReturnType: "array",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
				return values.NewArray(), nil
			}
			arr := args[0].Data.(*values.Array)
			result := values.NewArray()
			idx := int64(0)
			for key := range arr.Elements {
				var keyVal *values.Value
				switch k := key.(type) {
				case string:
					keyVal = values.NewString(k)
				case int:
					keyVal = values.NewInt(int64(k))
				case int64:
					keyVal = values.NewInt(k)
				default:
					keyVal = values.NewString(fmt.Sprintf("%v", k))
				}
				result.Data.(*values.Array).Elements[idx] = keyVal
				idx++
			}
			result.Data.(*values.Array).NextIndex = idx
			return result, nil
		},
	},
	{
		Name:       "array_values",
		Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
		ReturnType: "array",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
				return values.NewArray(), nil
			}
			arr := args[0].Data.(*values.Array)
			result := values.NewArray()
			idx := int64(0)
			for _, v := range arr.Elements {
				result.Data.(*values.Array).Elements[idx] = v
				idx++
			}
			result.Data.(*values.Array).NextIndex = idx
			return result, nil
		},
	},
	{
		Name:       "array_merge",
		Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
		ReturnType: "array",
		IsVariadic: true,
		MinArgs:    1,
		MaxArgs:    -1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			result := values.NewArray()
			targetArr := result.Data.(*values.Array)
			for _, arg := range args {
				if arg == nil || !arg.IsArray() {
					continue
				}
				src := arg.Data.(*values.Array)
				for key, val := range src.Elements {
					targetArr.Elements[key] = val
				}
				if src.NextIndex > targetArr.NextIndex {
					targetArr.NextIndex = src.NextIndex
				}
			}
			return result, nil
		},
	},
	{
		Name:       "strlen",
		Parameters: []*registry.Parameter{{Name: "str", Type: "string"}},
		ReturnType: "int",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil {
				return values.NewInt(0), nil
			}
			return values.NewInt(int64(len(args[0].ToString()))), nil
		},
	},
	{
		Name: "strpos",
		Parameters: []*registry.Parameter{
			{Name: "haystack", Type: "string"},
			{Name: "needle", Type: "string"},
		},
		ReturnType: "int|false",
		MinArgs:    2,
		MaxArgs:    2,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewBool(false), nil
			}
			haystack := args[0].ToString()
			needle := args[1].ToString()
			idx := strings.Index(haystack, needle)
			if idx == -1 {
				return values.NewBool(false), nil
			}
			return values.NewInt(int64(idx)), nil
		},
	},
	{
		Name: "substr",
		Parameters: []*registry.Parameter{
			{Name: "string", Type: "string"},
			{Name: "offset", Type: "int"},
		},
		ReturnType: "string",
		MinArgs:    2,
		MaxArgs:    -1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 {
				return values.NewString(""), nil
			}
			str := args[0].ToString()
			offset := int(args[1].ToInt())
			length := len(str)
			if len(args) >= 3 {
				requested := int(args[2].ToInt())
				if requested >= 0 && offset+requested < len(str) {
					length = offset + requested
				}
			}
			if offset < 0 {
				offset = len(str) + offset
			}
			if offset < 0 {
				offset = 0
			}
			if offset > len(str) {
				return values.NewString(""), nil
			}
			if length < offset {
				length = len(str)
			}
			return values.NewString(str[offset:length]), nil
		},
	},
	{
		Name:       "strtolower",
		Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
		ReturnType: "string",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil {
				return values.NewString(""), nil
			}
			return values.NewString(strings.ToLower(args[0].ToString())), nil
		},
	},
	{
		Name:       "strtoupper",
		Parameters: []*registry.Parameter{{Name: "string", Type: "string"}},
		ReturnType: "string",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil {
				return values.NewString(""), nil
			}
			return values.NewString(strings.ToUpper(args[0].ToString())), nil
		},
	},
	{
		Name:       "is_string",
		Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
		ReturnType: "bool",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil {
				return values.NewBool(false), nil
			}
			return values.NewBool(args[0].IsString()), nil
		},
	},
	{
		Name:       "is_int",
		Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
		ReturnType: "bool",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil {
				return values.NewBool(false), nil
			}
			return values.NewBool(args[0].IsInt()), nil
		},
	},
	{
		Name:       "is_array",
		Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
		ReturnType: "bool",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil {
				return values.NewBool(false), nil
			}
			return values.NewBool(args[0].IsArray()), nil
		},
	},
	{
		Name:       "print",
		Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
		ReturnType: "int",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) > 0 {
				if err := ctx.WriteOutput(args[0]); err != nil {
					return nil, err
				}
			}
			return values.NewInt(1), nil
		},
	},
	{
		Name:       "var_dump",
		Parameters: []*registry.Parameter{{Name: "value", Type: "mixed"}},
		ReturnType: "bool",
		IsVariadic: true,
		MinArgs:    1,
		MaxArgs:    -1,
		Impl: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if ctx != nil {
				for _, arg := range args {
					_ = ctx.WriteOutput(values.NewString(arg.VarDump()))
				}
			}
			return values.NewNull(), nil
		},
	},
	{
		Name:       "go",
		Parameters: []*registry.Parameter{{Name: "closure", Type: "callable"}},
		ReturnType: "Goroutine",
		IsVariadic: true,
		MinArgs:    1,
		MaxArgs:    -1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 {
				return nil, fmt.Errorf("go() expects at least one argument")
			}
			closureVal := args[0]
			if closureVal == nil || !closureVal.IsCallable() {
				return nil, fmt.Errorf("go() expects a callable as first argument")
			}
			closure := closureVal.ClosureGet()
			if closure == nil {
				return nil, fmt.Errorf("go() invalid closure")
			}
			useVars := make(map[string]*values.Value)
			for i, arg := range args[1:] {
				useVars[fmt.Sprintf("var_%d", i)] = arg
			}
			return values.NewGoroutine(closure, useVars), nil
		},
	},
	{
		Name:       "array_change_key_case",
		Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
		ReturnType: "array",
		MinArgs:    1,
		MaxArgs:    2,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
				return values.NewArray(), nil
			}
			arr := args[0].Data.(*values.Array)
			caseMode := int64(0)
			if len(args) > 1 && args[1] != nil {
				caseMode = args[1].ToInt()
			}
			result := values.NewArray()
			resultArr := result.Data.(*values.Array)
			for key, val := range arr.Elements {
				newKey := key
				if strKey, ok := key.(string); ok {
					if caseMode == 0 {
						newKey = strings.ToLower(strKey)
					} else {
						newKey = strings.ToUpper(strKey)
					}
				}
				resultArr.Elements[newKey] = val
			}
			resultArr.NextIndex = arr.NextIndex
			return result, nil
		},
	},
	{
		Name: "array_chunk",
		Parameters: []*registry.Parameter{
			{Name: "array", Type: "array"},
			{Name: "length", Type: "int"},
		},
		ReturnType: "array",
		MinArgs:    2,
		MaxArgs:    3,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
				return values.NewArray(), nil
			}
			arr := args[0].Data.(*values.Array)
			chunkSize := int(args[1].ToInt())
			if chunkSize <= 0 {
				return values.NewArray(), nil
			}
			preserveKeys := false
			if len(args) > 2 && args[2] != nil {
				preserveKeys = args[2].ToBool()
			}
			result := values.NewArray()
			resultArr := result.Data.(*values.Array)
			currentChunk := values.NewArray()
			currentChunkArr := currentChunk.Data.(*values.Array)
			chunkIdx := int64(0)
			itemCount := 0
			for key, val := range arr.Elements {
				if preserveKeys {
					currentChunkArr.Elements[key] = val
				} else {
					currentChunkArr.Elements[int64(itemCount)] = val
				}
				itemCount++
				if itemCount >= chunkSize {
					resultArr.Elements[chunkIdx] = currentChunk
					chunkIdx++
					currentChunk = values.NewArray()
					currentChunkArr = currentChunk.Data.(*values.Array)
					itemCount = 0
				}
			}
			if itemCount > 0 {
				if !preserveKeys {
					currentChunkArr.NextIndex = int64(itemCount)
				}
				resultArr.Elements[chunkIdx] = currentChunk
				chunkIdx++
			}
			resultArr.NextIndex = chunkIdx
			return result, nil
		},
	},
	{
		Name: "array_combine",
		Parameters: []*registry.Parameter{
			{Name: "keys", Type: "array"},
			{Name: "values", Type: "array"},
		},
		ReturnType: "array",
		MinArgs:    2,
		MaxArgs:    2,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 || args[0] == nil || !args[0].IsArray() || args[1] == nil || !args[1].IsArray() {
				return values.NewBool(false), nil
			}
			keysArr := args[0].Data.(*values.Array)
			valsArr := args[1].Data.(*values.Array)
			if args[0].ArrayCount() != args[1].ArrayCount() {
				return values.NewBool(false), nil
			}
			result := values.NewArray()
			resultArr := result.Data.(*values.Array)
			keysList := make([]*values.Value, 0, args[0].ArrayCount())
			for _, k := range keysArr.Elements {
				keysList = append(keysList, k)
			}
			valsList := make([]*values.Value, 0, args[1].ArrayCount())
			for _, v := range valsArr.Elements {
				valsList = append(valsList, v)
			}
			for i := 0; i < len(keysList) && i < len(valsList); i++ {
				keyVal := keysList[i]
				if keyVal.IsInt() {
					resultArr.Elements[keyVal.ToInt()] = valsList[i]
				} else {
					resultArr.Elements[keyVal.ToString()] = valsList[i]
				}
			}
			return result, nil
		},
	},
	{
		Name:       "array_count_values",
		Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
		ReturnType: "array",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
				return values.NewArray(), nil
			}
			arr := args[0].Data.(*values.Array)
			result := values.NewArray()
			resultArr := result.Data.(*values.Array)
			for _, val := range arr.Elements {
				if val == nil {
					continue
				}
				key := val.ToString()
				if existing, ok := resultArr.Elements[key]; ok && existing != nil {
					resultArr.Elements[key] = values.NewInt(existing.ToInt() + 1)
				} else {
					resultArr.Elements[key] = values.NewInt(1)
				}
			}
			return result, nil
		},
	},
	{
		Name: "array_diff",
		Parameters: []*registry.Parameter{
			{Name: "array", Type: "array"},
		},
		ReturnType: "array",
		IsVariadic: true,
		MinArgs:    2,
		MaxArgs:    -1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
				return values.NewArray(), nil
			}
			arr1 := args[0].Data.(*values.Array)
			otherValues := make(map[string]bool)
			for i := 1; i < len(args); i++ {
				if args[i] != nil && args[i].IsArray() {
					arr := args[i].Data.(*values.Array)
					for _, v := range arr.Elements {
						if v != nil {
							otherValues[v.ToString()] = true
						}
					}
				}
			}
			result := values.NewArray()
			resultArr := result.Data.(*values.Array)
			for key, val := range arr1.Elements {
				if val != nil && !otherValues[val.ToString()] {
					resultArr.Elements[key] = val
				}
			}
			return result, nil
		},
	},
	{
		Name:       "array_flip",
		Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
		ReturnType: "array",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
				return values.NewArray(), nil
			}
			arr := args[0].Data.(*values.Array)
			result := values.NewArray()
			resultArr := result.Data.(*values.Array)
			for key, val := range arr.Elements {
				if val == nil {
					continue
				}
				var keyStr string
				switch k := key.(type) {
				case string:
					keyStr = k
				case int:
					keyStr = fmt.Sprintf("%d", k)
				case int64:
					keyStr = fmt.Sprintf("%d", k)
				default:
					keyStr = fmt.Sprintf("%v", key)
				}
				if val.IsInt() {
					resultArr.Elements[val.ToInt()] = values.NewString(keyStr)
				} else {
					resultArr.Elements[val.ToString()] = values.NewString(keyStr)
				}
			}
			return result, nil
		},
	},
	{
		Name: "array_intersect",
		Parameters: []*registry.Parameter{
			{Name: "array", Type: "array"},
		},
		ReturnType: "array",
		IsVariadic: true,
		MinArgs:    2,
		MaxArgs:    -1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 2 || args[0] == nil || !args[0].IsArray() {
				return values.NewArray(), nil
			}
			arr1 := args[0].Data.(*values.Array)
			result := values.NewArray()
			resultArr := result.Data.(*values.Array)
			for key, val := range arr1.Elements {
				if val == nil {
					continue
				}
				found := true
				for i := 1; i < len(args); i++ {
					if args[i] == nil || !args[i].IsArray() {
						continue
					}
					arr := args[i].Data.(*values.Array)
					hasValue := false
					for _, v := range arr.Elements {
						if v != nil && v.ToString() == val.ToString() {
							hasValue = true
							break
						}
					}
					if !hasValue {
						found = false
						break
					}
				}
				if found {
					resultArr.Elements[key] = val
				}
			}
			return result, nil
		},
	},
	{
		Name:       "array_reverse",
		Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
		ReturnType: "array",
		MinArgs:    1,
		MaxArgs:    2,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
				return values.NewArray(), nil
			}
			arr := args[0].Data.(*values.Array)
			preserveKeys := false
			if len(args) > 1 && args[1] != nil {
				preserveKeys = args[1].ToBool()
			}
			elements := make([]struct {
				key interface{}
				val *values.Value
			}, 0, args[0].ArrayCount())
			for k, v := range arr.Elements {
				elements = append(elements, struct {
					key interface{}
					val *values.Value
				}{k, v})
			}
			for i, j := 0, len(elements)-1; i < j; i, j = i+1, j-1 {
				elements[i], elements[j] = elements[j], elements[i]
			}
			result := values.NewArray()
			resultArr := result.Data.(*values.Array)
			if preserveKeys {
				for _, elem := range elements {
					resultArr.Elements[elem.key] = elem.val
				}
			} else {
				idx := int64(0)
				for _, elem := range elements {
					resultArr.Elements[idx] = elem.val
					idx++
				}
				resultArr.NextIndex = idx
			}
			return result, nil
		},
	},
	{
		Name:       "array_sum",
		Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
		ReturnType: "number",
		MinArgs:    1,
		MaxArgs:    1,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
				return values.NewInt(0), nil
			}
			arr := args[0].Data.(*values.Array)
			sum := float64(0)
			hasFloat := false
			for _, val := range arr.Elements {
				if val == nil {
					continue
				}
				if val.IsFloat() {
					hasFloat = true
					sum += val.ToFloat()
				} else if val.IsInt() {
					sum += float64(val.ToInt())
				}
			}
			if hasFloat {
				return values.NewFloat(sum), nil
			}
			return values.NewInt(int64(sum)), nil
		},
	},
	{
		Name:       "array_unique",
		Parameters: []*registry.Parameter{{Name: "array", Type: "array"}},
		ReturnType: "array",
		MinArgs:    1,
		MaxArgs:    2,
		Impl: func(_ registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) == 0 || args[0] == nil || !args[0].IsArray() {
				return values.NewArray(), nil
			}
			arr := args[0].Data.(*values.Array)
			result := values.NewArray()
			resultArr := result.Data.(*values.Array)
			seen := make(map[string]bool)
			for key, val := range arr.Elements {
				if val == nil {
					continue
				}
				valStr := val.ToString()
				if !seen[valStr] {
					seen[valStr] = true
					resultArr.Elements[key] = val
				}
			}
			return result, nil
		},
	},
}

var builtinConstants = []*registry.ConstantDescriptor{
	// Case conversion constants for array_change_key_case
	intConst("CASE_LOWER", 0),
	intConst("CASE_UPPER", 1),

	// Sort flags for array functions
	intConst("SORT_REGULAR", 0),
	intConst("SORT_NUMERIC", 1),
	intConst("SORT_STRING", 2),
	intConst("SORT_DESC", 3),
	intConst("SORT_ASC", 4),
	intConst("SORT_LOCALE_STRING", 5),
	intConst("SORT_NATURAL", 6),
	intConst("SORT_FLAG_CASE", 8),

	// others
}

func intConst(name string, v int64) *registry.ConstantDescriptor {
	return &registry.ConstantDescriptor{Name: name, Value: values.NewInt(v)}
}

// helper to normalise missing args to NULL when builtin expects them.
func ensureArgs(args []*values.Value, expected int) []*values.Value {
	if len(args) >= expected {
		return args
	}
	padded := make([]*values.Value, expected)
	copy(padded, args)
	for i := len(args); i < expected; i++ {
		padded[i] = values.NewNull()
	}
	return padded
}

func registerWaitGroupClass() error {
	if registry.GlobalRegistry == nil {
		return fmt.Errorf("registry not initialized")
	}

	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{},
		},
		"Add": {
			Name:       "Add",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{{Name: "delta", Type: "int"}},
		},
		"Done": {
			Name:       "Done",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{},
		},
		"Wait": {
			Name:       "Wait",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{},
		},
	}

	desc := &registry.ClassDescriptor{
		Name:       "WaitGroup",
		Properties: make(map[string]*registry.PropertyDescriptor),
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}

	return registry.GlobalRegistry.RegisterClass(desc)
}

func registerExceptionClass() error {
	if registry.GlobalRegistry == nil {
		return fmt.Errorf("registry not initialized")
	}

	// Create builtin method implementations for Exception class
	constructorImpl := &registry.Function{
		Name:      "__construct",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Constructor is called after object creation, modify the object's properties
			// The 'this' object is passed as first argument in method calls
			if len(args) < 1 {
				return values.NewNull(), nil
			}
			thisObj := args[0]
			if !thisObj.IsObject() {
				return values.NewNull(), fmt.Errorf("__construct called on non-object")
			}

			obj := thisObj.Data.(*values.Object)
			if obj.Properties == nil {
				obj.Properties = make(map[string]*values.Value)
			}

			// Set message (arg[1] if present)
			message := ""
			if len(args) > 1 && args[1] != nil {
				message = args[1].ToString()
			}
			obj.Properties["message"] = values.NewString(message)

			// Set code (arg[2] if present)
			code := int64(0)
			if len(args) > 2 && args[2] != nil {
				code = args[2].ToInt()
			}
			obj.Properties["code"] = values.NewInt(code)

			// Set file and line (simplified - would need actual source tracking)
			obj.Properties["file"] = values.NewString("")
			obj.Properties["line"] = values.NewInt(0)

			// Constructor should not return a value (void), but return the object for now
			return thisObj, nil
		},
		Parameters: []*registry.Parameter{
			{Name: "message", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
			{Name: "code", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
			{Name: "previous", Type: "?Throwable", HasDefault: true, DefaultValue: values.NewNull()},
		},
	}

	getMessageImpl := &registry.Function{
		Name:      "getMessage",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewString(""), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if msg, ok := obj.Properties["message"]; ok {
					return msg, nil
				}
			}
			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getCodeImpl := &registry.Function{
		Name:      "getCode",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewInt(0), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if code, ok := obj.Properties["code"]; ok {
					return code, nil
				}
			}
			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getFileImpl := &registry.Function{
		Name:      "getFile",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewString(""), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if file, ok := obj.Properties["file"]; ok {
					return file, nil
				}
			}
			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getLineImpl := &registry.Function{
		Name:      "getLine",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewInt(0), nil
			}
			obj := args[0].Data.(*values.Object)
			if obj.Properties != nil {
				if line, ok := obj.Properties["line"]; ok {
					return line, nil
				}
			}
			return values.NewInt(0), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getTraceImpl := &registry.Function{
		Name:      "getTrace",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Return empty array for now - would need full stack trace implementation
			return values.NewArray(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	getTraceAsStringImpl := &registry.Function{
		Name:      "getTraceAsString",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			// Return empty string for now - would need full stack trace implementation
			return values.NewString(""), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors that point to the builtin implementations
	methods := map[string]*registry.MethodDescriptor{
		"__construct": {
			Name:       "__construct",
			Visibility: "public",
			Parameters: []*registry.ParameterDescriptor{
				{Name: "message", Type: "string", HasDefault: true, DefaultValue: values.NewString("")},
				{Name: "code", Type: "int", HasDefault: true, DefaultValue: values.NewInt(0)},
				{Name: "previous", Type: "?Throwable", HasDefault: true, DefaultValue: values.NewNull()},
			},
			Implementation: &BuiltinMethodImpl{Function: constructorImpl},
		},
		"getMessage": {
			Name:           "getMessage",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: getMessageImpl},
		},
		"getCode": {
			Name:           "getCode",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: getCodeImpl},
		},
		"getFile": {
			Name:           "getFile",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: getFileImpl},
		},
		"getLine": {
			Name:           "getLine",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: getLineImpl},
		},
		"getTrace": {
			Name:           "getTrace",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: getTraceImpl},
		},
		"getTraceAsString": {
			Name:           "getTraceAsString",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: getTraceAsStringImpl},
		},
	}

	// Define class properties
	properties := map[string]*registry.PropertyDescriptor{
		"message": {
			Name:         "message",
			Visibility:   "protected",
			Type:         "string",
			DefaultValue: values.NewString(""),
		},
		"code": {
			Name:         "code",
			Visibility:   "protected",
			Type:         "int",
			DefaultValue: values.NewInt(0),
		},
		"file": {
			Name:         "file",
			Visibility:   "protected",
			Type:         "string",
			DefaultValue: values.NewString(""),
		},
		"line": {
			Name:         "line",
			Visibility:   "protected",
			Type:         "int",
			DefaultValue: values.NewInt(0),
		},
	}

	desc := &registry.ClassDescriptor{
		Name:       "Exception",
		Properties: properties,
		Methods:    methods,
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}

	return registry.GlobalRegistry.RegisterClass(desc)
}

// BuiltinMethodImpl implements MethodImplementation for builtin methods
type BuiltinMethodImpl struct {
	Function *registry.Function
}

func (b *BuiltinMethodImpl) ImplementationKind() string { return "builtin" }

func (b *BuiltinMethodImpl) GetFunction() *registry.Function {
	return b.Function
}

func registerTraversableInterface() error {
	if registry.GlobalRegistry == nil {
		return fmt.Errorf("registry not initialized")
	}

	iface := &registry.Interface{
		Name:    "Traversable",
		Methods: make(map[string]*registry.InterfaceMethod),
		Extends: []string{},
	}

	return registry.GlobalRegistry.RegisterInterface(iface)
}

func registerIteratorInterface() error {
	if registry.GlobalRegistry == nil {
		return fmt.Errorf("registry not initialized")
	}

	// Create method definitions for Iterator interface
	methods := map[string]*registry.InterfaceMethod{
		"current": {
			Name:       "current",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "mixed",
		},
		"key": {
			Name:       "key",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "mixed",
		},
		"next": {
			Name:       "next",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "void",
		},
		"rewind": {
			Name:       "rewind",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "void",
		},
		"valid": {
			Name:       "valid",
			Visibility: "public",
			Parameters: []*registry.Parameter{},
			ReturnType: "bool",
		},
	}

	iface := &registry.Interface{
		Name:    "Iterator",
		Methods: methods,
		Extends: []string{"Traversable"},
	}

	return registry.GlobalRegistry.RegisterInterface(iface)
}

func registerGeneratorClass() error {
	if registry.GlobalRegistry == nil {
		return fmt.Errorf("registry not initialized")
	}

	// Create method implementations for Generator class
	currentImpl := &registry.Function{
		Name:      "current",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)

			// Get channel generator
			if channelGen, ok := obj.Properties["__channel_generator"]; ok && channelGen != nil {
				if gen, ok := channelGen.Data.(*ChannelGenerator); ok {
					return gen.Current(), nil
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	keyImpl := &registry.Function{
		Name:      "key",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)

			// Get channel generator
			if channelGen, ok := obj.Properties["__channel_generator"]; ok && channelGen != nil {
				if gen, ok := channelGen.Data.(*ChannelGenerator); ok {
					return gen.Key(), nil
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	nextImpl := &registry.Function{
		Name:      "next",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)

			// Get channel generator
			if channelGen, ok := obj.Properties["__channel_generator"]; ok && channelGen != nil {
				if gen, ok := channelGen.Data.(*ChannelGenerator); ok {
					gen.Next() // Advance to next value
					return values.NewNull(), nil
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	rewindImpl := &registry.Function{
		Name:      "rewind",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewNull(), nil
			}
			obj := args[0].Data.(*values.Object)

			// Get channel generator
			if channelGen, ok := obj.Properties["__channel_generator"]; ok && channelGen != nil {
				if gen, ok := channelGen.Data.(*ChannelGenerator); ok {
					if err := gen.Rewind(); err != nil {
						return values.NewNull(), err
					}
					gen.Next() // Start and get first value
					return values.NewNull(), nil
				}
			}

			return values.NewNull(), nil
		},
		Parameters: []*registry.Parameter{},
	}

	validImpl := &registry.Function{
		Name:      "valid",
		IsBuiltin: true,
		Builtin: func(ctx registry.BuiltinCallContext, args []*values.Value) (*values.Value, error) {
			if len(args) < 1 || !args[0].IsObject() {
				return values.NewBool(false), nil
			}
			obj := args[0].Data.(*values.Object)

			// Get channel generator
			if channelGen, ok := obj.Properties["__channel_generator"]; ok && channelGen != nil {
				if gen, ok := channelGen.Data.(*ChannelGenerator); ok {
					return values.NewBool(gen.Valid()), nil
				}
			}

			return values.NewBool(false), nil
		},
		Parameters: []*registry.Parameter{},
	}

	// Create method descriptors
	methods := map[string]*registry.MethodDescriptor{
		"current": {
			Name:           "current",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: currentImpl},
		},
		"key": {
			Name:           "key",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: keyImpl},
		},
		"next": {
			Name:           "next",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: nextImpl},
		},
		"rewind": {
			Name:           "rewind",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: rewindImpl},
		},
		"valid": {
			Name:           "valid",
			Visibility:     "public",
			Parameters:     []*registry.ParameterDescriptor{},
			Implementation: &BuiltinMethodImpl{Function: validImpl},
		},
	}

	desc := &registry.ClassDescriptor{
		Name:       "Generator",
		IsFinal:    true,
		Interfaces: []string{"Iterator"},
		Methods:    methods,
		Properties: make(map[string]*registry.PropertyDescriptor),
		Constants:  make(map[string]*registry.ConstantDescriptor),
	}

	return registry.GlobalRegistry.RegisterClass(desc)
}
