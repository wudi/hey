<?php

echo "=== Testing exit() with error code ===\n";

// Test 5: Exit with error code (non-zero)
echo "Before exit with error code 1\n";
exit(1);
echo "This should not be printed\n";

?>