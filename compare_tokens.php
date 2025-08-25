#!/usr/bin/env php
<?php

/**
 * Compare PHP's native token_get_all() output with our Go parser
 */

function compareTokens($phpCode) {
    // Get PHP's native tokens
    $phpTokens = token_get_all($phpCode);
    
    echo "=== PHP NATIVE TOKENS ===\n";
    foreach ($phpTokens as $i => $token) {
        if (is_array($token)) {
            printf("%3d: %-30s %s\n", $i + 1, token_name($token[0]), json_encode($token[1]));
        } else {
            printf("%3d: %-30s %s\n", $i + 1, "SINGLE_CHAR", json_encode($token));
        }
    }
}

// Test with simple PHP code
$testCode = '<?php
function test(): void {
    echo "Hello World";
}
';

compareTokens($testCode);

?>