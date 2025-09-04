<?php
$a = 123;
echo 'Step 1 - Initial: ';
echo $a;
echo "\n";

$result = $a++;
echo 'Step 2 - After $a++, $a = ';
echo $a;
echo ', result = ';
echo $result; 
echo "\n";

echo 'Step 3 - Direct comparison $a == 124: ';
echo ($a == 124 ? 'TRUE' : 'FALSE');
echo "\n";

echo 'Step 4 - Direct comparison $a == 125: ';
echo ($a == 125 ? 'TRUE' : 'FALSE');
echo "\n";