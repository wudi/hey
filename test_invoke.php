<?php

class InvokableClass {
    public function __invoke($param) {
        return "Invoked with: " . $param;
    }
}

$callable = new InvokableClass();
echo $callable("test parameter") . "
";
