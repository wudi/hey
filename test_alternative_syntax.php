<?php
// 测试 Alternative 语法

// Alternative if
if ($condition):
    echo "true";
elseif ($another):
    echo "another";
else:
    echo "false";
endif;

// Alternative while
while ($x < 10):
    echo $x;
    $x++;
endwhile;

// Alternative for
for ($i = 0; $i < 10; $i++):
    echo $i;
endfor;

// Alternative foreach
foreach ($array as $item):
    echo $item;
endforeach;

// Alternative switch
switch ($value):
    case 1:
        echo "one";
        break;
    case 2:
        echo "two";
        break;
    default:
        echo "other";
endswitch;
?>