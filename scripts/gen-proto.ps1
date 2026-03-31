$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$protoRoot = Join-Path $repoRoot "proto"
$moduleName = "eco-knock-be-embedded"

if (-not (Test-Path $protoRoot)) {
    throw "proto directory not found: $protoRoot"
}

$protoc = Get-Command protoc.exe -ErrorAction SilentlyContinue
if (-not $protoc) {
    throw "protoc.exe not found in PATH"
}

$protocGenGo = Get-Command protoc-gen-go.exe -ErrorAction SilentlyContinue
if (-not $protocGenGo) {
    throw "protoc-gen-go.exe not found in PATH"
}

$protocGenGoGrpc = Get-Command protoc-gen-go-grpc.exe -ErrorAction SilentlyContinue
if (-not $protocGenGoGrpc) {
    throw "protoc-gen-go-grpc.exe not found in PATH"
}

$protoFiles = Get-ChildItem -Path $protoRoot -Recurse -Filter *.proto
if (-not $protoFiles) {
    throw "no .proto files found under $protoRoot"
}

Push-Location $repoRoot
try {
    foreach ($protoFile in $protoFiles) {
        $relativePath = Resolve-Path -Relative $protoFile.FullName
        Write-Host "Generating $relativePath"

        & $protoc.Source `
            "--go_out=." `
            "--go_opt=module=$moduleName" `
            "--go-grpc_out=." `
            "--go-grpc_opt=module=$moduleName" `
            $relativePath

        if ($LASTEXITCODE -ne 0) {
            throw "protoc failed for $relativePath"
        }
    }
}
finally {
    Pop-Location
}
