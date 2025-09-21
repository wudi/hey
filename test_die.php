<?php

echo "=== Testing die() Function (alias of exit) ===\n";

// Test 4: die() is an alias of exit()
echo "Before die with message\n";
die("Die message from script");
echo "This should not be printed\n";

?>