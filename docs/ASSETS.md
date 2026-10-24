# 3D 模型资产

41 个 GLB 模型约 230 MB，体积不适合放入 Git 历史。模型通过 GitHub Release 分发，仓库内保留清单、校验脚本和下载脚本。

## 一键获取

在仓库根目录执行：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\fetch-models.ps1
```

首次下载约 230 MB；脚本会解压到 `client/public/models/`，并按 `docs/models-manifest.json` 逐个校验 SHA-256。已存在完整模型时会跳过；需要重新下载时加 `-Force`。

## 下载地址

- Release: https://github.com/zwy819/feiyi-huashang/releases/tag/models-v1
- 资产: `feiyi-models-v1.zip`
- 清单: [`docs/models-manifest.json`](models-manifest.json)

## 校验

清单包含每个文件的 `name`、`size` 和 `sha256`。如需单独校验：

```powershell
$manifest = Get-Content .\docs\models-manifest.json -Raw -Encoding UTF8 | ConvertFrom-Json
$manifest.files | ForEach-Object {
  $p = Join-Path .\client\public\models $_.name
  [pscustomobject]@{
    Name = $_.name
    Exists = Test-Path $p
    HashOK = (Test-Path $p) -and ((Get-FileHash $p -Algorithm SHA256).Hash.ToLower() -eq $_.sha256)
  }
}
```

## 使用边界

- 模型是依据服饰资料制作的数字复原/艺术化展示，不是文物级三维扫描。
- 模型的廓形、配色、纹样位置和材质可能经过简化或补全。
- 不得把模型用于文物鉴定、交易估值或民族身份判断。
- 模型生成工具、参考图片与作者信息应在正式提交前补充到 `THIRD_PARTY_NOTICES.md`。

## 为什么不用 Git LFS

Git LFS 有存储与月流量配额，评审反复克隆时容易耗尽；Release 资产对公开仓库访问更稳定，也避免把 230 MB 二进制写进提交历史。仓库仍保留完整代码历史。
