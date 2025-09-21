<?php

class MagicClass {
    public function __call($method, $args) {
        return "method=" . $method . " args=array";
    }
}

$magic = new MagicClass();
echo $magic->test() . "
";
