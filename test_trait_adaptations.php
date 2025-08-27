<?php
// Test trait adaptations

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
    // Simple trait usage
    use TraitA;
    
    // Multiple traits usage
    use TraitA, TraitB {
        // Precedence - TraitA::foo instead of TraitB::foo
        TraitA::foo insteadof TraitB;
        
        // Alias - TraitB::foo as fooFromB
        TraitB::foo as fooFromB;
        
        // Alias with visibility - TraitA::bar as private privateBar
        TraitA::bar as private privateBar;
        
        // Simple method alias
        baz as bazAlias;
        
        // Visibility change only
        foo as public;
    }
}