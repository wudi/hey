package vm

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/wudi/php-parser/compiler/opcodes"
	"github.com/wudi/php-parser/compiler/values"
)

// DebugLevel represents different debugging levels
type DebugLevel int

const (
	DebugLevelNone DebugLevel = iota
	DebugLevelBasic
	DebugLevelDetailed
	DebugLevelVerbose
)

// Debugger provides advanced debugging capabilities for the VM
type Debugger struct {
	Level           DebugLevel
	Output          io.Writer
	BreakPoints     map[int]bool       // IP -> enabled
	WatchVariables  map[string]bool    // Variable names to watch
	InstructionLog  []InstructionTrace // Execution trace
	CallStack       []CallTrace        // Function call trace
	MaxTraceEntries int                // Maximum trace entries to keep
	StepMode        bool               // Step-by-step execution
	ProfilerEnabled bool               // Enable profiling
	ProfileData     *ProfileData       // Profiling data
}

// InstructionTrace represents a traced instruction execution
type InstructionTrace struct {
	Timestamp   time.Time
	IP          int
	Instruction opcodes.Instruction
	OpcodeName  string
	StackSize   int
	Variables   map[uint32]*values.Value // Copy of variables at execution
	Duration    time.Duration            // Execution time for this instruction
}

// CallTrace represents a function call trace entry
type CallTrace struct {
	Timestamp    time.Time
	FunctionName string
	Arguments    []*values.Value
	ReturnValue  *values.Value
	Duration     time.Duration
	CallDepth    int
}

// ProfileData contains profiling information
type ProfileData struct {
	FunctionProfiles map[string]*FunctionProfile
	InstructionTimes map[string]time.Duration
	HotPaths         []HotPath
	MemoryProfile    *MemoryProfile
}

// FunctionProfile contains profiling data for a function
type FunctionProfile struct {
	Name            string
	CallCount       uint64
	TotalTime       time.Duration
	AverageTime     time.Duration
	MinTime         time.Duration
	MaxTime         time.Duration
	MemoryAllocated uint64
	MemoryPeak      uint64
}

// HotPath represents a frequently executed code path
type HotPath struct {
	StartIP        int
	EndIP          int
	ExecutionCount uint64
	TotalTime      time.Duration
	Instructions   []string
}

// MemoryProfile tracks memory usage patterns
type MemoryProfile struct {
	AllocationsPerType   map[string]uint64
	DeallocationsPerType map[string]uint64
	PeakUsagePerType     map[string]uint64
	LeakDetection        map[string]uint64 // Potential memory leaks
}

// NewDebugger creates a new debugger instance
func NewDebugger(level DebugLevel, output io.Writer) *Debugger {
	if output == nil {
		output = os.Stderr
	}

	return &Debugger{
		Level:           level,
		Output:          output,
		BreakPoints:     make(map[int]bool),
		WatchVariables:  make(map[string]bool),
		InstructionLog:  make([]InstructionTrace, 0, 1000),
		CallStack:       make([]CallTrace, 0, 100),
		MaxTraceEntries: 10000,
		ProfileData: &ProfileData{
			FunctionProfiles: make(map[string]*FunctionProfile),
			InstructionTimes: make(map[string]time.Duration),
			HotPaths:         make([]HotPath, 0, 100),
			MemoryProfile: &MemoryProfile{
				AllocationsPerType:   make(map[string]uint64),
				DeallocationsPerType: make(map[string]uint64),
				PeakUsagePerType:     make(map[string]uint64),
				LeakDetection:        make(map[string]uint64),
			},
		},
	}
}

// SetBreakpoint sets a breakpoint at the specified instruction pointer
func (d *Debugger) SetBreakpoint(ip int) {
	d.BreakPoints[ip] = true
	if d.Level >= DebugLevelBasic {
		fmt.Fprintf(d.Output, "[DEBUGGER] Breakpoint set at IP %d\n", ip)
	}
}

// RemoveBreakpoint removes a breakpoint
func (d *Debugger) RemoveBreakpoint(ip int) {
	delete(d.BreakPoints, ip)
	if d.Level >= DebugLevelBasic {
		fmt.Fprintf(d.Output, "[DEBUGGER] Breakpoint removed at IP %d\n", ip)
	}
}

// WatchVariable adds a variable to watch list
func (d *Debugger) WatchVariable(varName string) {
	d.WatchVariables[varName] = true
	if d.Level >= DebugLevelBasic {
		fmt.Fprintf(d.Output, "[DEBUGGER] Watching variable: %s\n", varName)
	}
}

// TraceInstruction records the execution of an instruction
func (d *Debugger) TraceInstruction(ip int, inst *opcodes.Instruction, ctx *ExecutionContext, duration time.Duration) {
	if d.Level < DebugLevelDetailed {
		return
	}

	// Create a copy of variables for the trace
	varsCopy := make(map[uint32]*values.Value)
	for k, v := range ctx.Variables {
		if v != nil {
			varsCopy[k] = &values.Value{Type: v.Type, Data: v.Data}
		}
	}

	trace := InstructionTrace{
		Timestamp:   time.Now(),
		IP:          ip,
		Instruction: *inst,
		OpcodeName:  inst.Opcode.String(),
		StackSize:   ctx.SP,
		Variables:   varsCopy,
		Duration:    duration,
	}

	d.InstructionLog = append(d.InstructionLog, trace)

	// Keep trace log size manageable
	if len(d.InstructionLog) > d.MaxTraceEntries {
		d.InstructionLog = d.InstructionLog[1000:] // Remove oldest 1000 entries
	}

	// Update profiler data
	if d.ProfilerEnabled {
		d.ProfileData.InstructionTimes[inst.Opcode.String()] += duration
	}

	if d.Level >= DebugLevelVerbose {
		d.printInstructionTrace(&trace, ctx)
	}
}

// printInstructionTrace prints instruction trace information
func (d *Debugger) printInstructionTrace(trace *InstructionTrace, ctx *ExecutionContext) {
	fmt.Fprintf(d.Output, "[TRACE] IP:%04d %-20s SP:%d Duration:%v\n",
		trace.IP,
		trace.OpcodeName,
		trace.StackSize,
		trace.Duration,
	)

	// Print operands
	fmt.Fprintf(d.Output, "        Op1:%d/%s Op2:%d/%s Result:%d/%s\n",
		trace.Instruction.Op1,
		opcodes.DecodeOpType1(trace.Instruction.OpType1).String(),
		trace.Instruction.Op2,
		opcodes.DecodeOpType2(trace.Instruction.OpType1).String(),
		trace.Instruction.Result,
		opcodes.DecodeResultType(trace.Instruction.OpType2).String(),
	)

	// Print watched variables
	for varName := range d.WatchVariables {
		if slot, exists := ctx.VarSlotNames[0]; exists && slot == varName {
			if val, exists := ctx.Variables[0]; exists {
				fmt.Fprintf(d.Output, "        WATCH %s = %s\n", varName, val.String())
			}
		}
	}
}

// TraceFunctionCall records a function call
func (d *Debugger) TraceFunctionCall(functionName string, args []*values.Value, callDepth int) {
	if d.Level < DebugLevelBasic {
		return
	}

	trace := CallTrace{
		Timestamp:    time.Now(),
		FunctionName: functionName,
		Arguments:    args,
		CallDepth:    callDepth,
	}

	d.CallStack = append(d.CallStack, trace)

	if d.Level >= DebugLevelDetailed {
		indent := strings.Repeat("  ", callDepth)
		argStrs := make([]string, len(args))
		for i, arg := range args {
			if arg != nil {
				argStrs[i] = arg.String()
			} else {
				argStrs[i] = "null"
			}
		}
		fmt.Fprintf(d.Output, "[CALL] %s-> %s(%s)\n", indent, functionName, strings.Join(argStrs, ", "))
	}
}

// TraceFunctionReturn records a function return
func (d *Debugger) TraceFunctionReturn(functionName string, returnValue *values.Value, callDepth int) {
	if d.Level < DebugLevelBasic {
		return
	}

	// Update the last call trace with return information
	if len(d.CallStack) > 0 {
		lastTrace := &d.CallStack[len(d.CallStack)-1]
		lastTrace.ReturnValue = returnValue
		lastTrace.Duration = time.Since(lastTrace.Timestamp)

		// Update profiler
		if d.ProfilerEnabled {
			d.updateFunctionProfile(functionName, lastTrace.Duration)
		}
	}

	if d.Level >= DebugLevelDetailed {
		indent := strings.Repeat("  ", callDepth)
		retVal := "null"
		if returnValue != nil {
			retVal = returnValue.String()
		}
		fmt.Fprintf(d.Output, "[RETURN] %s<- %s = %s\n", indent, functionName, retVal)
	}
}

// updateFunctionProfile updates profiling data for a function
func (d *Debugger) updateFunctionProfile(functionName string, duration time.Duration) {
	profile, exists := d.ProfileData.FunctionProfiles[functionName]
	if !exists {
		profile = &FunctionProfile{
			Name:    functionName,
			MinTime: duration,
			MaxTime: duration,
		}
		d.ProfileData.FunctionProfiles[functionName] = profile
	}

	profile.CallCount++
	profile.TotalTime += duration
	profile.AverageTime = time.Duration(int64(profile.TotalTime) / int64(profile.CallCount))

	if duration < profile.MinTime {
		profile.MinTime = duration
	}
	if duration > profile.MaxTime {
		profile.MaxTime = duration
	}
}

// ShouldBreak checks if execution should break at current IP
func (d *Debugger) ShouldBreak(ip int) bool {
	return d.BreakPoints[ip] || d.StepMode
}

// PrintStack prints the current call stack
func (d *Debugger) PrintStack() {
	fmt.Fprintf(d.Output, "\n=== Call Stack ===\n")
	for i := len(d.CallStack) - 1; i >= 0; i-- {
		trace := d.CallStack[i]
		fmt.Fprintf(d.Output, "#%d %s called at %v\n",
			len(d.CallStack)-1-i,
			trace.FunctionName,
			trace.Timestamp.Format("15:04:05.000"))
	}
	fmt.Fprintf(d.Output, "==================\n\n")
}

// PrintVariables prints current variable state
func (d *Debugger) PrintVariables(ctx *ExecutionContext) {
	fmt.Fprintf(d.Output, "\n=== Variables ===\n")

	for slot, value := range ctx.Variables {
		varName := fmt.Sprintf("var_%d", slot)
		if name, exists := ctx.VarSlotNames[slot]; exists {
			varName = name
		}
		fmt.Fprintf(d.Output, "  %s = %s\n", varName, value.String())
	}

	fmt.Fprintf(d.Output, "\n=== Temporaries ===\n")
	for slot, value := range ctx.Temporaries {
		fmt.Fprintf(d.Output, "  tmp_%d = %s\n", slot, value.String())
	}

	fmt.Fprintf(d.Output, "==================\n\n")
}

// GenerateReport generates a comprehensive debugging report
func (d *Debugger) GenerateReport() string {
	var report strings.Builder

	report.WriteString("=== VM Debugger Report ===\n\n")

	// Execution summary
	report.WriteString(fmt.Sprintf("Instructions traced: %d\n", len(d.InstructionLog)))
	report.WriteString(fmt.Sprintf("Function calls: %d\n", len(d.CallStack)))
	report.WriteString(fmt.Sprintf("Breakpoints: %d\n", len(d.BreakPoints)))
	report.WriteString(fmt.Sprintf("Watched variables: %d\n\n", len(d.WatchVariables)))

	// Function profiling
	if d.ProfilerEnabled && len(d.ProfileData.FunctionProfiles) > 0 {
		report.WriteString("=== Function Performance ===\n")
		for name, profile := range d.ProfileData.FunctionProfiles {
			report.WriteString(fmt.Sprintf("Function: %s\n", name))
			report.WriteString(fmt.Sprintf("  Calls: %d\n", profile.CallCount))
			report.WriteString(fmt.Sprintf("  Total time: %v\n", profile.TotalTime))
			report.WriteString(fmt.Sprintf("  Average time: %v\n", profile.AverageTime))
			report.WriteString(fmt.Sprintf("  Min/Max time: %v / %v\n\n", profile.MinTime, profile.MaxTime))
		}
	}

	// Instruction timing
	if len(d.ProfileData.InstructionTimes) > 0 {
		report.WriteString("=== Instruction Timing ===\n")
		for opcode, totalTime := range d.ProfileData.InstructionTimes {
			report.WriteString(fmt.Sprintf("%s: %v\n", opcode, totalTime))
		}
		report.WriteString("\n")
	}

	// Recent trace (last 10 instructions)
	if len(d.InstructionLog) > 0 {
		report.WriteString("=== Recent Instruction Trace ===\n")
		start := len(d.InstructionLog) - 10
		if start < 0 {
			start = 0
		}

		for i := start; i < len(d.InstructionLog); i++ {
			trace := d.InstructionLog[i]
			report.WriteString(fmt.Sprintf("IP:%04d %-15s Duration:%v\n",
				trace.IP, trace.OpcodeName, trace.Duration))
		}
	}

	return report.String()
}
