<?php

trait Logger {
    public function log($message) {
        echo "Log: " . $message . "
";
    }
}

class MyClass {
    use Logger;

    public function doSomething() {
        $this->log("doing something");
    }
}

$obj = new MyClass();
$obj->doSomething();
