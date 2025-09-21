<?php
class TestException extends Exception {}

echo "Testing exception class creation\n";
$e = new Exception("test");
echo "Exception created: " . get_class($e) . "\n";

$e2 = new TestException("test2");
echo "TestException created: " . get_class($e2) . "\n";