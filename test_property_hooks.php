<?php

class PropertyHooksExample {
    // Simple get hook with arrow syntax
    public string $name {
        get => $this->firstName . ' ' . $this->lastName;
    }
    
    // Set hook with parameter and arrow syntax
    public string $email {
        set(string $value) => strtolower($value);
    }
    
    // Both get and set hooks with block syntax
    public int $age {
        get {
            return $this->birthYear ? date('Y') - $this->birthYear : 0;
        }
        
        set(int $value) {
            if ($value < 0 || $value > 150) {
                throw new InvalidArgumentException('Invalid age');
            }
            $this->birthYear = date('Y') - $value;
        }
    }
    
    // Reference get hook
    public array $data {
        &get => $this->internalData;
    }
    
    private string $firstName;
    private string $lastName;
    private int $birthYear;
    private array $internalData = [];
}