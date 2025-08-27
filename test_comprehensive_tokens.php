<?php
require 'test.php';
require_once 'config.php';
include 'header.php';
include_once 'footer.php';

static $globalVar = 1;
abstract class TestClass {
    static function test() {
        return static::$instance;
    }
}

\namespace\test();
?>