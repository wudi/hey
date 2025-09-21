<?php

class MagicClass {
    public function __toString() {
        return "MagicClass object";
    }
}

$magic = new MagicClass();
echo "Direct call: " . $magic->__toString() . "
";
echo "Automatic call: " . $magic . "
";
