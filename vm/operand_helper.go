package vm

import (
	"github.com/wudi/hey/opcodes"
	"github.com/wudi/hey/values"
)

// OperandSet holds decoded operands for instruction execution
type OperandSet struct {
	Op1    *values.Value
	Op2    *values.Value
	Result *values.Value
}

// OperandReader handles reading operands from instructions
type OperandReader struct {
	ctx   *ExecutionContext
	frame *CallFrame
	inst  *opcodes.Instruction
}

// NewOperandReader creates a new operand reader for the given context
func NewOperandReader(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) *OperandReader {
	return &OperandReader{
		ctx:   ctx,
		frame: frame,
		inst:  inst,
	}
}

// ReadOperand1 reads the first operand from the instruction
func (r *OperandReader) ReadOperand1() (*values.Value, error) {
	opType := opcodes.DecodeOpType1(r.inst.OpType1)
	return r.readOperand(opType, r.inst.Op1)
}

// ReadOperand2 reads the second operand from the instruction
func (r *OperandReader) ReadOperand2() (*values.Value, error) {
	opType := opcodes.DecodeOpType2(r.inst.OpType1)
	return r.readOperand(opType, r.inst.Op2)
}

// ReadBothOperands reads both operands from the instruction
func (r *OperandReader) ReadBothOperands() (*values.Value, *values.Value, error) {
	op1, err := r.ReadOperand1()
	if err != nil {
		return nil, nil, err
	}

	op2, err := r.ReadOperand2()
	if err != nil {
		return nil, nil, err
	}

	return op1, op2, nil
}

// WriteResult writes the result to the instruction's result location
func (r *OperandReader) WriteResult(result *values.Value) error {
	resultType := opcodes.DecodeResultType(r.inst.OpType2)
	return r.writeOperand(resultType, r.inst.Result, result)
}

// readOperand is the internal method for reading operands
func (r *OperandReader) readOperand(opType opcodes.OpType, operand uint32) (*values.Value, error) {
	switch opType {
	case opcodes.IS_UNUSED:
		return values.NewNull(), nil
	case opcodes.IS_CONST:
		if int(operand) >= len(r.frame.Constants) {
			return nil, NewConstantError(operand, len(r.frame.Constants))
		}
		return r.frame.Constants[operand], nil
	case opcodes.IS_TMP_VAR:
		return r.frame.getTemp(operand), nil
	case opcodes.IS_VAR, opcodes.IS_CV:
		return r.frame.getLocal(operand), nil
	default:
		return nil, NewOperandError(opType, "read")
	}
}

// writeOperand is the internal method for writing operands
func (r *OperandReader) writeOperand(opType opcodes.OpType, operand uint32, value *values.Value) error {
	switch opType {
	case opcodes.IS_UNUSED:
		return nil
	case opcodes.IS_TMP_VAR:
		r.frame.setTemp(operand, value)
		r.ctx.setTemporary(operand, value)
		return nil
	case opcodes.IS_VAR, opcodes.IS_CV:
		r.frame.setLocal(operand, value)
		if globalName, ok := r.frame.globalSlotName(operand); ok {
			r.ctx.bindGlobalValue(globalName, value)
		}
		r.ctx.recordAssignment(r.frame, operand, value)
		if name, ok := r.frame.SlotNames[operand]; ok {
			r.ctx.setVariable(name, value)
		}
		return nil
	case opcodes.IS_CONST:
		if r.frame == nil || int(operand) >= len(r.frame.Constants) {
			return NewConstantError(operand, len(r.frame.Constants))
		}
		constVal := r.frame.Constants[operand]
		if r.ctx != nil && r.ctx.currentClass != nil && constVal.IsString() {
			propName := constVal.ToString()
			if prop, ok := r.ctx.currentClass.Properties[propName]; ok {
				prop.Default = copyValue(value)
				if prop.IsStatic {
					if r.ctx.currentClass.StaticProps == nil {
						r.ctx.currentClass.StaticProps = make(map[string]*values.Value)
					}
					r.ctx.currentClass.StaticProps[propName] = copyValue(value)
				}
				return nil
			}
		}
		return NewOperandError(opType, "write")
	default:
		return NewOperandError(opType, "write")
	}
}

// DecodeOperands is a convenience function that creates an OperandReader and reads common operands
func DecodeOperands(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction) (*OperandSet, error) {
	reader := NewOperandReader(ctx, frame, inst)

	op1, err := reader.ReadOperand1()
	if err != nil {
		return nil, err
	}

	op2, err := reader.ReadOperand2()
	if err != nil {
		return nil, err
	}

	return &OperandSet{
		Op1: op1,
		Op2: op2,
	}, nil
}

// WriteResult is a convenience function for writing results
func WriteResult(ctx *ExecutionContext, frame *CallFrame, inst *opcodes.Instruction, result *values.Value) error {
	reader := NewOperandReader(ctx, frame, inst)
	return reader.WriteResult(result)
}