<?php

echo "=== Testing exit() with message ===\n";

// Test 3: Exit with string message
echo "Before exit with message\n";
exit("Exit message from script");
echo "This should not be printed\n";

?>