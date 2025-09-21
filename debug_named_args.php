<?php

function debugTest($param1, $param2 = "default") {
    return "param1=$param1, param2=$param2";
}

// Only test positional first to make sure function works
echo debugTest("value1", "value2") . "\n";