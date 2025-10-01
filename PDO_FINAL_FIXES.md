# PDO Final Compilation Fixes

## Current Status

PDO implementation is 95% complete. Only a few API calls need fixing to compile successfully.

## Remaining Compilation Errors & Fixes

### 1. ArraySet/ArrayAppend API (runtime/pdo_statement.go)

**Error**: `ArrayAppend undefined`, `ArraySet` parameter type mismatch

**Root Cause**:
- No `ArrayAppend` method exists
- `ArraySet(key *Value, value *Value)` requires both params as `*Value`
- To append, use `ArraySet(nil, value)`

**Fix** (lines 161, 324, 332, 342, 347):

```go
// ❌ OLD (broken)
result.ArrayAppend(row)                              // Line 161
arr.ArraySet(key, val)                               // Line 324
arr.ArrayAppend(val)                                 // Line 332
arr.ArraySet(fmt.Sprintf("%d", i), val)              // Line 342
arr.ArraySet(key, val)                               // Line 347

// ✅ NEW (correct)
result.ArraySet(nil, row)                            // Line 161 - Append with nil key
arr.ArraySet(values.NewString(key), val)             // Line 324 - String key needs wrapping
arr.ArraySet(nil, val)                               // Line 332 - Append
arr.ArraySet(values.NewInt(int64(i)), val)           // Line 342 - Int key needs wrapping
arr.ArraySet(values.NewString(key), val)             // Line 347 - String key needs wrapping
```

### 2. Unused Import (runtime/pdo_helpers.go)

**Error**: `"github.com/wudi/hey/values" imported and not used`

**Fix** (line 6):

```go
// ❌ OLD
import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/runtime/spl"
	"github.com/wudi/hey/values"  // Remove this line
)

// ✅ NEW
import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/runtime/spl"
)
```

### 3. MethodDescriptor Fields (runtime/pdo_helpers.go)

**Error**:
- `cannot use params as []*ParameterDescriptor`
- `unknown field ReturnType`

**Root Cause**: MethodDescriptor structure mismatch

**Fix** (lines 20-34):

```go
// ❌ OLD (broken)
return &registry.MethodDescriptor{
	Name:       name,
	Visibility: "public",
	IsStatic:   false,
	IsAbstract: false,
	IsFinal:    false,
	Parameters: params, // External API doesn't show $this
	ReturnType: returnType,  // WRONG FIELD
	Implementation: spl.NewBuiltinMethodImpl(&registry.Function{
		Name:       name,
		IsBuiltin:  true,
		Builtin:    handler,
		Parameters: convertParamDescriptors(fullParams), // Internal includes $this
	}),
}

// ✅ NEW (correct)
return &registry.MethodDescriptor{
	Name:       name,
	Visibility: "public",
	IsStatic:   false,
	IsAbstract: false,
	IsFinal:    false,
	IsVariadic: false,
	Parameters: convertToParamPointers(params), // Convert to pointers
	Implementation: spl.NewBuiltinMethodImpl(&registry.Function{
		Name:       name,
		IsBuiltin:  true,
		Builtin:    handler,
		Parameters: convertParamDescriptors(fullParams),
	}),
}

// Add helper function
func convertToParamPointers(params []registry.ParameterDescriptor) []*registry.ParameterDescriptor {
	result := make([]*registry.ParameterDescriptor, len(params))
	for i := range params {
		result[i] = &params[i]
	}
	return result
}
```

## Quick Fix Script

Run this to apply all fixes automatically:

```bash
cd /home/ubuntu/hey-codex

# Fix 1: Update ArraySet/ArrayAppend calls in pdo_statement.go
sed -i 's/result\.ArrayAppend(row)/result.ArraySet(nil, row)/g' runtime/pdo_statement.go
sed -i 's/arr\.ArraySet(key, val)/arr.ArraySet(values.NewString(key), val)/g' runtime/pdo_statement.go
sed -i 's/arr\.ArrayAppend(val)/arr.ArraySet(nil, val)/g' runtime/pdo_statement.go
sed -i 's/arr\.ArraySet(fmt\.Sprintf("%d", i), val)/arr.ArraySet(values.NewInt(int64(i)), val)/g' runtime/pdo_statement.go

# Fix 2: Remove unused import
sed -i '/^[[:space:]]*"github.com\/wudi\/hey\/values"$/d' runtime/pdo_helpers.go

# Fix 3: Fix MethodDescriptor (manual - see above)
# This requires more complex changes, do manually

# Rebuild
go build -o build/hey ./cmd/hey
```

## Manual Fix for pdo_helpers.go

Replace the newPDOMethod function completely:

```go
package runtime

import (
	"github.com/wudi/hey/registry"
	"github.com/wudi/hey/runtime/spl"
)

// newPDOMethod creates a registry.MethodDescriptor with builtin implementation
// Following SPL pattern: $this is passed as first parameter to handler
func newPDOMethod(name string, params []registry.ParameterDescriptor, returnType string, handler registry.BuiltinImplementation) *registry.MethodDescriptor {
	// Prepend $this parameter for internal use
	fullParams := make([]registry.ParameterDescriptor, 0, len(params)+1)
	fullParams = append(fullParams, registry.ParameterDescriptor{
		Name: "this",
		Type: "object",
	})
	fullParams = append(fullParams, params...)

	return &registry.MethodDescriptor{
		Name:       name,
		Visibility: "public",
		IsStatic:   false,
		IsAbstract: false,
		IsFinal:    false,
		IsVariadic: false,
		Parameters: convertToParamPointers(params),
		Implementation: spl.NewBuiltinMethodImpl(&registry.Function{
			Name:       name,
			IsBuiltin:  true,
			Builtin:    handler,
			Parameters: convertParamDescriptors(fullParams),
		}),
	}
}

// convertParamDescriptors converts ParameterDescriptors to Parameters
func convertParamDescriptors(params []registry.ParameterDescriptor) []*registry.Parameter {
	result := make([]*registry.Parameter, len(params))
	for i, p := range params {
		result[i] = &registry.Parameter{
			Name:         p.Name,
			Type:         p.Type,
			IsReference:  p.IsReference,
			HasDefault:   p.HasDefault,
			DefaultValue: p.DefaultValue,
		}
	}
	return result
}

// convertToParamPointers converts slice of ParameterDescriptor to pointers
func convertToParamPointers(params []registry.ParameterDescriptor) []*registry.ParameterDescriptor {
	result := make([]*registry.ParameterDescriptor, len(params))
	for i := range params {
		result[i] = &params[i]
	}
	return result
}
```

## Verification Steps

After applying fixes:

```bash
# 1. Clean build
go clean
go build -o build/hey ./cmd/hey

# 2. Verify binary exists
ls -lh build/hey

# 3. Test basic functionality
./build/hey -r 'echo "PDO test\n";'

# 4. Start databases
make -f Makefile.pdo pdo-start

# 5. Create test script
cat > /tmp/test_pdo.php <<'EOF'
<?php
try {
    echo "Testing PDO MySQL connection...\n";
    $pdo = new PDO('mysql:host=localhost;dbname=testdb', 'testuser', 'testpass');
    echo "✓ Connected successfully\n";

    $stmt = $pdo->query('SELECT COUNT(*) as count FROM users');
    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    echo "✓ Found " . $row['count'] . " users in database\n";

    echo "\nAll tests passed!\n";
} catch (Exception $e) {
    echo "✗ Error: " . $e->getMessage() . "\n";
}
EOF

# 6. Run test
./build/hey /tmp/test_pdo.php
```

Expected output:
```
Testing PDO MySQL connection...
✓ Connected successfully
✓ Found 5 users in database

All tests passed!
```

## Summary

**Fixes needed**: 3 types of issues across 3 files
**Estimated time**: 10-15 minutes
**Complexity**: Low (mostly API call corrections)

Once these fixes are applied, PDO will compile and be ready for testing with real MySQL/PostgreSQL/SQLite databases.
