<?php

class MagicClass {
    public static function __callStatic($method, $args) {
        return "Static magic method called!";
    }
}

echo MagicClass::testStatic() . "
";
