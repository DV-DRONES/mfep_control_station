$ErrorActionPreference = "Stop"

$projectDirectory = $PSScriptRoot
$buildDirectory = Join-Path $projectDirectory "build"
$classesDirectory = Join-Path $buildDirectory "classes"
$jarPath = Join-Path $buildDirectory "wifi.jar"
$sourcePath = Join-Path $projectDirectory "src\wifi\Wifi.java"

New-Item -ItemType Directory -Force -Path $classesDirectory | Out-Null

& javac -Xlint:all -d $classesDirectory $sourcePath
if ($LASTEXITCODE -ne 0) {
    throw "javac failed with exit code $LASTEXITCODE"
}

& jar --create --file $jarPath --main-class wifi.Wifi -C $classesDirectory .
if ($LASTEXITCODE -ne 0) {
    throw "jar failed with exit code $LASTEXITCODE"
}

Write-Output "Built reusable library: $jarPath"
