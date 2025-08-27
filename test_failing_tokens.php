<?php
require 'test.php';
require_once 'config.php';
?>
<?php
static $var = 1;
abstract class TestClass {
    static function test() {}
}
\namespace\test();