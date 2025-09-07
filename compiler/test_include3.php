<?php
function included_function($param) {
    return "Function called with: " . $param;
}

$global_from_include = "Global variable";
echo "Include 3 executed\n";