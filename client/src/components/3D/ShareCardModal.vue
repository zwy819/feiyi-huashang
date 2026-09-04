<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="onClose" />

        <div class="relative w-full mx-4 max-w-lg bg-mogao-leadwhite-100 rounded-lg shadow-mogao-lg border border-mogao-ochre-200 max-h-[92vh] flex flex-col">
          <div class="flex items-center justify-between px-6 py-4 border-b border-mogao-ochre-100">
            <h3 class="font-serif text-lg font-semibold text-mogao-lapis-800">分享卡片</h3>
            <button
              @click="onClose"
              class="p-1 rounded-mogao text-mogao-ochre-500 hover:text-mogao-cinnabar hover:bg-mogao-ochre-50 transition-colors"
              aria-label="关闭"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="px-6 py-4 overflow-auto">
            <p v-if="!cardUrl && generating" class="py-16 text-center text-sm text-mogao-ochre-500">
              正在生成分享卡片…
            </p>
            <img
              v-else-if="cardUrl"
              :src="cardUrl"
              alt="服饰分享卡片"
              class="w-full rounded-mogao border border-mogao-ochre-200 shadow-mogao"
            />
          </div>

          <div class="px-6 py-4 border-t border-mogao-ochre-100 flex justify-end gap-2">
            <button @click="onClose" class="btn-mogao-secondary text-sm">关闭</button>
            <button
              @click="download"
              :disabled="!cardUrl"
              class="btn-mogao-primary text-sm disabled:opacity-40 disabled:cursor-not-allowed"
            >
              ⬇ 下载 PNG
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, type Ref } from 'vue'
import type { Garment } from '@/types/garment'

/**
 * ShareCardModal — 3D 截图分享海报
 *
 * Canvas 2D 合成竖版海报：主视觉（3D 截图 / 服饰图 / 程序化纹样底）
 * + 莫高窟配色边框角花 + 服饰信息 + 朱砂印章落款，可下载 PNG。
 */

const props = defineProps<{
  isOpen?: boolean
  garment: Garment | null
  /** ModelViewer 截图 dataURL（可为空，降级用服饰图片或纹样底） */
  screenshot?: string | null
}>()
const emit = defineEmits<{ close: [] }>()

const CARD_W = 900
const CARD_H = 1200

const cardUrl: Ref<string | null> = ref(null)
const generating: Ref<boolean> = ref(false)

// 敦煌色常量（与 tailwind 莫高窟色系一致）
const CINNABAR = '#C43A31'
const CINNABAR_DARK = '#A42E27'
const LAPIS_DARK = '#15243E'
const OCHRE_DARK = '#8F6639'
const GOLD = '#B8862D'
const PAPER = '#F5F0E8'
const PAPER_LIGHT = '#FCFAF7'

function onClose(): void {
  emit('close')
}

/** 手动换行（Canvas 无自动换行） */
function wrapText(ctx: CanvasRenderingContext2D, text: string, maxWidth: number): string[] {
  const lines: string[] = []
  let current = ''
  for (const ch of text) {
    if (ctx.measureText(current + ch).width > maxWidth && current) {
      lines.push(current)
      current = ch
    } else {
      current += ch
    }
  }
  if (current) lines.push(current)
  return lines
}

/** 将主视觉以 cover 模式裁剪绘制到目标矩形 */
function drawCover(
  ctx: CanvasRenderingContext2D,
  img: HTMLImageElement | HTMLCanvasElement,
  x: number, y: number, w: number, h: number,
): void {
  const iw = img.width
  const ih = img.height
  const scale = Math.max(w / iw, h / ih)
  const dw = iw * scale
  const dh = ih * scale
  ctx.drawImage(img, x + (w - dw) / 2, y + (h - dh) / 2, dw, dh)
}

/** 金色角花（四角对称的卷草纹抽象） */
function drawCornerFlower(ctx: CanvasRenderingContext2D, cx: number, cy: number, size: number, flipX: number, flipY: number): void {
  ctx.save()
  ctx.translate(cx, cy)
  ctx.scale(flipX, flipY)
  ctx.strokeStyle = GOLD
  ctx.lineWidth = 3
  ctx.lineCap = 'round'

  ctx.beginPath()
  ctx.moveTo(size * 0.9, 0)
  ctx.quadraticCurveTo(size * 0.35, 0, size * 0.28, size * 0.28)
  ctx.stroke()

  ctx.beginPath()
  ctx.moveTo(0, size * 0.9)
  ctx.quadraticCurveTo(0, size * 0.35, size * 0.28, size * 0.28)
  ctx.stroke()

  ctx.beginPath()
  ctx.arc(size * 0.28, size * 0.28, 4.5, 0, Math.PI * 2)
  ctx.fillStyle = GOLD
  ctx.fill()
  ctx.restore()
}

/** 程序化回退底纹：几何菱格 + 波浪，呼应纺织纹样 */
function drawFallbackPattern(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number): void {
  const grad = ctx.createLinearGradient(0, y, 0, y + h)
  grad.addColorStop(0, '#E8ECF3')
  grad.addColorStop(1, PAPER)
  ctx.fillStyle = grad
  ctx.fillRect(x, y, w, h)

  ctx.save()
  ctx.beginPath()
  ctx.rect(x, y, w, h)
  ctx.clip()
  ctx.strokeStyle = 'rgba(196, 58, 49, 0.16)'
  ctx.lineWidth = 2
  const step = 56
  for (let py = y - h; py < y + h * 2; py += step) {
    for (let px = x - w; px < x + w * 2; px += step) {
      ctx.beginPath()
      ctx.moveTo(px, py + step / 2)
      ctx.lineTo(px + step / 2, py)
      ctx.lineTo(px + step, py + step / 2)
      ctx.lineTo(px + step / 2, py + step)
      ctx.closePath()
      ctx.stroke()
    }
  }
  ctx.restore()
}

/** 朱砂印章：圆角方框 + 二字竖排 */
function drawSeal(ctx: CanvasRenderingContext2D, x: number, y: number, size: number): void {
  ctx.save()
  ctx.translate(x, y)
  ctx.rotate(-0.06)
  ctx.fillStyle = CINNABAR
  ctx.beginPath()
  ctx.roundRect(0, 0, size, size, 8)
  ctx.fill()
  ctx.fillStyle = PAPER_LIGHT
  ctx.font = `600 ${Math.round(size * 0.34)}px "Noto Serif SC", "Songti SC", STSong, serif`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText('非遗', size / 2, size * 0.3)
  ctx.fillText('华裳', size / 2, size * 0.72)
  ctx.restore()
}

async function loadImage(src: string): Promise<HTMLImageElement | null> {
  return new Promise((resolve) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = () => resolve(null)
    img.src = src
  })
}

async function renderCard(): Promise<void> {
  if (!props.garment) return
  generating.value = true

  const g = props.garment
  const canvas = document.createElement('canvas')
  canvas.width = CARD_W
  canvas.height = CARD_H
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    generating.value = false
    return
  }

  // ---- 纸色底 + 底部渐变条 ----
  ctx.fillStyle = PAPER_LIGHT
  ctx.fillRect(0, 0, CARD_W, CARD_H)

  // ---- 主视觉区 ----
  const imgX = 70
  const imgY = 110
  const imgW = CARD_W - imgX * 2
  const imgH = 620

  let mainVisual: HTMLImageElement | HTMLCanvasElement | null = null
  if (props.screenshot) {
    mainVisual = await loadImage(props.screenshot)
  }
  if (!mainVisual && g.image_url) {
    mainVisual = await loadImage(g.image_url)
  }

  if (mainVisual) {
    drawCover(ctx, mainVisual, imgX, imgY, imgW, imgH)
  } else {
    drawFallbackPattern(ctx, imgX, imgY, imgW, imgH)
    // 居中主字
    ctx.fillStyle = 'rgba(196, 58, 49, 0.75)'
    ctx.font = '600 110px "Noto Serif SC", "Songti SC", STSong, serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText('华裳', CARD_W / 2, imgY + imgH / 2 - 40)
    ctx.fillStyle = OCHRE_DARK
    ctx.font = '400 30px "Noto Serif SC", "Songti SC", STSong, serif'
    ctx.fillText('中国非物质文化遗产服饰', CARD_W / 2, imgY + imgH / 2 + 60)
  }

  // 主视觉描边：朱砂外框 + 金色内线
  ctx.strokeStyle = CINNABAR
  ctx.lineWidth = 6
  ctx.strokeRect(imgX - 3, imgY - 3, imgW + 6, imgH + 6)
  ctx.strokeStyle = GOLD
  ctx.lineWidth = 2
  ctx.strokeRect(imgX - 12, imgY - 12, imgW + 24, imgH + 24)

  // ---- 角花 ----
  drawCornerFlower(ctx, 26, 26, 46, 1, 1)
  drawCornerFlower(ctx, CARD_W - 26, 26, 46, -1, 1)
  drawCornerFlower(ctx, 26, CARD_H - 26, 46, 1, -1)
  drawCornerFlower(ctx, CARD_W - 26, CARD_H - 26, 46, -1, -1)

  // ---- 标题区 ----
  ctx.textAlign = 'center'
  ctx.textBaseline = 'alphabetic'
  ctx.fillStyle = LAPIS_DARK
  ctx.font = '600 54px "Noto Serif SC", "Songti SC", STSong, serif'
  const titleLines = wrapText(ctx, g.name, imgW - 100)
  let ty = imgY + imgH + 96
  for (const line of titleLines.slice(0, 2)) {
    ctx.fillText(line, CARD_W / 2, ty)
    ty += 66
  }

  ctx.fillStyle = OCHRE_DARK
  ctx.font = '400 28px "Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif'
  const meta = [g.dynasty ? `${g.dynasty}代` : '', g.ethnic, g.province].filter(Boolean).join(' · ')
  if (meta) {
    ctx.fillText(meta, CARD_W / 2, ty + 8)
    ty += 52
  }

  // 纹样 / 工艺
  ctx.font = '400 24px "Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.fillStyle = '#665D52'
  const detail = [g.pattern ? `纹样 · ${g.pattern}` : '', g.craft ? `工艺 · ${g.craft}` : '']
    .filter(Boolean)
    .join('　　')
  if (detail) {
    const detailLines = wrapText(ctx, detail, imgW - 80)
    for (const line of detailLines.slice(0, 2)) {
      ctx.fillText(line, CARD_W / 2, ty + 6)
      ty += 38
    }
  }

  // ---- 底部落款条 ----
  const footY = CARD_H - 92
  const footGrad = ctx.createLinearGradient(0, footY, CARD_W, footY + 92)
  footGrad.addColorStop(0, CINNABAR)
  footGrad.addColorStop(1, CINNABAR_DARK)
  ctx.fillStyle = footGrad
  ctx.fillRect(0, footY, CARD_W, CARD_H - footY)

  ctx.fillStyle = PAPER_LIGHT
  ctx.font = '600 30px "Noto Serif SC", "Songti SC", STSong, serif'
  ctx.textAlign = 'left'
  ctx.textBaseline = 'middle'
  ctx.fillText('非遗华裳', 56, footY + 46)
  ctx.font = '400 20px "Noto Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.fillStyle = 'rgba(252, 250, 247, 0.85)'
  ctx.fillText('中国非物质文化遗产服饰数字博物馆', 190, footY + 47)

  // 印章
  drawSeal(ctx, CARD_W - 76 - 56, imgY + imgH + 30, 76)

  cardUrl.value = canvas.toDataURL('image/png')
  generating.value = false
}

function download(): void {
  if (!cardUrl.value || !props.garment) return
  const a = document.createElement('a')
  a.href = cardUrl.value
  a.download = `非遗华裳-${props.garment.name || '分享卡片'}.png`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

// 打开时生成，关闭时清理
watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      renderCard()
    } else {
      cardUrl.value = null
    }
  },
)
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
