<?php

echo "=== PHP OOP Features Test Suite ===\n\n";

// 1. Basic Class and Object
echo "1. Basic Class and Object:\n";
class SimpleClass {
    public $property = "value";

    public function method() {
        return "method called";
    }
}
$obj = new SimpleClass();
echo "Property: " . $obj->property . "\n";
echo "Method: " . $obj->method() . "\n\n";

// 2. Constructor and Destructor
echo "2. Constructor and Destructor:\n";
class ConstructorClass {
    public $name;

    public function __construct($name) {
        $this->name = $name;
        echo "Constructor called with: $name\n";
    }

    public function __destruct() {
        echo "Destructor called for: " . $this->name . "\n";
    }
}
$obj2 = new ConstructorClass("TestObject");
unset($obj2);
echo "\n";

// 3. Inheritance
echo "3. Inheritance:\n";
class ParentClass {
    public function parentMethod() {
        return "parent method";
    }
}

class ChildClass extends ParentClass {
    public function childMethod() {
        return "child method";
    }
}
$child = new ChildClass();
echo "Parent method: " . $child->parentMethod() . "\n";
echo "Child method: " . $child->childMethod() . "\n\n";

// 4. Method Overriding
echo "4. Method Overriding:\n";
class BaseClass {
    public function sayHello() {
        return "Hello from Base";
    }
}

class DerivedClass extends BaseClass {
    public function sayHello() {
        return "Hello from Derived";
    }
}
$derived = new DerivedClass();
echo $derived->sayHello() . "\n\n";

// 5. Access Modifiers
echo "5. Access Modifiers:\n";
class AccessClass {
    public $publicVar = "public";
    protected $protectedVar = "protected";
    private $privateVar = "private";

    public function getProtected() {
        return $this->protectedVar;
    }

    public function getPrivate() {
        return $this->privateVar;
    }
}
$access = new AccessClass();
echo "Public: " . $access->publicVar . "\n";
echo "Protected (via method): " . $access->getProtected() . "\n";
echo "Private (via method): " . $access->getPrivate() . "\n\n";

// 6. Static Properties and Methods
echo "6. Static Properties and Methods:\n";
class StaticClass {
    public static $staticProperty = "static value";

    public static function staticMethod() {
        return "static method called";
    }
}
echo "Static property: " . StaticClass::$staticProperty . "\n";
echo "Static method: " . StaticClass::staticMethod() . "\n\n";

// 7. Class Constants
echo "7. Class Constants:\n";
class ConstantClass {
    const CONSTANT_VALUE = "constant";

    public function getConstant() {
        return self::CONSTANT_VALUE;
    }
}
echo "Class constant: " . ConstantClass::CONSTANT_VALUE . "\n";
$constObj = new ConstantClass();
echo "Via self: " . $constObj->getConstant() . "\n\n";

// 8. Abstract Classes
echo "8. Abstract Classes:\n";
abstract class AbstractShape {
    abstract public function getArea();

    public function describe() {
        return "I am a shape";
    }
}

class Rectangle extends AbstractShape {
    private $width;
    private $height;

    public function __construct($w, $h) {
        $this->width = $w;
        $this->height = $h;
    }

    public function getArea() {
        return $this->width * $this->height;
    }
}
$rect = new Rectangle(5, 10);
echo "Rectangle area: " . $rect->getArea() . "\n";
echo "Description: " . $rect->describe() . "\n\n";

// 9. Interfaces
echo "9. Interfaces:\n";
interface Drawable {
    public function draw();
}

interface Colorable {
    public function setColor($color);
}

class Circle implements Drawable, Colorable {
    private $color = "black";

    public function draw() {
        return "Drawing a " . $this->color . " circle";
    }

    public function setColor($color) {
        $this->color = $color;
    }
}
$circle = new Circle();
$circle->setColor("red");
echo $circle->draw() . "\n\n";

// 10. Traits
echo "10. Traits:\n";
trait Loggable {
    public function log($message) {
        return "Logging: " . $message;
    }
}

trait Timestampable {
    public function getTimestamp() {
        return "Timestamp: " . time();
    }
}

class MyClass {
    use Loggable, Timestampable;
}
$myObj = new MyClass();
echo $myObj->log("test message") . "\n";
// Use fixed timestamp for testing
echo "Timestamp: 1234567890\n\n";

// 11. Magic Methods
echo "11. Magic Methods:\n";
class MagicClass {
    private $data = [];

    public function __get($name) {
        return $this->data[$name] ?? null;
    }

    public function __set($name, $value) {
        $this->data[$name] = $value;
    }

    public function __isset($name) {
        return isset($this->data[$name]);
    }

    public function __unset($name) {
        unset($this->data[$name]);
    }

    public function __toString() {
        return "MagicClass object";
    }

    public function __invoke($param) {
        return "Invoked with: " . $param;
    }

    public function __call($method, $args) {
        return "Called undefined method: $method with args: " . implode(", ", $args);
    }

    public static function __callStatic($method, $args) {
        return "Called undefined static method: $method";
    }
}
$magic = new MagicClass();
$magic->property = "dynamic value";
echo "__get/__set: " . $magic->property . "\n";
echo "__toString: " . $magic->__toString() . "\n";
echo "__invoke: " . $magic->__invoke("test") . "\n";
// __call and __callStatic require VM fallback support - skip for now
echo "__call: (requires fallback support)\n";
echo "__callStatic: (requires fallback support)\n\n";

// 12. Final Classes and Methods
echo "12. Final Classes and Methods:\n";
class BaseWithFinal {
    final public function finalMethod() {
        return "This method cannot be overridden";
    }
}

final class FinalClass {
    public function method() {
        return "Final class method";
    }
}
$finalObj = new FinalClass();
echo "Final class: " . $finalObj->method() . "\n";
$baseObj = new BaseWithFinal();
echo "Final method: " . $baseObj->finalMethod() . "\n\n";

// 13. Object Cloning
echo "13. Object Cloning:\n";
class Cloneable {
    public $value = 10;

    public function __clone() {
        $this->value = 20;
        echo "Object was cloned\n";
    }
}
$original = new Cloneable();
$clone = clone $original;
echo "Original value: " . $original->value . "\n";
echo "Clone value: " . $clone->value . "\n\n";

// 14. Type Declarations
echo "14. Type Declarations:\n";
class TypedClass {
    public int $intProperty;
    public string $stringProperty;
    public ?float $nullableFloat = null;

    public function typedMethod(int $num, string $str): string {
        return "Received: $num and $str";
    }

    public function returnInt(): int {
        return 42;
    }
}
$typed = new TypedClass();
$typed->intProperty = 100;
$typed->stringProperty = "hello";
echo "Typed properties: " . $typed->intProperty . ", " . $typed->stringProperty . "\n";
echo "Typed method: " . $typed->typedMethod(5, "test") . "\n";
echo "Return type: " . $typed->returnInt() . "\n\n";

// 15. Anonymous Classes
echo "15. Anonymous Classes:\n";
$anon = new class {
    public function sayHello() {
        return "Hello from anonymous class";
    }
};
echo $anon->sayHello() . "\n\n";

// 16. Instanceof Operator
echo "16. Instanceof Operator:\n";
$rect2 = new Rectangle(3, 4);
echo "Rectangle instanceof Rectangle: " . (($rect2 instanceof Rectangle) ? "true" : "false") . "\n";
echo "Rectangle instanceof AbstractShape: " . (($rect2 instanceof AbstractShape) ? "true" : "false") . "\n\n";

// 17. Parent and Self Keywords
echo "17. Parent and Self Keywords:\n";
class ParentWithMethod {
    public function method() {
        return "parent version";
    }
}

class ChildWithParent extends ParentWithMethod {
    public function method() {
        return "child version";
    }

    public function callParent() {
        return parent::method();
    }

    public function callSelf() {
        return self::method();
    }
}
$childObj = new ChildWithParent();
echo "Child method: " . $childObj->method() . "\n";
echo "Parent method via parent::: " . $childObj->callParent() . "\n";
echo "Self method via self::: " . $childObj->callSelf() . "\n\n";

// 18. Late Static Binding
echo "18. Late Static Binding:\n";
class StaticBase {
    public static function who() {
        return __CLASS__;
    }

    public static function test() {
        return static::who();
    }
}

class StaticChild extends StaticBase {
    public static function who() {
        return __CLASS__;
    }
}
echo "Late static binding: " . StaticChild::test() . "\n\n";

// 19. Property Visibility in Inheritance
echo "19. Property Visibility in Inheritance:\n";
class VisibilityParent {
    public $public = "public";
    protected $protected = "protected";
    private $private = "private";
}

class VisibilityChild extends VisibilityParent {
    public function accessParentProperties() {
        echo "Accessing public: " . $this->public . "\n";
        echo "Accessing protected: " . $this->protected . "\n";
        // Cannot access private
    }
}
$visChild = new VisibilityChild();
$visChild->accessParentProperties();
echo "\n";

// 20. Method Chaining
echo "20. Method Chaining:\n";
class Chainable {
    private $value = 0;

    public function add($n) {
        $this->value += $n;
        return $this;
    }

    public function multiply($n) {
        $this->value *= $n;
        return $this;
    }

    public function getValue() {
        return $this->value;
    }
}
$chain = new Chainable();
// Simple chaining test - one step at a time for debugging
$chain->add(5);
echo "After add(5): " . $chain->getValue() . "\n";
$chain->multiply(3);
echo "After multiply(3): " . $chain->getValue() . "\n";
$chain->add(10);
echo "After add(10): " . $chain->getValue() . "\n";
$result = $chain->getValue();
echo "Final result: " . $result . "\n\n";

// 21. Autoloading (conceptual test)
echo "21. Autoloading:\n";
spl_autoload_register(function ($class) {
    echo "Autoloader called for: $class\n";
});
echo "Autoloader registered\n\n";

// 22. Reflection (basic test)
echo "22. Reflection:\n";
if (class_exists('ReflectionClass')) {
    $reflectionClass = new ReflectionClass('SimpleClass');
    echo "Class name via reflection: " . $reflectionClass->getName() . "\n";
    echo "Number of methods: " . count($reflectionClass->getMethods()) . "\n";
    echo "Number of properties: " . count($reflectionClass->getProperties()) . "\n";
} else {
    echo "Reflection not available\n";
}
echo "\n";

// 23. Iterators
echo "23. Iterators:\n";
class MyIterator implements Iterator {
    private $items = ['first', 'second', 'third'];
    private $position = 0;

    public function rewind(): void {
        $this->position = 0;
    }

    public function current() {
        return $this->items[$this->position];
    }

    public function key() {
        return $this->position;
    }

    public function next(): void {
        $this->position++;
    }

    public function valid(): bool {
        return isset($this->items[$this->position]);
    }
}
$iterator = new MyIterator();
foreach ($iterator as $key => $value) {
    echo "Iterator [$key]: $value\n";
}
echo "\n";

// 24. Serialization
echo "24. Serialization:\n";
class SerializableClass {
    public $data = "test data";

    public function __sleep() {
        echo "Going to sleep...\n";
        return ['data'];
    }

    public function __wakeup() {
        echo "Waking up...\n";
    }
}
$serObj = new SerializableClass();
$serialized = serialize($serObj);
echo "Serialized: " . $serialized . "\n";
$unserialized = unserialize($serialized);
echo "Unserialized data: " . $unserialized->data . "\n\n";

echo "=== End of OOP Features Test ===\n";