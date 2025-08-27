<?php
// 基本匿名类
$obj1 = new class {};

// 带构造参数的匿名类
$obj2 = new class($arg1, $arg2) {};

// 继承和实现接口的匿名类
$obj3 = new class extends BaseClass implements Interface1, Interface2 {};

// 带修饰符的匿名类
$obj4 = new final class {};
$obj5 = new readonly class {};

// 复杂的匿名类
$obj6 = new class($param) extends Parent implements I1, I2 {
    private $property;
    
    public function __construct($param) {
        $this->property = $param;
    }
    
    public function method() {
        return $this->property;
    }
};

// 带属性的匿名类
$obj7 = new #[Attribute] class {};