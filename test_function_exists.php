<?php

echo "=== Testing function_exists ===\n";

echo "function_exists('strlen'): " . (function_exists('strlen') ? "true" : "false") . "\n";
echo "function_exists('trim'): " . (function_exists('trim') ? "true" : "false") . "\n";
echo "function_exists('nonexistent'): " . (function_exists('nonexistent') ? "true" : "false") . "\n";

echo "=== Test completed ===\n";

?>