<?php

echo <<<HELP
Synopsis:
    php run-tests.php [options] [files] [directories]

Options:
    -j<workers> Run up to <workers> simultaneous testing processes in parallel for
                quicker testing on systems with multiple logical processors.
                Note that this is experimental feature.

    -l <file>   Read the testfiles to be executed from <file>. After the test
                has finished all failed tests are written to the same <file>.
                If the list is empty and no further test is specified then
                all tests are executed (same as: -r <file> -w <file>).

    -r <file>   Read the testfiles to be executed from <file>.

    -w <file>   Write a list of all failed tests to <file>.

    -a <file>   Same as -w but append rather then truncating <file>.

    -W <file>   Write a list of all tests and their result status to <file>.

    -c <file>   Look for php.ini in directory <file> or use <file> as ini.

    -n          Pass -n option to the php binary (Do not use a php.ini).

    -d foo=bar  Pass -d option to the php binary (Define INI entry foo
                with value 'bar').

    -g          Comma separated list of groups to show during test run
                (possible values: PASS, FAIL, XFAIL, XLEAK, SKIP, BORK, WARN, LEAK, REDIRECT).

    -m          Test for memory leaks with Valgrind (equivalent to -M memcheck).

    -M <tool>   Test for errors with Valgrind tool.

    -p <php>    Specify PHP executable to run.

    -P          Use PHP_BINARY as PHP executable to run (default).

    -q          Quiet, no user interaction (same as environment NO_INTERACTION).

    -s <file>   Write output to <file>.

    -x          Sets 'SKIP_SLOW_TESTS' environmental variable.

    --offline   Sets 'SKIP_ONLINE_TESTS' environmental variable.

    --verbose
    -v          Verbose mode.

    --help
    -h          This Help.

    --temp-source <sdir>  --temp-target <tdir> [--temp-urlbase <url>]
                Write temporary files to <tdir> by replacing <sdir> from the
                filenames to generate with <tdir>. In general you want to make
                <sdir> the path to your source files and <tdir> some patch in
                your web page hierarchy with <url> pointing to <tdir>.

    --keep-[all|php|skip|clean]
                Do not delete 'all' files, 'php' test file, 'skip' or 'clean'
                file.

    --set-timeout <n>
                Set timeout for individual tests, where <n> is the number of
                seconds. The default value is 60 seconds, or 300 seconds when
                testing for memory leaks.

    --context <n>
                Sets the number of lines of surrounding context to print for diffs.
                The default value is 3.

    --show-[all|php|skip|clean|exp|diff|out|mem]
                Show 'all' files, 'php' test file, 'skip' or 'clean' file. You
                can also use this to show the output 'out', the expected result
                'exp', the difference between them 'diff' or the valgrind log
                'mem'. The result types get written independent of the log format,
                however 'diff' only exists when a test fails.

    --show-slow <n>
                Show all tests that took longer than <n> milliseconds to run.

    --no-clean  Do not execute clean section if any.

    --color
    --no-color  Do/Don't colorize the result type in the test result.

    --progress
    --no-progress  Do/Don't show the current progress.

    --repeat [n]
                Run the tests multiple times in the same process and check the
                output of the last execution (CLI SAPI only).

    --bless     Bless failed tests using scripts/dev/bless_tests.php.

HELP;

/**
 * One function to rule them all, one function to find them, one function to
 * bring them all and in the darkness bind them.
 * This is the entry point and exit point überfunction. It contains all the
 * code that was previously found at the top level. It could and should be
 * refactored to be smaller and more manageable.
 */
function main(): void
{
}