<?php

echo "=== Testing JSON with Complex Data ===\n";

// Test nested arrays and objects
$complex = [
    'user' => [
        'name' => 'John',
        'age' => 30,
        'active' => true,
        'scores' => [85, 92, 78]
    ],
    'settings' => [
        'theme' => 'dark',
        'notifications' => false
    ]
];

$json = json_encode($complex);
echo "Complex json_encode:\n$json\n\n";

// Test decoding back
$decoded = json_decode($json, true);
echo "Decoded user name: " . $decoded['user']['name'] . "\n";
echo "Decoded first score: " . $decoded['user']['scores'][0] . "\n";
echo "Decoded theme: " . $decoded['settings']['theme'] . "\n";

// Test edge cases
echo "\nEdge cases:\n";
echo "json_encode(null): " . json_encode(null) . "\n";
echo "json_encode(true): " . json_encode(true) . "\n";
echo "json_encode(false): " . json_encode(false) . "\n";
echo "json_encode(123): " . json_encode(123) . "\n";
echo "json_encode(3.14): " . json_encode(3.14) . "\n";

// Test json_decode edge cases
echo "json_decode('null'): " . (json_decode('null') === null ? "null" : "not null") . "\n";

echo "=== Complex JSON test completed ===\n";

?>