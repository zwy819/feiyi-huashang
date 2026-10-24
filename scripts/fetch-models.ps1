<#
.SYNOPSIS
  下载并校验“非遗华裳”的 3D 模型资产。

.DESCRIPTION
  模型约 230 MB，不放入 Git 历史。脚本从公开 GitHub Release 下载压缩包，
  解压到 client/public/models/，并按 docs/models-manifest.json 逐个校验 SHA-256。

.EXAMPLE
  powershell -ExecutionPolicy Bypass -File .\scripts\fetch-models.ps1
  powershell -ExecutionPolicy Bypass -File .\scripts\fetch-models.ps1 -Force
#>
param(
    [switch]$Force
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$modelsDir = Join-Path $repoRoot 'client\public\models'
$manifestPath = Join-Path $repoRoot 'docs\models-manifest.json'
$tag = 'models-v1'
$assetName = 'feiyi-models-v1.zip'
$downloadUrl = "https://github.com/zwy819/feiyi-huashang/releases/download/$tag/$assetName"

if (-not (Test-Path $manifestPath)) {
    throw "未找到模型清单：$manifestPath"
}
$manifest = Get-Content $manifestPath -Raw -Encoding UTF8 | ConvertFrom-Json
$expected = [int]$manifest.count

$existing = @(Get-ChildItem $modelsDir -Filter '*.glb' -ErrorAction SilentlyContinue)
if (-not $Force -and $existing.Count -ge $expected) {
    Write-Host "[跳过] 已存在 $($existing.Count) 个 GLB（期望 $expected 个）。使用 -Force 可重新下载。"
    exit 0
}

New-Item -ItemType Directory -Path $modelsDir -Force | Out-Null
$zipPath = Join-Path $env:TEMP $assetName
$extractDir = Join-Path $env:TEMP 'feiyi-models-extract'

Write-Host "[1/3] 下载模型资产（约 230 MB）..."
if ($Force -and (Test-Path $zipPath)) { Remove-Item $zipPath -Force }
if (-not (Test-Path $zipPath) -or (Get-Item $zipPath).Length -lt 100MB) {
    & curl.exe -L --fail --retry 5 --retry-all-errors --connect-timeout 20 --max-time 1800 -o $zipPath $downloadUrl
    if ($LASTEXITCODE -ne 0) {
        throw "下载失败（curl exit $LASTEXITCODE）。请检查网络，或手动下载：$downloadUrl"
    }
}

Write-Host "[2/3] 解压到临时目录..."
if (Test-Path $extractDir) { Remove-Item $extractDir -Recurse -Force }
Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force
Get-ChildItem $extractDir -Filter '*.glb' | Copy-Item -Destination $modelsDir -Force
Remove-Item (Join-Path $extractDir 'README.txt') -Force -ErrorAction SilentlyContinue
Remove-Item $extractDir -Recurse -Force

Write-Host "[3/3] 校验 SHA-256..."
$failed = @()
foreach ($item in $manifest.files) {
    $path = Join-Path $modelsDir $item.name
    if (-not (Test-Path $path)) {
        $failed += "$($item.name): 缺失"
        continue
    }
    $actual = (Get-FileHash $path -Algorithm SHA256).Hash.ToLower()
    if ($actual -ne $item.sha256.ToLower()) {
        $failed += "$($item.name): 哈希不匹配"
    }
}
if ($failed.Count -gt 0) {
    $failed | ForEach-Object { Write-Host "  ✗ $_" -ForegroundColor Red }
    throw "模型校验失败。可删除 client/public/models/ 后使用 -Force 重试。"
}

$totalMB = [math]::Round((Get-ChildItem $modelsDir -Filter '*.glb' | Measure-Object Length -Sum).Sum / 1MB, 1)
Write-Host "[完成] $($manifest.count) 个模型已就绪，共 $totalMB MB。" -ForegroundColor Green
Write-Host "启动前端后即可在详情页查看 3D 模型。"
