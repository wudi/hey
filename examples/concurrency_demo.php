<?php
// Demo of Go-style concurrency in PHP using go() function and WaitGroup class

echo "=== Go-style Concurrency Demo ===\n";

// Example 1: Simple goroutine with go() function
echo "\n1. Simple goroutine example:\n";

$closure = function() use ($var1, $var2) {
    echo "Hello from goroutine! var1=$var1, var2=$var2\n";
    return "goroutine result";
};

$var1 = "test";
$var2 = 42;

// Start a goroutine
$goroutine = go($closure) use($var1, $var2);
echo "Goroutine started: $goroutine\n";

// Example 2: Using WaitGroup to coordinate multiple goroutines
echo "\n2. WaitGroup coordination example:\n";

$wg = new WaitGroup();

// Add 3 work items
$wg->Add(3);

// Simulate starting 3 goroutines
for ($i = 1; $i <= 3; $i++) {
    $worker = function() use ($i, $wg) {
        echo "Worker $i starting...\n";
        // Simulate some work
        sleep(1);
        echo "Worker $i finished!\n";
        $wg->Done();
    };
    
    go($worker) use($i, $wg);
}

echo "Waiting for all workers to complete...\n";
$wg->Wait();
echo "All workers completed!\n";

// Example 3: More complex coordination
echo "\n3. Complex coordination example:\n";

$wg2 = new WaitGroup();
$results = array();

$wg2->Add(5);

for ($i = 1; $i <= 5; $i++) {
    $task = function() use ($i, $wg2, &$results) {
        $result = $i * $i; // Square the number
        echo "Task $i: result = $result\n";
        $results[$i] = $result;
        $wg2->Done();
    };
    
    go($task) use($i, $wg2, $results);
}

echo "Processing tasks...\n";
$wg2->Wait();
echo "All tasks completed!\n";

echo "Results: ";
print_r($results);

echo "\n=== Demo Complete ===\n";
?>