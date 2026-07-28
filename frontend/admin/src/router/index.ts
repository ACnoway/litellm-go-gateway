import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import MainLayout from '../components/MainLayout.vue'
import AdminDashboard from '../views/AdminDashboard.vue'
import DeploymentsList from '../views/DeploymentsList.vue'
import DeploymentsForm from '../views/DeploymentsForm.vue'
import ModelsList from '../views/ModelsList.vue'
import ProvidersList from '../views/ProvidersList.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/admin',
    component: MainLayout,
    children: [
      {
        path: '',
        name: 'admin-dashboard',
        component: AdminDashboard,
      },
      {
        path: 'deployments',
        name: 'admin-deployments-list',
        component: DeploymentsList,
      },
      {
        path: 'deployments/new',
        name: 'admin-deployments-create',
        component: DeploymentsForm,
        props: { isEdit: false },
      },
      {
        path: 'deployments/:id/edit',
        name: 'admin-deployments-edit',
        component: DeploymentsForm,
        props: (route) => ({ isEdit: true, id: Number(route.params.id) }),
      },
      {
        path: 'models',
        name: 'admin-models-list',
        component: ModelsList,
      },
      {
        path: 'providers',
        name: 'admin-providers-list',
        component: ProvidersList,
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
