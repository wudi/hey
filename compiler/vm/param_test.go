package vm

import (
	"testing"

	"github.com/wudi/hey/compiler/opcodes"
	"github.com/wudi/hey/compiler/values"
)

func TestRecvOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up parameters
	ctx.Parameters = []*values.Value{
		values.NewString("param1"),
		values.NewInt(42),
		values.NewBool(true),
	}
	ctx.Temporaries = make(map[uint32]*values.Value)

	// Create RECV instruction to receive parameter 1 (42)
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     1, // Parameter index
		Op2:     0,
		Result:  0, // Store in temporary variable 0
	}

	// Execute RECV
	err := vm.executeRecv(ctx, &inst)
	if err != nil {
		t.Fatalf("RECV execution failed: %v", err)
	}

	// Check result
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("RECV result is nil")
	}

	if result.Type != values.TypeInt || result.Data.(int64) != 42 {
		t.Errorf("Expected parameter value 42, got %v", result.Data)
	}
}

func TestRecvNonExistentParameter(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up only 2 parameters
	ctx.Parameters = []*values.Value{
		values.NewString("param1"),
		values.NewInt(42),
	}
	ctx.Temporaries = make(map[uint32]*values.Value)

	// Create RECV instruction to receive parameter 5 (doesn't exist)
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     5, // Parameter index (out of bounds)
		Result:  0, // Store in temporary variable 0
	}

	// Execute RECV
	err := vm.executeRecv(ctx, &inst)
	if err != nil {
		t.Fatalf("RECV execution failed: %v", err)
	}

	// Check result - should be null
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("RECV result is nil")
	}

	if !result.IsNull() {
		t.Error("Expected null for non-existent parameter")
	}
}

func TestRecvInitOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up parameters (missing parameter 2)
	ctx.Parameters = []*values.Value{
		values.NewString("param1"),
		values.NewInt(42),
	}
	ctx.Temporaries = make(map[uint32]*values.Value)
	// Set up default value in temporary variable 1
	ctx.Temporaries[1] = values.NewString("default_value")

	// Create RECV_INIT instruction to receive parameter 2 with default
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV_INIT,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     2, // Parameter index (doesn't exist)
		Op2:     1, // Default value from temporary variable 1
		Result:  0, // Store in temporary variable 0
	}

	// Execute RECV_INIT
	err := vm.executeRecvInit(ctx, &inst)
	if err != nil {
		t.Fatalf("RECV_INIT execution failed: %v", err)
	}

	// Check result - should be default value
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("RECV_INIT result is nil")
	}

	if result.Type != values.TypeString || result.Data.(string) != "default_value" {
		t.Errorf("Expected default value 'default_value', got %v", result.Data)
	}
}

func TestRecvInitWithProvidedParameter(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up parameters (parameter 1 exists)
	ctx.Parameters = []*values.Value{
		values.NewString("param1"),
		values.NewInt(42),
		values.NewString("provided_param"),
	}
	ctx.Temporaries = make(map[uint32]*values.Value)
	// Set up default value in temporary variable 1
	ctx.Temporaries[1] = values.NewString("default_value")

	// Create RECV_INIT instruction to receive parameter 2 with default
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV_INIT,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     2, // Parameter index (exists)
		Op2:     1, // Default value from temporary variable 1
		Result:  0, // Store in temporary variable 0
	}

	// Execute RECV_INIT
	err := vm.executeRecvInit(ctx, &inst)
	if err != nil {
		t.Fatalf("RECV_INIT execution failed: %v", err)
	}

	// Check result - should be provided parameter, not default
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("RECV_INIT result is nil")
	}

	if result.Type != values.TypeString || result.Data.(string) != "provided_param" {
		t.Errorf("Expected provided parameter 'provided_param', got %v", result.Data)
	}
}

func TestRecvVariadicOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Set up parameters (5 total, variadic starts from index 2)
	ctx.Parameters = []*values.Value{
		values.NewString("param1"),
		values.NewInt(42),
		values.NewString("variadic1"),
		values.NewInt(123),
		values.NewBool(true),
	}
	ctx.Temporaries = make(map[uint32]*values.Value)

	// Create RECV_VARIADIC instruction starting from parameter 2
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV_VARIADIC,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     2, // Start collecting from parameter 2
		Result:  0, // Store array in temporary variable 0
	}

	// Execute RECV_VARIADIC
	err := vm.executeRecvVariadic(ctx, &inst)
	if err != nil {
		t.Fatalf("RECV_VARIADIC execution failed: %v", err)
	}

	// Check result - should be array with 3 elements
	result := ctx.Temporaries[0]
	if result == nil {
		t.Fatal("RECV_VARIADIC result is nil")
	}

	if result.Type != values.TypeArray {
		t.Errorf("Expected array type, got %v", result.Type)
	}

	// Check array contents
	if result.ArrayCount() != 3 {
		t.Errorf("Expected 3 variadic parameters, got %d", result.ArrayCount())
	}

	// Check first variadic element (index 0 in array, parameter 2 in call)
	elem0 := result.ArrayGet(values.NewInt(0))
	if elem0.Type != values.TypeString || elem0.Data.(string) != "variadic1" {
		t.Errorf("Expected first variadic element 'variadic1', got %v", elem0.Data)
	}

	// Check second variadic element
	elem1 := result.ArrayGet(values.NewInt(1))
	if elem1.Type != values.TypeInt || elem1.Data.(int64) != 123 {
		t.Errorf("Expected second variadic element 123, got %v", elem1.Data)
	}
}

func TestSendVarExOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("send_value")

	// Create SEND_VAR_EX instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_SEND_VAR_EX,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Send from temporary variable 0
		Result:  1, // Store result in temporary variable 1
	}

	// Execute SEND_VAR_EX
	err := vm.executeSendVarEx(ctx, &inst)
	if err != nil {
		t.Fatalf("SEND_VAR_EX execution failed: %v", err)
	}

	// Check result - should be same as sent value
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("SEND_VAR_EX result is nil")
	}

	if result.Type != values.TypeString || result.Data.(string) != "send_value" {
		t.Errorf("Expected sent value 'send_value', got %v", result.Data)
	}

	// Check call arguments
	if len(ctx.CallArguments) != 1 {
		t.Errorf("Expected 1 call argument, got %d", len(ctx.CallArguments))
	}

	if ctx.CallArguments[0].Data.(string) != "send_value" {
		t.Errorf("Expected call argument 'send_value', got %v", ctx.CallArguments[0].Data)
	}
}

func TestSendVarNoRefOpcode(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[0] = values.NewString("send_value")

	// Create SEND_VAR_NO_REF instruction
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_TMP_VAR, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	inst := opcodes.Instruction{
		Opcode:  opcodes.OP_SEND_VAR_NO_REF,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Send from temporary variable 0
		Result:  1, // Store result in temporary variable 1
	}

	// Execute SEND_VAR_NO_REF
	err := vm.executeSendVarNoRef(ctx, &inst)
	if err != nil {
		t.Fatalf("SEND_VAR_NO_REF execution failed: %v", err)
	}

	// Check result - should be copy of sent value
	result := ctx.Temporaries[1]
	if result == nil {
		t.Fatal("SEND_VAR_NO_REF result is nil")
	}

	if result.Type != values.TypeString || result.Data.(string) != "send_value" {
		t.Errorf("Expected copied value 'send_value', got %v", result.Data)
	}

	// Check call arguments
	if len(ctx.CallArguments) != 1 {
		t.Errorf("Expected 1 call argument, got %d", len(ctx.CallArguments))
	}

	// Verify the argument is a copy (different pointer but same data)
	if ctx.CallArguments[0] == ctx.Temporaries[0] {
		t.Error("Expected copied value, but got same pointer (reference)")
	}

	if ctx.CallArguments[0].Data.(string) != "send_value" {
		t.Errorf("Expected call argument 'send_value', got %v", ctx.CallArguments[0].Data)
	}
}

// Test a complete parameter passing simulation
func TestParameterPassingSimulation(t *testing.T) {
	vm := NewVirtualMachine()
	ctx := NewExecutionContext()

	// Simulate function call: func(param1, param2="default", ...rest)
	// Called with: func("hello", 42, "extra1", "extra2")

	// Set up parameters for the function
	ctx.Parameters = []*values.Value{
		values.NewString("hello"),  // param1
		values.NewInt(42),          // param2 (overrides default)
		values.NewString("extra1"), // variadic[0]
		values.NewString("extra2"), // variadic[1]
	}
	ctx.Temporaries = make(map[uint32]*values.Value)
	ctx.Temporaries[10] = values.NewString("default") // Default value for param2

	// 1. Receive first parameter
	op1Type, op2Type := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_UNUSED, opcodes.IS_TMP_VAR)
	recvInst1 := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     0, // Parameter 0
		Result:  0, // Store in temp var 0
	}

	err := vm.executeRecv(ctx, &recvInst1)
	if err != nil {
		t.Fatalf("First RECV failed: %v", err)
	}

	param1 := ctx.Temporaries[0]
	if param1.Data.(string) != "hello" {
		t.Errorf("Expected param1 'hello', got %v", param1.Data)
	}

	// 2. Receive second parameter with default
	op1Type2, op2Type2 := opcodes.EncodeOpTypes(opcodes.IS_UNUSED, opcodes.IS_TMP_VAR, opcodes.IS_TMP_VAR)
	recvInitInst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV_INIT,
		OpType1: op1Type2,
		OpType2: op2Type2,
		Op1:     1,  // Parameter 1
		Op2:     10, // Default from temp var 10
		Result:  1,  // Store in temp var 1
	}

	err = vm.executeRecvInit(ctx, &recvInitInst)
	if err != nil {
		t.Fatalf("RECV_INIT failed: %v", err)
	}

	param2 := ctx.Temporaries[1]
	if param2.Data.(int64) != 42 {
		t.Errorf("Expected param2 42, got %v", param2.Data)
	}

	// 3. Receive variadic parameters
	variadicInst := opcodes.Instruction{
		Opcode:  opcodes.OP_RECV_VARIADIC,
		OpType1: op1Type,
		OpType2: op2Type,
		Op1:     2, // Start from parameter 2
		Result:  2, // Store in temp var 2
	}

	err = vm.executeRecvVariadic(ctx, &variadicInst)
	if err != nil {
		t.Fatalf("RECV_VARIADIC failed: %v", err)
	}

	variadicArray := ctx.Temporaries[2]
	if variadicArray.ArrayCount() != 2 {
		t.Errorf("Expected 2 variadic parameters, got %d", variadicArray.ArrayCount())
	}

	extra1 := variadicArray.ArrayGet(values.NewInt(0))
	if extra1.Data.(string) != "extra1" {
		t.Errorf("Expected variadic[0] 'extra1', got %v", extra1.Data)
	}

	extra2 := variadicArray.ArrayGet(values.NewInt(1))
	if extra2.Data.(string) != "extra2" {
		t.Errorf("Expected variadic[1] 'extra2', got %v", extra2.Data)
	}
}
