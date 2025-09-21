<?php
function simpleGenerator() {
    yield 1;
    yield 2;
    yield 3;
}

foreach (simpleGenerator() as $value) {
    echo "Value: $value\n";
}