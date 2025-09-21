<?php
class TestClass {
}

$obj = new TestClass();
echo "Basic works\n";

$obj2 = new \TestClass();
echo "FQ works\n";