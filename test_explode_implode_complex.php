<?php

echo "=== Testing Explode/Implode Complex Cases ===\n";

// Test explode with different delimiters
$csv = "John,25,Engineer";
$parts = explode(",", $csv);
echo "CSV explode: " . count($parts) . " parts - " . $parts[1] . "\n";

$path = "home/user/documents/file.txt";
$pathParts = explode("/", $path);
echo "Path explode: " . count($pathParts) . " parts - last: " . $pathParts[3] . "\n";

// Test implode with different data
$numbers = [1, 2, 3, 4, 5];
echo "Numbers implode: " . implode(" + ", $numbers) . "\n";

$words = ["Hello", "beautiful", "world"];
echo "Words implode: " . implode(" ", $words) . "\n";

// Test round trip (explode then implode)
$original = "apple,banana,orange";
$exploded = explode(",", $original);
$roundTrip = implode(",", $exploded);
echo "Round trip: " . ($original === $roundTrip ? "PASS" : "FAIL") . "\n";

echo "=== Complex test completed ===\n";

?>