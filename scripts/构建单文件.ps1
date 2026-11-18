#Requires -Version 7.0
param(
  [string]$OutputDirectory = (Join-Path $PSScriptRoot "../release/windows-amd64")
)

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
$assetsDirectory = [IO.Path]::GetFullPath((Join-Path $projectRoot "server/internal/portable/assets"))
$OutputDirectory = [IO.Path]::GetFullPath($OutputDirectory)
$previousCgo = $env:CGO_ENABLED
$previousGoos = $env:GOOS
$previousGoarch = $env:GOARCH
Push-Location $projectRoot
try {
  if ($assetsDirectory -ne [IO.Path]::GetFullPath("$projectRoot/server/internal/portable/assets")) {
    throw "生成资源目录不在预期位置。"
  }
  if (Test-Path -LiteralPath $assetsDirectory) {
    Remove-Item -LiteralPath $assetsDirectory -Recurse -Force
  }
  New-Item -ItemType Directory -Path $assetsDirectory, $OutputDirectory -Force | Out-Null
  Push-Location client
  try {
    pnpm exec node scripts/build-portable.mjs
    if ($LASTEXITCODE -ne 0) { throw "前端构建失败。" }
  } finally { Pop-Location }

  New-Item -ItemType Directory -Path "$assetsDirectory/models", "$assetsDirectory/knowledge/articles", "$assetsDirectory/seed" -Force | Out-Null
  $manifest = Get-Content docs/models-manifest.json -Raw -Encoding utf8 | ConvertFrom-Json
  foreach ($model in $manifest.files) {
    $source = Join-Path "client/public/models" $model.name
    if (!(Test-Path -LiteralPath $source) -or (Get-Item -LiteralPath $source).Length -ne $model.size) {
      throw "模型缺失或大小不符：$($model.name)。请先获取清单对应的模型资产。"
    }
    Copy-Item -LiteralPath $source -Destination "$assetsDirectory/models" -Force
  }
  Copy-Item server/knowledge/articles/*.md "$assetsDirectory/knowledge/articles" -Force
  Copy-Item server/seed/catalog.json, server/seed/annotations.json "$assetsDirectory/seed" -Force
  $template = Get-Content server/.env.example -Raw -Encoding utf8
  $template.Replace("APP_ENV=development", "APP_ENV=production").Replace("CORS_ORIGIN=http://localhost:5173", "CORS_ORIGIN=http://localhost:3001") | Set-Content "$assetsDirectory/environment.example" -Encoding utf8NoBOM

  $env:CGO_ENABLED = "0"
  $env:GOOS = "windows"
  $env:GOARCH = "amd64"
  Push-Location server
  try {
    go build -tags portable -trimpath -ldflags "-s -w" -o "$OutputDirectory/非遗华裳.exe" ./cmd/api
    if ($LASTEXITCODE -ne 0) { throw "Go 单文件构建失败。" }
  } finally { Pop-Location }
  Copy-Item scripts/发行版一键启动.bat "$OutputDirectory/一键启动.bat" -Force
  Copy-Item docs/免安装启动.md "$OutputDirectory/使用说明.md" -Force
  Write-Host "已生成：$OutputDirectory/非遗华裳.exe"
} finally {
  $env:CGO_ENABLED = $previousCgo
  $env:GOOS = $previousGoos
  $env:GOARCH = $previousGoarch
  Pop-Location
}
