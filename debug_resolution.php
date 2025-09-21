<?php
class GlobalClass {
    public function test() {
        return "Global";
    }
}

// Only test the basic case first
$global = new GlobalClass();
echo $global->test() . "\n";