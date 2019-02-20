#!/usr/bin/env pwsh
# on mac install powershell: https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell-core-on-macos?view=powershell-6

Function FormatElapsedTime($ts) {
    $elapsedTime = ""

    if ( $ts.Minutes -gt 0 ) {
        $elapsedTime = [string]::Format( "{0:00} min. {1:00}.{2:00} sec.", $ts.Minutes, $ts.Seconds, $ts.Milliseconds / 10 );
    }
    else {
        $elapsedTime = [string]::Format( "{0:00}.{1:00} sec.", $ts.Seconds, $ts.Milliseconds / 10 );
    }

    if ($ts.Hours -eq 0 -and $ts.Minutes -eq 0 -and $ts.Seconds -eq 0) {
        $elapsedTime = [string]::Format("{0:00} ms.", $ts.Milliseconds);
    }

    if ($ts.Milliseconds -eq 0) {
        $elapsedTime = [string]::Format("{0} ms", $ts.TotalMilliseconds);
    }

    return $elapsedTime
}

$Env:LOG_HTTP_REQUEST_SUMMARY = "true"
$Env:LOG_FAILED_HTTP_REQUESTS = "true"
$Env:LOG_ALL_REQUESTS = "true"
#$Env:ENABLE_FAILING_TESTS = "true"
#$Env:ENABLE_FLAKY_TESTS = "true"

# go test -tags for_tests -covermode=atomic -coverprofile=coverage.txt

$sw = [Diagnostics.Stopwatch]::StartNew()
# -parallel 1 to disable parallel execution of tests
go test -tags for_tests -parallel 1 -race -timeout 20m -v ./tests
Start-Sleep -s 3
$sw.Stop()
FormatElapsedTime $sw.Elapsed
