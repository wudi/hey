<?php

class NonInvokableClass {
}

$obj = new NonInvokableClass();
echo $obj("test");
