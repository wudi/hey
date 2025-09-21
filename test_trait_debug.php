<?php

trait DebugTrait {
    public function debugThis() {
        echo "In debugThis method\n";

        // Try to access $this directly
        $thisVar = $this;
        echo "Assigned \$this to variable: ";
        var_dump($thisVar);

        // Try to access a property
        echo "Property access test: " . $this->prop . "\n";

        // Try to modify property
        $this->prop = "modified";
        echo "Modified property: " . $this->prop . "\n";

        return $thisVar;
    }
}

class DebugClass {
    use DebugTrait;
    public $prop = "initial";
}

echo "=== Debug trait $this ===\n";
$obj = new DebugClass();
$result = $obj->debugThis();
echo "Method returned: ";
var_dump($result);