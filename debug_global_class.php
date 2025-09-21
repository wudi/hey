<?php
class GlobalClass {
    public function test() {
        return "Global";
    }
}

$global = new GlobalClass();
echo $global->test() . "\n";

$globalFQ = new \GlobalClass();
echo $globalFQ->test() . "\n";