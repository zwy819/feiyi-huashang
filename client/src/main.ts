import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { toastPlugin } from './composables'
import './styles/index.css'

const app = createApp(App)

app.use(router)
app.use(toastPlugin)

app.config.errorHandler = (err: unknown, _instance, info) => {
  console.error('[Vue ErrorHandler]', err, info)
}

app.mount('#app')
