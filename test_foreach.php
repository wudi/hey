<?php
function foo($n):array {
    $ret = [];
    for($i=0; $i<$n; $i++) {
        $ret[] = $i;
    }
    return $ret;
}

foreach(foo(5) as $v) {
    echo "$v
";
}
