<?php

class MagicClass {
    public function __call($method, $args) {
        return "Magic method called!";
    }
}

$magic = new MagicClass();
echo $magic->test() . "
";
