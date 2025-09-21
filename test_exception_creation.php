<?php
echo "Creating exception...\n";
$e = new Exception("test message");
echo "Exception created\n";
echo "Message: " . $e->getMessage() . "\n";