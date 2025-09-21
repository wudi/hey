<?php

echo "=== Testing JSON Functions ===\n";

// Test JSON encoding
try {
    $data = ['name' => 'John', 'age' => 30, 'active' => true];
    echo "json_encode: " . json_encode($data) . "\n";
} catch (Error $e) {
    echo "json_encode NOT IMPLEMENTED\n";
}

// Test JSON decoding
try {
    $json = '{"name":"Jane","age":25}';
    $decoded = json_decode($json, true);
    echo "json_decode: " . $decoded['name'] . "\n";
} catch (Error $e) {
    echo "json_decode NOT IMPLEMENTED\n";
}

echo "=== JSON test completed ===\n";

?>