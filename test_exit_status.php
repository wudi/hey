<?php

echo "=== Testing exit() with status code ===\n";

// Test 2: Exit with status code
echo "Before exit with status 0\n";
exit(0);
echo "This should not be printed\n";

?>