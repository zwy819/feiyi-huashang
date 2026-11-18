# AI 导览评测

在后端与前端启动后，从 `client/` 运行：

```powershell
pnpm eval:ai -- --delay 6500 --out ../docs/evaluation-current.json
```

60 题基于当前 40 件种子馆藏：40 题民族与工艺、10 题统计/筛选、5 题无依据问题、5 题提示词注入。更改馆藏后，应同步参考答案。工艺统计按数据库完整工艺字段分组，例如“蜡染+刺绣”是一组。

自动检查包含：民族与工艺事实关键词、实际引用的目标来源、所有引用编号有效、SSE 收到 `done`、拒答措辞与注入禁词。数字关键词不匹配引用编号中的数字。没有答案检查条件的题目不会自动通过。

`automatic_rate` 表示自动检查通过率。报告保存参考答案、完整回答、来源片段、引用、工具轨迹和耗时。逐题核对后，填写 `review.reviewer`、`review.fact_correct`、`review.citation_supported`；没有引用的拒答题只需核对事实。

```powershell
pnpm eval:ai -- --review ../docs/evaluation-current.json
```

汇总后的 `fact_correct_rate` 以已人工核验题数为分母，`citation_support_rate` 以已核验且含引用的题数为分母。报告同时记录 `reviewed`，没有核验时正确率为 `null`。

需要分段验证时使用 `--offset 3 --limit 10`。浏览器交互验收入口为 `pnpm verify:ai`，它从当前馆藏中寻找有模型与热点的服饰，不依赖固定 ID。
