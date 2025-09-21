<?php
try {
    echo "before throw\n";
    throw new Exception("test error");
    echo "after throw\n";  // This should not execute
} catch (Exception $e) {
    echo "caught: " . $e->getMessage() . "\n";
}
echo "after try-catch\n";