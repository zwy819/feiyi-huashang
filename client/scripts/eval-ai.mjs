import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { checkAnswer } from "./eval-checks.mjs";

const args = process.argv.slice(2);
function arg(name, fallback) {
  const i = args.indexOf(`--${name}`);
  return i >= 0 && args[i + 1] ? args[i + 1] : fallback;
}
const reviewPath = arg("review", null);
if (reviewPath) {
  const file = path.resolve(reviewPath);
  const reviewedReport = JSON.parse(await readFile(file, "utf8"));
  const reviewed = reviewedReport.items.filter(item => item.review?.reviewer?.trim()
    && typeof item.review.fact_correct === "boolean"
    && (!item.cited_ids.length || typeof item.review.citation_supported === "boolean"));
  const withCitations = reviewed.filter(item => item.cited_ids.length > 0);
  reviewedReport.reviewed = reviewed.length;
  reviewedReport.fact_correct_rate = reviewed.length ? reviewed.filter(item => item.review.fact_correct).length / reviewed.length : null;
  reviewedReport.citation_support_rate = withCitations.length ? withCitations.filter(item => item.review.citation_supported).length / withCitations.length : null;
  await writeFile(file, JSON.stringify(reviewedReport, null, 2) + "\n", "utf8");
  console.log(JSON.stringify({ reviewed: reviewed.length, fact_correct_rate: reviewedReport.fact_correct_rate, citation_support_rate: reviewedReport.citation_support_rate }, null, 2));
  process.exit(0);
}
const baseURL = arg("url", process.env.APP_URL || "http://localhost:3001");
const outPath = path.resolve(arg("out", "eval-report.json"));
const delay = Number(arg("delay", "0"));
const limit = Number(arg("limit", "0"));
const offset = Number(arg("offset", "0"));
const dataset = JSON.parse(await readFile(new URL("../../server/knowledge/eval/questions.json", import.meta.url), "utf8"));
const questions = limit > 0 ? dataset.questions.slice(offset, offset + limit) : dataset.questions.slice(offset);

function parseSSE(text) {
  const result = { sources: [], content: "", action: null, errors: [], done: null, tools: [] };
  for (const block of text.replace(/\r\n/g, "\n").split("\n\n")) {
    let event = "message";
    const data = [];
    for (const line of block.split("\n")) {
      if (line.startsWith("event:")) event = line.slice(6).trim();
      else if (line.startsWith("data:")) data.push(line.slice(5).trimStart());
    }
    if (!data.length) continue;
    let payload;
    try { payload = JSON.parse(data.join("\n")); } catch { continue; }
    if (event === "meta") { result.model = payload.model; result.mock_mode = payload.mock_mode; }
    else if (event === "sources") result.sources = payload;
    else if (event === "tool") result.tools.push(payload);
    else if (event === "delta") result.content += payload.text || "";
    else if (event === "action") result.action = payload;
    else if (event === "done") { result.done = payload; result.content = payload.content || result.content; }
    else if (event === "error") result.errors.push(payload);
  }
  return result;
}

const report = {
  started_at: new Date().toISOString(),
  dataset_version: dataset.version,
  offset,
  base_url: baseURL,
  total: questions.length,
  review_instructions: "逐条核对参考答案、完整回答和所引来源；填写 review.fact_correct 与 review.citation_supported，并记录 reviewer。automatic_ok 表示自动检查通过。",
  items: [],
};
for (const item of questions) {
  const started = Date.now();
  let parsed = { sources: [], content: "", action: null, errors: [], done: null, tools: [] };
  let status = 0;
  try {
    const response = await fetch(baseURL + "/api/ai/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json", Accept: "text/event-stream" },
      body: JSON.stringify({ message: item.question, history: [], context: { page: "home" } }),
      signal: AbortSignal.timeout(120000),
    });
    status = response.status;
    parsed = parseSSE(await response.text());
  } catch (error) {
    parsed.errors.push({ message: String(error) });
  }
  const checks = checkAnswer(item, parsed, status);
  report.items.push({
    id: item.id,
    category: item.category,
    question: item.question,
    reference: item.reference,
    expected_facts: item.expected_facts || [],
    expected_terms: item.expected_terms || [],
    expected_source_groups: item.expected_source_groups || [],
    ...checks,
    status,
    model: parsed.model || null,
    mock_mode: parsed.mock_mode ?? null,
    usage: parsed.done?.usage || null,
    completed: !!parsed.done,
    answer: parsed.content,
    sources: parsed.sources,
    action: parsed.action,
    tools: parsed.tools,
    errors: parsed.errors,
    latency_ms: Date.now() - started,
    review: { reviewer: "", fact_correct: null, citation_supported: null, notes: "" },
  });
  console.log(`${checks.automatic_ok ? "AUTO_PASS" : "AUTO_FAIL"} ${item.id} ${item.category} (${Date.now() - started}ms)`);
  await writeFile(outPath, JSON.stringify(report, null, 2) + "\n", "utf8");
  if (delay > 0) await new Promise(resolve => setTimeout(resolve, delay));
}
report.finished_at = new Date().toISOString();
report.automatic_passed = report.items.filter(item => item.automatic_ok).length;
report.automatic_rate = report.total ? report.automatic_passed / report.total : 0;
report.reviewed = 0;
report.fact_correct_rate = null;
report.citation_support_rate = null;
await writeFile(outPath, JSON.stringify(report, null, 2) + "\n", "utf8");
console.log(JSON.stringify({ out: outPath, automatic_passed: report.automatic_passed, total: report.total, automatic_rate: report.automatic_rate, reviewed: 0 }, null, 2));
if (report.automatic_passed !== report.total) process.exitCode = 1;
