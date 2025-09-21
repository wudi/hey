<?php
class BaseException extends Exception {}
class DerivedException extends BaseException {}

echo "Starting inheritance test\n";

try {
    throw new DerivedException("Derived error");
} catch (BaseException $e) {
    echo "Caught as BaseException\n";
}

echo "Test completed\n";