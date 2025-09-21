<?php

class Test {
    public readonly string $prop;

    public function __construct() {
        echo "Setting readonly property...\n";
        $this->prop = "initial";
        echo "Readonly property set to: " . $this->prop . "\n";
    }

    public function modify() {
        echo "Trying to modify readonly property...\n";
        $this->prop = "modified";
        echo "Modified to: " . $this->prop . "\n";
    }
}

$obj = new Test();
$obj->modify();

?>