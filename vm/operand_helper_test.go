package vm

import (
	"testing"

	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

func TestOperandReader_ReadOperand1(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewString("test_constant"),
		values.NewInt(42),
	})

	// Set up temporary variable
	frame.setTemp(0, values.NewString("temp_value"))

	// Set up local variable
	frame.setLocal(1, values.NewString("local_value"))

	tests := []struct {
		name     string
		opType1  byte
		op1      uint32
		expected *values.Value
		wantErr  bool
	}{
		{
			name:     "read constant",
			opType1:  byte(opcodes.IS_CONST) << 4,
			op1:      0,
			expected: values.NewString("test_constant"),
			wantErr:  false,
		},
		{
			name:     "read temp var",
			opType1:  byte(opcodes.IS_TMP_VAR) << 4,
			op1:      0,
			expected: values.NewString("temp_value"),
			wantErr:  false,
		},
		{
			name:     "read local var",
			opType1:  byte(opcodes.IS_VAR) << 4,
			op1:      1,
			expected: values.NewString("local_value"),
			wantErr:  false,
		},
		{
			name:     "read unused",
			opType1:  byte(opcodes.IS_UNUSED) << 4,
			op1:      0,
			expected: values.NewNull(),
			wantErr:  false,
		},
		{
			name:     "constant out of range",
			opType1:  byte(opcodes.IS_CONST) << 4,
			op1:      99,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &opcodes.Instruction{
				OpType1: tt.opType1,
				Op1:     tt.op1,
			}

			reader := NewOperandReader(ctx, frame, inst)
			result, err := reader.ReadOperand1()

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadOperand1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !result.Equal(tt.expected) {
				t.Errorf("ReadOperand1() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestOperandReader_WriteResult(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, nil)

	tests := []struct {
		name      string
		opType2   byte
		result    uint32
		value     *values.Value
		wantErr   bool
		checkFunc func() bool
	}{
		{
			name:    "write to temp var",
			opType2: (byte(opcodes.IS_TMP_VAR) << 4),
			result:  5,
			value:   values.NewString("temp_result"),
			wantErr: false,
			checkFunc: func() bool {
				return frame.getTemp(5).Equal(values.NewString("temp_result"))
			},
		},
		{
			name:    "write to local var",
			opType2: (byte(opcodes.IS_VAR) << 4),
			result:  3,
			value:   values.NewInt(123),
			wantErr: false,
			checkFunc: func() bool {
				return frame.getLocal(3).Equal(values.NewInt(123))
			},
		},
		{
			name:    "write to unused",
			opType2: (byte(opcodes.IS_UNUSED) << 4),
			result:  0,
			value:   values.NewString("ignored"),
			wantErr: false,
			checkFunc: func() bool {
				return true // Should not error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &opcodes.Instruction{
				OpType2: tt.opType2,
				Result:  tt.result,
			}

			reader := NewOperandReader(ctx, frame, inst)
			err := reader.WriteResult(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("WriteResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.checkFunc() {
				t.Errorf("WriteResult() did not write value correctly")
			}
		})
	}
}

func TestDecodeOperands(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, []*values.Value{
		values.NewString("const1"),
		values.NewInt(100),
	})

	// Create instruction with two constant operands
	inst := &opcodes.Instruction{
		OpType1: (byte(opcodes.IS_CONST) << 4) | byte(opcodes.IS_CONST),
		Op1:     0,
		Op2:     1,
	}

	operands, err := DecodeOperands(ctx, frame, inst)
	if err != nil {
		t.Fatalf("DecodeOperands() error = %v", err)
	}

	if !operands.Op1.Equal(values.NewString("const1")) {
		t.Errorf("DecodeOperands() Op1 = %v, want %v", operands.Op1, values.NewString("const1"))
	}

	if !operands.Op2.Equal(values.NewInt(100)) {
		t.Errorf("DecodeOperands() Op2 = %v, want %v", operands.Op2, values.NewInt(100))
	}
}

func TestWriteResult(t *testing.T) {
	ctx := NewExecutionContext()
	frame := newCallFrame("test", nil, nil, nil)

	inst := &opcodes.Instruction{
		OpType2: (byte(opcodes.IS_TMP_VAR) << 4),
		Result:  7,
	}

	result := values.NewString("test_result")
	err := WriteResult(ctx, frame, inst, result)
	if err != nil {
		t.Fatalf("WriteResult() error = %v", err)
	}

	if !frame.getTemp(7).Equal(result) {
		t.Errorf("WriteResult() did not write correctly")
	}
}