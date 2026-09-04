<template>
  <div class="page-container max-w-3xl mx-auto">
    <!-- ======== 开始页 ======== -->
    <div v-if="phase === 'start'" class="text-center py-10">
      <span class="text-7xl block mb-4">&#x1f3af;</span>
      <h1 class="font-serif text-3xl font-bold text-mogao-lapis-800">华裳问答</h1>
      <p class="text-sm text-mogao-ochre-500 mt-3 max-w-md mx-auto leading-relaxed">
        从馆藏 40 余件非遗服饰中自动出题，考考你民族、朝代、纹样与工艺的知识。
        每答对一题得 10 分，答错也有纹样科普可以涨知识。
      </p>

      <div class="flex items-center justify-center gap-4 mt-8">
        <button @click="startGame(5)" class="btn-mogao-primary px-8 py-3 text-base">快玩 5 题</button>
        <button @click="startGame(10)" class="btn-mogao-secondary px-8 py-3 text-base">挑战 10 题</button>
      </div>

      <p v-if="!isLoggedIn" class="text-xs text-mogao-ochre-400 mt-6">
        游客可正常游玩，登录后成绩会记入排行榜
      </p>
      <p v-else-if="myBest > 0" class="text-xs text-mogao-ochre-500 mt-6">
        你的最高分：<span class="text-mogao-cinnabar font-semibold">{{ myBest }}</span> 分
      </p>

      <!-- 排行榜 -->
      <div v-if="leaderboard.length > 0" class="card-mogao p-5 mt-10 text-left">
        <h3 class="font-serif text-base font-semibold text-mogao-lapis-800 mb-3">&#x1f3c6; 排行榜</h3>
        <div v-for="(item, i) in leaderboard" :key="item.user_id"
          class="flex items-center gap-3 py-1.5 text-sm border-b border-mogao-ochre-100 last:border-0">
          <span class="w-6 text-center font-semibold" :class="medalClass(i)">{{ i + 1 }}</span>
          <span class="flex-1 text-mogao-lapis-700">{{ item.username }}</span>
          <span class="text-xs text-mogao-ochre-500">{{ item.correct_count }}/{{ item.total }} 题</span>
          <span class="font-semibold text-mogao-cinnabar w-14 text-right">{{ item.score }} 分</span>
        </div>
      </div>
    </div>

    <!-- ======== 答题页 ======== -->
    <div v-else-if="phase === 'playing' && questions.length > 0">
      <!-- 进度与计时 -->
      <div class="flex items-center justify-between mb-2 text-xs text-mogao-ochre-500">
        <span>第 {{ currentIndex + 1 }} / {{ questions.length }} 题</span>
        <span class="font-mono">&#x23f1; {{ elapsed }}s</span>
      </div>
      <div class="h-1.5 bg-mogao-ochre-100 rounded-full overflow-hidden mb-6">
        <div
          class="h-full bg-gradient-to-r from-mogao-cinnabar to-mogao-gold transition-all duration-300"
          :style="{ width: `${((currentIndex + 1) / questions.length) * 100}%` }"
        />
      </div>

      <!-- 题目卡片 -->
      <div v-if="currentQuestion" class="card-mogao p-6">
        <h2 class="font-serif text-lg font-semibold text-mogao-lapis-800 leading-relaxed">
          {{ currentQuestion.stem }}
        </h2>

        <div class="grid gap-3 mt-5">
          <button
            v-for="(option, i) in currentQuestion.options"
            :key="i"
            @click="choose(option)"
            :disabled="chosen !== null"
            class="px-4 py-3 text-left text-sm rounded-mogao border transition-all duration-200"
            :class="optionClass(option)"
          >
            <span class="inline-block w-5 font-mono text-xs mr-2 opacity-60">{{ 'ABCD'[i] }}</span>
            {{ option }}
          </button>
        </div>

        <!-- 即时判定 + 解析 -->
        <Transition name="fade">
          <div v-if="chosen !== null" class="mt-5">
            <div
              class="rounded-mogao p-4 text-sm"
              :class="chosenCorrect
                ? 'bg-mogao-malachite-50 border border-mogao-malachite-200'
                : 'bg-mogao-cinnabar-50 border border-mogao-cinnabar-200'"
            >
              <p class="font-semibold" :class="chosenCorrect ? 'text-mogao-malachite-600' : 'text-mogao-cinnabar-600'">
                {{ chosenCorrect ? '✔ 回答正确' : `✘ 回答错误，正确答案是「${correctAnswer}」` }}
              </p>
              <p v-if="currentQuestion.explanation" class="text-xs text-mogao-lapis-600 leading-relaxed mt-2">
                {{ currentQuestion.explanation }}
              </p>
              <router-link
                :to="`/detail/${currentQuestion.garment_id}`"
                class="inline-block text-xs text-mogao-lapis-500 hover:text-mogao-cinnabar mt-2 underline underline-offset-2"
              >
                查看该服饰 →
              </router-link>
            </div>

            <button @click="next" class="btn-mogao-primary w-full mt-4">
              {{ currentIndex + 1 < questions.length ? '下一题' : '查看结果' }}
            </button>
          </div>
        </Transition>
      </div>
    </div>

    <!-- ======== 结果页 ======== -->
    <div v-else-if="phase === 'result' && result" class="text-center py-8">
      <div class="card-mogao p-8">
        <p class="text-sm text-mogao-ochre-500">
          {{ gradeComment }}
        </p>
        <p class="font-serif text-6xl font-bold text-mogao-cinnabar mt-4">{{ result.score }}</p>
        <p class="text-xs text-mogao-ochre-500 mt-1">分</p>

        <div class="flex items-center justify-center gap-8 mt-6 text-sm">
          <div>
            <p class="text-mogao-lapis-800 font-semibold text-lg">{{ result.correct_count }}/{{ result.total }}</p>
            <p class="text-xs text-mogao-ochre-500 mt-0.5">答对题数</p>
          </div>
          <div>
            <p class="text-mogao-lapis-800 font-semibold text-lg">{{ result.duration }}s</p>
            <p class="text-xs text-mogao-ochre-500 mt-0.5">用时</p>
          </div>
          <div>
            <p class="text-mogao-lapis-800 font-semibold text-lg">{{ result.best_score }}</p>
            <p class="text-xs text-mogao-ochre-500 mt-0.5">历史最高</p>
          </div>
        </div>

        <p v-if="!isLoggedIn" class="text-xs text-mogao-ochre-400 mt-5">
          本次为游客模式，
          <router-link to="/login" class="text-mogao-cinnabar underline underline-offset-2">登录</router-link>
          后成绩将记入排行榜
        </p>
        <p v-else-if="isRecorded" class="text-xs text-mogao-malachite-600 mt-5">✔ 成绩已记入排行榜</p>

        <div class="flex items-center justify-center gap-3 mt-6">
          <button @click="backHome" class="btn-mogao-secondary text-sm">返回首页</button>
          <button @click="restart" class="btn-mogao-primary text-sm">再来一局</button>
        </div>
      </div>

      <!-- 排行榜 -->
      <div v-if="leaderboard.length > 0" class="card-mogao p-5 mt-6 text-left">
        <h3 class="font-serif text-base font-semibold text-mogao-lapis-800 mb-3">&#x1f3c6; 排行榜</h3>
        <div v-for="(item, i) in leaderboard" :key="item.user_id"
          class="flex items-center gap-3 py-1.5 text-sm border-b border-mogao-ochre-100 last:border-0">
          <span class="w-6 text-center font-semibold" :class="medalClass(i)">{{ i + 1 }}</span>
          <span class="flex-1 text-mogao-lapis-700">{{ item.username }}</span>
          <span class="text-xs text-mogao-ochre-500">{{ item.correct_count }}/{{ item.total }} 题</span>
          <span class="font-semibold text-mogao-cinnabar w-14 text-right">{{ item.score }} 分</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, type Ref } from 'vue'
import { useRouter, type Router } from 'vue-router'
import { quizAPI } from '@/services/api'
import type {
  QuizQuestion,
  QuizSubmitResult,
  QuizLeaderboardItem,
  QuizAnswerItem,
} from '@/types/quiz'
import { useToast } from '@/composables'

const { showToast } = useToast()
const router: Router = useRouter()

type Phase = 'start' | 'playing' | 'result'

const phase: Ref<Phase> = ref('start')
const questions: Ref<QuizQuestion[]> = ref([])
const currentIndex: Ref<number> = ref(0)
const chosen: Ref<string | null> = ref(null)
const chosenCorrect: Ref<boolean> = ref(false)
const correctAnswer: Ref<string> = ref('')
const result: Ref<QuizSubmitResult | null> = ref(null)
const leaderboard: Ref<QuizLeaderboardItem[]> = ref([])
const myBest: Ref<number> = ref(0)
const isRecorded: Ref<boolean> = ref(false)
/** 每题作答记录（question_id → 所选文本），最后统一提交判分 */
const chosenMap: Ref<Record<string, string>> = ref({})

const isLoggedIn = computed(() => !!localStorage.getItem('token'))

// 计时器
const elapsed: Ref<number> = ref(0)
let startedAt = 0
let timerId: ReturnType<typeof setInterval> | null = null

const currentQuestion = computed<QuizQuestion | null>(() => questions.value[currentIndex.value] || null)

const gradeComment = computed<string>(() => {
  if (!result.value) return ''
  const ratio = result.value.correct_count / Math.max(result.value.total, 1)
  if (ratio === 1) return '满贯！你已是华裳大家 👘'
  if (ratio >= 0.8) return '博古通今，功底不俗！'
  if (ratio >= 0.5) return '渐入佳境，再接再厉～'
  return '初识华裳，慢慢逛逛博物馆吧～'
})

function startTimer(): void {
  startedAt = Date.now()
  elapsed.value = 0
  stopTimer()
  timerId = setInterval(() => {
    elapsed.value = Math.floor((Date.now() - startedAt) / 1000)
  }, 1000)
}

function stopTimer(): void {
  if (timerId) {
    clearInterval(timerId)
    timerId = null
  }
}

function medalClass(index: number): string {
  if (index === 0) return 'text-mogao-gold'
  if (index === 1) return 'text-mogao-ochre-600'
  if (index === 2) return 'text-mogao-ochre-500'
  return 'text-mogao-ochre-400'
}

async function loadLeaderboard(): Promise<void> {
  try {
    const res = await quizAPI.leaderboard(10)
    leaderboard.value = res.data?.items || []
  } catch {
    leaderboard.value = []
  }
}

async function startGame(count: number): Promise<void> {
  try {
    const res = await quizAPI.getQuestions(count)
    questions.value = res.data?.questions || []
  } catch (err: unknown) {
    showToast(err instanceof Error ? err.message : '题目加载失败', 'error')
    return
  }
  if (questions.value.length === 0) {
    showToast('暂无可用题目', 'error')
    return
  }
  currentIndex.value = 0
  chosen.value = null
  chosenCorrect.value = false
  correctAnswer.value = ''
  chosenMap.value = {}
  result.value = null
  isRecorded.value = false
  phase.value = 'playing'
  startTimer()
}

function optionClass(option: string): string {
  if (chosen.value === null) {
    return 'bg-white border-mogao-ochre-200 text-mogao-lapis-700 hover:border-mogao-cinnabar hover:bg-mogao-ochre-50'
  }
  // 已选择但判定未返回：选中项中性高亮，避免先红后绿的闪烁
  if (correctAnswer.value === '') {
    return option === chosen.value
      ? 'bg-mogao-lapis-50 border-mogao-lapis-300 text-mogao-lapis-700'
      : 'bg-white border-mogao-ochre-100 text-mogao-ochre-400 opacity-60'
  }
  // 已作答：正确项绿色高亮，错选红色
  if (option === correctAnswer.value) {
    return 'bg-mogao-malachite-50 border-mogao-malachite-400 text-mogao-malachite-700 font-medium'
  }
  if (option === chosen.value) {
    return 'bg-mogao-cinnabar-50 border-mogao-cinnabar-400 text-mogao-cinnabar-600'
  }
  return 'bg-white border-mogao-ochre-100 text-mogao-ochre-400 opacity-60'
}

async function choose(option: string): Promise<void> {
  const q = currentQuestion.value
  if (chosen.value !== null || !q) return
  chosen.value = option
  chosenMap.value[q.id] = option

  // 单题即时判定（答案在选择后才由服务端返回）
  try {
    const res = await quizAPI.check({ garment_id: q.garment_id, field: q.field, choice: option })
    chosenCorrect.value = !!res.data?.correct
    correctAnswer.value = res.data?.answer || option
  } catch {
    // 判定失败时仅标记选项，不给正误反馈
    showToast('判定服务暂不可用', 'error')
    chosenCorrect.value = false
    correctAnswer.value = option
  }
}

function next(): void {
  if (currentIndex.value + 1 < questions.value.length) {
    currentIndex.value++
    chosen.value = null
    chosenCorrect.value = false
    correctAnswer.value = ''
  } else {
    submitAnswers()
  }
}

async function submitAnswers(): Promise<void> {
  stopTimer()
  const answers: QuizAnswerItem[] = questions.value.map((q) => ({
    garment_id: q.garment_id,
    field: q.field,
    choice: chosenMap.value[q.id] || '',
  }))

  try {
    const res = await quizAPI.submit({ duration_seconds: elapsed.value, answers })
    result.value = res.data
    isRecorded.value = isLoggedIn.value && !!result.value
  } catch (err: unknown) {
    showToast(err instanceof Error ? err.message : '提交失败', 'error')
    phase.value = 'start'
    return
  }

  // 登录用户刷新最高分
  if (isLoggedIn.value && result.value) {
    myBest.value = result.value.best_score
  }
  await loadLeaderboard()
  phase.value = 'result'
}

function backHome(): void {
  router.push('/')
}

function restart(): void {
  phase.value = 'start'
  questions.value = []
  result.value = null
  chosenMap.value = {}
}

onMounted(async () => {
  await loadLeaderboard()
  if (isLoggedIn.value) {
    try {
      const res = await quizAPI.myRecords()
      myBest.value = res.data?.best || 0
    } catch {
      myBest.value = 0
    }
  }
})

onUnmounted(stopTimer)
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
