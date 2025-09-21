<?php

echo "=== Testing File I/O Functions ===\n";

// Test file_put_contents
$filename = "test_file.txt";
$content = "Hello, World!\nThis is a test file.";

echo "Writing content to file...\n";
$bytes_written = file_put_contents($filename, $content);
echo "Bytes written: " . $bytes_written . "\n";

// Test file_get_contents
echo "Reading content from file...\n";
$read_content = file_get_contents($filename);
echo "Content read: " . $read_content . "\n";

// Test with non-existent file
echo "Reading non-existent file...\n";
$false_result = file_get_contents("nonexistent.txt");
echo "Result for non-existent: " . ($false_result === false ? "false" : $false_result) . "\n";

// Clean up
unlink($filename);

echo "=== File I/O test completed ===\n";

?>