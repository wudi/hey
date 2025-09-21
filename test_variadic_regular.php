<?php
function sum(...$numbers) {
    echo "Received " . count($numbers) . " parameters\n";
    $total = 0;
    foreach ($numbers as $num) {
        echo "Adding: " . $num . "\n";
        $total += $num;
    }
    return $total;
}

echo "Result 1: " . sum(1, 2, 3, 4, 5) . "\n";
echo "Result 2: " . sum(10, 20) . "\n";