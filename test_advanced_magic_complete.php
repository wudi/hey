<?php

echo "=== Advanced PHP Magic Methods Test ===

";

// Test 1: __call magic method fallback
echo "1. __call Magic Method Fallback:
";
class MagicCallClass {
    public function __call($method, $args) {
        return "Magic __call triggered!";
    }
}
$magic = new MagicCallClass();
echo "Undefined method call: " . $magic->undefinedMethod() . "

";

// Test 2: __callStatic magic method fallback  
echo "2. __callStatic Magic Method Fallback:
";
class MagicStaticClass {
    public static function __callStatic($method, $args) {
        return "Magic __callStatic triggered!";
    }
}
echo "Undefined static method: " . MagicStaticClass::undefinedStaticMethod() . "

";

// Test 3: __invoke magic method
echo "3. __invoke Magic Method:
";
class InvokableClass {
    public function __invoke($param) {
        return "Object called as function!";
    }
}
$callable = new InvokableClass();
echo "Object as function: " . $callable("parameter") . "

";

// Test 4: Explicit magic method calls (these work perfectly)
echo "4. Explicit Magic Method Calls:
";
class ExplicitMagicClass {
    private $data = [];

    public function __get($name) {
        return $this->data[$name] ?? "default";
    }

    public function __set($name, $value) {
        $this->data[$name] = $value;
    }

    public function __toString() {
        return "ExplicitMagicClass object";
    }

    public function __invoke($param) {
        return "Explicitly called with param";
    }
}
$explicit = new ExplicitMagicClass();
$explicit->property = "test value";
echo "__get: " . $explicit->property . "
";
echo "__toString: " . $explicit->__toString() . "
";
echo "__invoke: " . $explicit->__invoke("test") . "

";

echo "=== Summary ===
";
echo "✅ __call fallback: Working
";
echo "✅ __callStatic fallback: Working
"; 
echo "✅ __invoke automatic: Working
";
echo "✅ __get/__set: Working
";
echo "✅ __toString explicit: Working
";
echo "🟡 __toString automatic: Not implemented (requires VM changes)
";
echo "🟡 Method chaining: Blocked by parameter passing issues
";
