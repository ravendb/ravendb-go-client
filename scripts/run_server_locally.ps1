#!/usr/bin/env pwsh
# on mac install powershell: https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell-core-on-macos?view=powershell-6

# --ServerUrl.Tcp=" + tcpServerURL,

.\RavenDB\Server\Raven.Server.exe --ServerUrl=http://localhost:8080 --RunInMemory=true --License.Eula.Accepted=true --Setup.Mode=None --non-interactive
