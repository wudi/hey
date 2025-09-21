<?php

echo "=== Generator Delegation Debug ===\n";

function innerGen() {
    echo "Inner: yielding 'hello'\n";
    yield 'hello';
    echo "Inner: yielding 'world'\n";
    yield 'world';
    echo "Inner: done\n";
}

function outerGen() {
    echo "Outer: starting\n";
    yield from innerGen();
    echo "Outer: finished\n";
}

echo "Testing generator delegation:\n";
foreach (outerGen() as $value) {
    echo "Got: $value\n";
}

?>