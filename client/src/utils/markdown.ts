/**
 * A deliberately small Markdown renderer for AI answers.
 *
 * The model output is untrusted, so raw HTML is escaped before a fixed set of
 * tags is added. This avoids pulling a full Markdown + DOM sanitizer bundle into
 * the initial page while keeping citations clickable.
 */
export function renderAiMarkdown(input: string): string {
  const source = (input || '').replace(/\r\n/g, '\n')
  const lines = source.split('\n')
  const html: string[] = []
  let list: 'ul' | 'ol' | null = null

  const closeList = () => {
    if (list) {
      html.push(`</${list}>`)
      list = null
    }
  }

  for (const rawLine of lines) {
    const line = rawLine.trimEnd()
    if (!line.trim()) {
      closeList()
      continue
    }

    const heading = line.match(/^(#{1,3})\s+(.+)$/)
    if (heading) {
      closeList()
      html.push(`<h3>${inline(heading[2])}</h3>`)
      continue
    }

    const bullet = line.match(/^[-*]\s+(.+)$/)
    if (bullet) {
      if (list !== 'ul') {
        closeList()
        list = 'ul'
        html.push('<ul>')
      }
      html.push(`<li>${inline(bullet[1])}</li>`)
      continue
    }

    const ordered = line.match(/^\d+[.)]\s+(.+)$/)
    if (ordered) {
      if (list !== 'ol') {
        closeList()
        list = 'ol'
        html.push('<ol>')
      }
      html.push(`<li>${inline(ordered[1])}</li>`)
      continue
    }

    closeList()
    html.push(`<p>${inline(line)}</p>`)
  }
  closeList()
  return html.join('')
}

function inline(text: string): string {
  let value = escapeHtml(text)
  value = value.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  value = value.replace(/`([^`]+)`/g, '<code>$1</code>')
  value = value.replace(/\[S(\d+)\]/g, (_match, id: string) =>
    `<button type="button" class="ai-cite" data-cite="S${id}" aria-label="查看来源 S${id}">S${id}</button>`,
  )
  return value
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}
