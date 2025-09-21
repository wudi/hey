<?php

echo "=== Testing getcwd() Function ===\n";

// Test 1: Get current working directory
$cwd = getcwd();
echo "Current directory: " . $cwd . "\n";
echo "Is string: " . (is_string($cwd) ? 'true' : 'false') . "\n";
echo "Not empty: " . (!empty($cwd) ? 'true' : 'false') . "\n";

// Test 2: Verify it's an absolute path
echo "Starts with /: " . (substr($cwd, 0, 1) === '/' ? 'true' : 'false') . "\n";

// Test 3: Change directory and verify getcwd updates
$original_dir = getcwd();
echo "Original dir: " . $original_dir . "\n";

// Create a temporary directory for testing
$temp_dir = sys_get_temp_dir() . '/test_getcwd_' . time();
if (mkdir($temp_dir)) {
    echo "Created temp dir: " . $temp_dir . "\n";

    // Change to the temp directory
    if (chdir($temp_dir)) {
        $new_cwd = getcwd();
        echo "After chdir: " . $new_cwd . "\n";
        echo "Changed correctly: " . ($new_cwd === $temp_dir ? 'true' : 'false') . "\n";

        // Change back to original directory
        chdir($original_dir);
        $restored_cwd = getcwd();
        echo "After restore: " . $restored_cwd . "\n";
        echo "Restored correctly: " . ($restored_cwd === $original_dir ? 'true' : 'false') . "\n";
    } else {
        echo "Failed to change directory\n";
    }

    // Clean up temp directory
    rmdir($temp_dir);
    echo "Cleaned up temp dir\n";
} else {
    echo "Failed to create temp directory\n";
}

// Test 4: Return type consistency
$cwd1 = getcwd();
$cwd2 = getcwd();
echo "Consistent results: " . ($cwd1 === $cwd2 ? 'true' : 'false') . "\n";

// Test 5: No parameters accepted
// getcwd() doesn't accept any parameters in PHP
$final_cwd = getcwd();
echo "Final CWD type: " . gettype($final_cwd) . "\n";

// Test 6: Path exists and is a directory
if (file_exists($final_cwd) && is_dir($final_cwd)) {
    echo "Path exists and is directory: true\n";
} else {
    echo "Path exists and is directory: false\n";
}

echo "=== getcwd() test completed ===\n";

?>