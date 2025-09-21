<?php

trait VariadicTrait {
    public function sum(...$numbers) {
        echo "sum called with " . count($numbers) . " arguments\n";
        $total = 0;
        foreach ($numbers as $num) {
            echo "Adding: " . $num . "\n";
            $total += $num;
        }
        return $total;
    }
}

class Calculator {
    use VariadicTrait;
}

$calc = new Calculator();
echo "Result: " . $calc->sum(1, 2, 3) . "\n";