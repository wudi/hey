<?php
// Test 1: Basic exception type checking
echo "Test 1: Basic exception type checking\n";
class MyException extends Exception {}
class OtherException extends Exception {}

try {
    throw new MyException("My error");
} catch (OtherException $e) {
    echo "Caught OtherException\n";
} catch (MyException $e) {
    echo "Caught MyException: " . $e->getMessage() . "\n";
}

// Test 2: Exception hierarchy
echo "\nTest 2: Exception hierarchy\n";
class BaseException extends Exception {}
class DerivedException extends BaseException {}

try {
    throw new DerivedException("Derived error");
} catch (BaseException $e) {
    echo "Caught BaseException (actually Derived): " . get_class($e) . "\n";
}

// Test 3: Multiple catch blocks with order matters
echo "\nTest 3: Order matters\n";
try {
    throw new DerivedException("Error");
} catch (DerivedException $e) {
    echo "Caught DerivedException first\n";
} catch (BaseException $e) {
    echo "Should not reach here\n";
}

// Test 4: Interface catching
echo "\nTest 4: Interface catching\n";
interface ThrowableInterface {}
class CustomThrowable extends Exception implements ThrowableInterface {}

try {
    throw new CustomThrowable("Custom error");
} catch (ThrowableInterface $e) {
    echo "Caught via interface: " . get_class($e) . "\n";
}

// Test 5: Built-in exceptions
echo "\nTest 5: Built-in exceptions\n";
try {
    $arr = [];
    echo $arr['nonexistent'];
} catch (ErrorException $e) {
    echo "Caught ErrorException\n";
} catch (Exception $e) {
    echo "Caught generic Exception\n";
} catch (Error $e) {
    echo "Caught Error\n";
}

// Test 6: No matching catch
echo "\nTest 6: No matching catch with finally\n";
try {
    try {
        throw new MyException("Uncaught");
    } catch (OtherException $e) {
        echo "Won't catch\n";
    } finally {
        echo "Finally block executes\n";
    }
} catch (MyException $e) {
    echo "Caught by outer try: " . $e->getMessage() . "\n";
}

// Test 7: Throwable interface (PHP 7+)
echo "\nTest 7: Throwable interface\n";
try {
    throw new Exception("Standard exception");
} catch (Throwable $t) {
    echo "Caught Throwable: " . get_class($t) . "\n";
}

// Test 8: Type errors
echo "\nTest 8: Type errors\n";
function requiresInt(int $x) {
    return $x * 2;
}
try {
    requiresInt("not an int");
} catch (TypeError $e) {
    echo "Caught TypeError\n";
}

echo "\nAll tests completed\n";