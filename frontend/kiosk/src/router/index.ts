import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useKioskStore } from '@/stores/kiosk'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: { title: '首页', step: 0 }
  },
  {
    path: '/idcard',
    name: 'IdCard',
    component: () => import('@/views/IdCard.vue'),
    meta: { title: '身份证登记', step: 1 }
  },
  {
    path: '/dispute-type',
    name: 'DisputeType',
    component: () => import('@/views/DisputeType.vue'),
    meta: { title: '纠纷类型', step: 2 }
  },
  {
    path: '/form',
    name: 'Form',
    component: () => import('@/views/Form.vue'),
    meta: { title: '填写信息', step: 3 }
  },
  {
    path: '/evidence',
    name: 'Evidence',
    component: () => import('@/views/Evidence.vue'),
    meta: { title: '上传证据', step: 4 }
  },
  {
    path: '/confirm',
    name: 'Confirm',
    component: () => import('@/views/Confirm.vue'),
    meta: { title: '信息确认', step: 5 }
  },
  {
    path: '/success',
    name: 'Success',
    component: () => import('@/views/Success.vue'),
    meta: { title: '登记完成', step: 6 }
  },
  {
    path: '/ai-help',
    name: 'AIHelp',
    component: () => import('@/views/AIHelp.vue'),
    meta: { title: 'AI法律咨询' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  const store = useKioskStore()
  if (to.meta.step !== undefined) {
    store.currentStep = to.meta.step as number
  }
  next()
})

export default router
