import { computed, onUnmounted, ref } from 'vue'

export type SpeechKind = 'intro' | 'tour'

/**
 * 浏览器原生中文语音讲解。
 *
 * 不依赖后端、不生成音频文件；优先选择系统里的中文女声，
 * 没有中文语音时回退到浏览器默认语音。
 */
export function useSpeech() {
  const supported = typeof window !== 'undefined'
    && 'speechSynthesis' in window
    && 'SpeechSynthesisUtterance' in window

  const speaking = ref(false)
  const paused = ref(false)
  const activeKind = ref<SpeechKind | null>(null)
  const active = computed(() => activeKind.value !== null)

  let generation = 0
  let utterance: SpeechSynthesisUtterance | null = null

  function pickVoice(): SpeechSynthesisVoice | null {
    if (!supported) return null
    const voices = window.speechSynthesis.getVoices()
    const chinese = voices.filter((voice) => voice.lang?.toLowerCase().startsWith('zh'))
    return chinese.find((voice) => /huihui|yaoyao|xiaoxiao|female/i.test(voice.name))
      ?? chinese[0]
      ?? null
  }

  function reset(): void {
    speaking.value = false
    paused.value = false
    activeKind.value = null
    utterance = null
  }

  function stop(): void {
    generation += 1
    if (supported) window.speechSynthesis.cancel()
    reset()
  }

  function speak(text: string, kind: SpeechKind): boolean {
    if (!supported || !text.trim()) return false

    // 先取消上一段，避免“简介”和“导览”同时播放
    stop()
    const token = generation
    const item = new SpeechSynthesisUtterance(text.trim())
    item.lang = 'zh-CN'
    item.rate = 0.95
    item.pitch = 1
    item.volume = 1

    const voice = pickVoice()
    if (voice) item.voice = voice

    item.onstart = () => {
      if (token !== generation) return
      speaking.value = true
      paused.value = false
      activeKind.value = kind
    }
    item.onend = () => {
      if (token !== generation) return
      reset()
    }
    item.onerror = () => {
      if (token !== generation) return
      reset()
    }
    item.onpause = () => {
      if (token !== generation) return
      speaking.value = false
      paused.value = true
    }
    item.onresume = () => {
      if (token !== generation) return
      speaking.value = true
      paused.value = false
    }

    utterance = item
    // Chrome 在 cancel 后立刻 speak 偶尔不出声，下一帧再启动
    window.setTimeout(() => {
      if (token === generation) window.speechSynthesis.speak(item)
    }, 0)
    activeKind.value = kind
    return true
  }

  function pause(): void {
    if (!supported) return
    window.speechSynthesis.pause()
    speaking.value = false
    paused.value = true
  }

  function resume(): void {
    if (!supported) return
    window.speechSynthesis.resume()
    speaking.value = true
    paused.value = false
  }

  function toggle(text: string, kind: SpeechKind): void {
    if (activeKind.value === kind && speaking.value) {
      pause()
      return
    }
    if (activeKind.value === kind && paused.value) {
      resume()
      return
    }
    speak(text, kind)
  }

  onUnmounted(() => {
    if (utterance) stop()
  })

  return {
    supported,
    speaking,
    paused,
    activeKind,
    active,
    speak,
    pause,
    resume,
    stop,
    toggle,
  }
}
