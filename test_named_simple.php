<?php

function simpleTest($a, $b = "default") {
    return "a=$a, b=$b";
}

// Test positional - should work
echo simpleTest("first") . "\n";

// Test named - let's see what happens
echo simpleTest(a: "named_a") . "\n";