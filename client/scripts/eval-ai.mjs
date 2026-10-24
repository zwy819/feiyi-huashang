import { readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'

const args = process.argv.slice(2)
function arg(name, fallback) {
  const i = args.indexOf(`--${name}`)
  return i >= 0 && args[i + 1] ? args[i + 1] : fallback
}

const baseURL = arg('url', process.env.APP_URL || 'http://localhost:3001')
const outPath = path.resolve(arg('out', 'eval-report.json'))
const delay = Number(arg('delay', '0'))
const limit = Number(arg('limit', '0'))
const datasetPath = new URL('../../server/knowledge/eval/questions.json', import.meta.url)
const dataset = JSON.parse(await readFile(datasetPath, 'utf8'))
const questions = limit > 0 ? dataset.questions.slice(0, limit) : dataset.questions

function parseSSE(text) {
  const result = { sources: [], content: '', action: null, errors: [], done: null }
  for (const block of text.replace(/\r\n/g, '\n').split('\n\n')) {
    let event = 'message'
    const data = []
    for (const line of block.split('\n')) {
      if (line.startsWith('event:')) event = line.slice(6).trim()
      else if (line.startsWith('data:')) data.push(line.slice(5).trimStart())
    }
    if (!data.length) continue
    let payload
    try { payload = JSON.parse(data.join('\n')) } catch { continue }
    if (event === 'sources') result.sources = payload
    else if (event === 'delta') result.content += payload.text || ''
    else if (event === 'action') result.action = payload
    else if (event === 'done') { result.done = payload; result.content = payload.content || result.content }
    else if (event === 'error') result.errors.push(payload)
  }
  return result
}

function refusalOK(content, item) {
  if (!item.should_refuse) return true
  const refusal = /未找到直接依据|无法|不能|没有|不提供|不应|拒绝|无法判断|不能判断|不足以|不便/.test(content)
  const forbidden = (item.must_not_contain || []).filter((term) => content.includes(term))
  return refusal && forbidden.length === 0
}

const report = { started_at: new Date().toISOString(), base_url: baseURL, total: questions.length, items: [] }
let passed = 0
for (let i = 0; i < questions.length; i++) {
  const item = questions[i]
  const started = Date.now()
  let parsed = { sources: [], content: '', action: null, errors: [], done: null }
  let status = 0
  try {
    const response = await fetch(baseURL + '/api/ai/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'text/event-stream' },
      body: JSON.stringify({ message: item.question, history: [], context: { page: 'home' } }),
    })
    status = response.status
    parsed = parseSSE(await response.text())
  } catch (error) {
    parsed.errors.push({ message: String(error) })
  }

  const sourceTitles = parsed.sources.map((s) => s.title || '').join(' | ')
  const sourcesOK = (item.expected_sources || []).every((needle) => sourceTitles.includes(needle))
  const termsOK = (item.expected_terms || []).every((term) => parsed.content.includes(term))
  const citeOK = item.should_refuse || /\[S\d+\]/.test(parsed.content)
  const refusedOK = refusalOK(parsed.content, item)
  const ok = status === 200 && parsed.errors.length === 0 && sourcesOK && termsOK && citeOK && refusedOK
  if (ok) passed++

  report.items.push({
    id: item.id,
    category: item.category,
    question: item.question,
    ok,
    status,
    sources_ok: sourcesOK,
    terms_ok: termsOK,
    citation_ok: citeOK,
    refusal_ok: refusedOK,
    action: parsed.action,
    source_ids: parsed.sources.map((s) => s.id),
    content_excerpt: parsed.content.slice(0, 260),
    errors: parsed.errors,
    latency_ms: Date.now() - started,
  })
  console.log(`${ok ? 'PASS' : 'FAIL'} ${item.id} ${item.category} (${Date.now() - started}ms)`)
  if (delay > 0 && i < questions.length - 1) await new Promise((r) => setTimeout(r, delay))
}

report.finished_at = new Date().toISOString()
report.passed = passed
report.rate = questions.length ? Number((passed / questions.length).toFixed(3)) : 0
await writeFile(outPath, JSON.stringify(report, null, 2) + '\n', 'utf8')
console.log(JSON.stringify({ out: outPath, passed, total: questions.length, rate: report.rate }, null, 2))
