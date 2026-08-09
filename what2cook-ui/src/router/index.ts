import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import DashboardShell from '../layouts/DashboardShell.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/HomeView.vue'),
    },
    {
      path: '/app',
      redirect: '/',
    },
    {
      path: '/app/',
      redirect: '/',
    },
    {
      path: '/app/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/app/register',
      name: 'register',
      component: () => import('../views/RegisterView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/app/forgot-password',
      name: 'forgot-password',
      component: () => import('../views/ForgotPasswordView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/app/reset-password',
      name: 'reset-password',
      component: () => import('../views/ResetPasswordView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/app/verify-email',
      name: 'verify-email',
      component: () => import('../views/VerifyEmailView.vue'),
    },
    {
      path: '/app/change-password',
      redirect: { name: 'dashboard-account' },
    },
    {
      path: '/app/dashboard',
      component: DashboardShell,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'dashboard',
          redirect: { name: 'dashboard-quick-recipe' },
        },
        {
          path: 'quick-recipe',
          name: 'dashboard-quick-recipe',
          component: () => import('../views/dashboard/QuickRecipeView.vue'),
          meta: {
            breadcrumb: [{ label: 'Quick Recipe' }],
          },
        },
        {
          path: 'quick-recipe/results',
          name: 'dashboard-quick-recipe-results',
          component: () => import('../views/dashboard/RecipeResultsView.vue'),
          meta: {
            breadcrumb: [
              { label: 'Quick Recipe', name: 'dashboard-quick-recipe' },
              { label: 'Results' },
            ],
          },
        },
        {
          path: 'inventory',
          name: 'dashboard-inventory',
          component: () => import('../views/dashboard/InventoryView.vue'),
          meta: {
            breadcrumb: [{ label: 'Inventory' }],
          },
        },
        {
          path: 'saved-recipes',
          name: 'dashboard-saved-recipes',
          component: () => import('../views/dashboard/SavedRecipesView.vue'),
          meta: {
            breadcrumb: [{ label: 'Saved Recipes' }],
          },
        },
        {
          path: 'account',
          name: 'dashboard-account',
          component: () => import('../views/dashboard/AccountSettingsView.vue'),
          meta: {
            breadcrumb: [{ label: 'Settings' }],
          },
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  await auth.bootstrap()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }

  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: 'dashboard' }
  }

  return true
})

export default router
