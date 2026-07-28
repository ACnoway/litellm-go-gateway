import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import UserLayout from '../components/UserLayout.vue'
import ChatCompletions from '../views/ChatCompletions.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: UserLayout,
    children: [
      {
        path: '',
        name: 'user-chat',
        component: ChatCompletions,
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
