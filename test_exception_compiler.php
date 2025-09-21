<?php
// Test case to ensure our implementation matches PHP behavior
class BaseException extends Exception {}
class DerivedException extends BaseException {}
class OtherException extends Exception {}

// Test 1: Basic type matching
try {
    throw new DerivedException("Test message");
} catch (OtherException $e) {
    echo "Wrong catch 1\n";
} catch (BaseException $e) {
    echo "Correct: Caught as BaseException\n";
} catch (Exception $e) {
    echo "Wrong catch 2\n";
}

// Test 2: Order matters
try {
    throw new DerivedException("Test 2");
} catch (DerivedException $e) {
    echo "Correct: Caught as DerivedException\n";
} catch (BaseException $e) {
    echo "Wrong: Should not reach here\n";
}

// Test 3: No match with finally
try {
    try {
        throw new OtherException("No match");
    } catch (DerivedException $e) {
        echo "Won't catch\n";
    } finally {
        echo "Finally executes\n";
    }
} catch (Exception $e) {
    echo "Outer catch: " . get_class($e) . "\n";
}