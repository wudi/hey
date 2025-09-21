<?php

// Test edge cases for named arguments

// Test 1: All required parameters with named args
function requiredParams($first, $second, $third) {
    return "$first-$second-$third";
}

echo requiredParams(third: "3rd", first: "1st", second: "2nd") . "\n";

// Test 2: Error case - missing required parameter
function missingRequired($required, $optional = "default") {
    return "$required-$optional";
}

try {
    echo missingRequired(optional: "test") . "\n"; // Should error - missing $required
} catch (Error $e) {
    echo "Caught error: " . $e->getMessage() . "\n";
}

// Test 3: Mixed positional and named (positional first)
function mixedArgs($a, $b = "b_default", $c = "c_default") {
    return "a=$a, b=$b, c=$c";
}

echo mixedArgs("pos_a", c: "named_c") . "\n";

// Test 4: Function with many parameters, only some named
function manyParams($p1, $p2, $p3 = "3", $p4 = "4", $p5 = "5") {
    return "p1=$p1, p2=$p2, p3=$p3, p4=$p4, p5=$p5";
}

echo manyParams("1", "2", p5: "named5", p4: "named4") . "\n";