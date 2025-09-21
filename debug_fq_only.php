<?php
class GlobalClass {
    public function test() {
        return "Global";
    }
}

// Only test the fully qualified case
$globalFQ = new \GlobalClass();
echo "About to call method\n";
echo $globalFQ->test() . "\n";