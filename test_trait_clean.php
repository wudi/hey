<?php
trait TraitA {
    public function foo() {
        echo "TraitA::foo";
    }
    
    public function bar() {
        echo "TraitA::bar";
    }
}

trait TraitB {
    public function foo() {
        echo "TraitB::foo";
    }
    
    public function baz() {
        echo "TraitB::baz";
    }
}

class TestClass {
    use TraitA;
    
    use TraitA, TraitB {
        TraitA::foo insteadof TraitB;
        TraitB::foo as fooFromB;
        TraitA::bar as private privateBar;
        baz as bazAlias;
        foo as public;
    }
}