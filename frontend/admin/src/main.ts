import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter } from 'vue-router'

import App from './App.vue'
import router from './router'
import '/src/styles/tailwind.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.mount('#app')
