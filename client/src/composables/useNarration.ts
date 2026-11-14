import { computed, onUnmounted, ref } from 'vue'
import { useSpeech, type SpeechKind } from './useSpeech'

/**
 * 优先播放服务端 Qwen-TTS 生成的自然中文语音；
 * 服务端不可用/生成失败时，自动回退到浏览器语音。
 */
export function useNarration() {
  const browser = useSpeech()
  const audio = typeof Audio !== 'undefined' ? new Audio() : null
  if (audio) {
    audio.preload = 'auto'
  }

  const supported = Boolean(audio) || browser.supported
  const loading = ref(false)
  const remoteSpeaking = ref(false)
  const remotePaused = ref(false)
  const remoteKind = ref<SpeechKind | null>(null)

  const activeKind = computed(() => remoteKind.value ?? browser.activeKind.value)
  const speaking = computed(() => remoteSpeaking.value || browser.speaking.value)
  const paused = computed(() => remoteKind.value ? remotePaused.value : browser.paused.value)
  const active = computed(() => activeKind.value !== null)

  let token = 0
  let lastFallbackText = ''
  let lastKind: SpeechKind = 'intro'

  function resetRemote(): void {
    remoteSpeaking.value = false
    remotePaused.value = false
    remoteKind.value = null
    loading.value = false
  }

  function stop(): void {
    token += 1
    if (audio) {
      audio.onerror = null
      audio.pause()
      audio.removeAttribute('src')
      audio.load()
    }
    resetRemote()
    browser.stop()
  }

  function playRemote(url: string, fallbackText: string, kind: SpeechKind): void {
    if (!audio) {
      browser.speak(fallbackText, kind)
      return
    }

    stop()
    const current = ++token
    lastFallbackText = fallbackText
    lastKind = kind
    remoteKind.value = kind
    remoteSpeaking.value = false
    remotePaused.value = false
    loading.value = true

    const fallback = (): void => {
      if (current !== token) return
      audio.onerror = null
      audio.removeAttribute('src')
      audio.load()
      resetRemote()
      browser.speak(lastFallbackText, lastKind)
    }

    audio.onplaying = () => {
      if (current !== token) return
      loading.value = false
      remoteSpeaking.value = true
      remotePaused.value = false
    }
    audio.onpause = () => {
      if (current !== token) return
      if (!audio.ended) {
        remoteSpeaking.value = false
        remotePaused.value = true
      }
    }
    audio.onended = () => {
      if (current !== token) return
      resetRemote()
    }
    audio.onerror = fallback
    audio.onstalled = () => {
      if (current === token) loading.value = true
    }
    audio.onwaiting = () => {
      if (current === token) loading.value = true
    }

    audio.src = url
    audio.currentTime = 0
    void audio.play().catch(fallback)
  }

  function speakGarment(id: number, fallbackText: string): void {
    playRemote(`/api/tts/garment/${id}`, fallbackText, 'intro')
  }

  function speakAnnotation(id: number, step: number, fallbackText: string): void {
    playRemote(`/api/tts/annotation/${id}?step=${step}`, fallbackText, 'tour')
  }

  function pause(): void {
    if (remoteKind.value && audio) {
      audio.pause()
      remoteSpeaking.value = false
      remotePaused.value = true
      return
    }
    browser.pause()
  }

  function resume(): void {
    if (remoteKind.value && audio) {
      void audio.play().then(() => {
        remoteSpeaking.value = true
        remotePaused.value = false
      }).catch(() => {
        resetRemote()
        browser.speak(lastFallbackText, lastKind)
      })
      return
    }
    browser.resume()
  }

  onUnmounted(stop)

  return {
    supported,
    loading,
    speaking,
    paused,
    activeKind,
    active,
    speakGarment,
    speakAnnotation,
    pause,
    resume,
    stop,
  }
}
