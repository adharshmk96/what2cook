<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Breadcrumb from '../components/Breadcrumb.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const userMenu = ref<HTMLElement | null>(null)
const isUserMenuOpen = ref(false)
const userInitial = computed(() => auth.user?.email.charAt(0).toUpperCase() || '?')

const navItems = [
  { name: 'dashboard-quick-recipe', label: 'Quick Recipe' },
  { name: 'dashboard-inventory', label: 'Inventory' },
  { name: 'dashboard-saved-recipes', label: 'Saved Recipes' },
] as const

function isNavActive(name: (typeof navItems)[number]['name']) {
  if (name === 'dashboard-quick-recipe') {
    return (
      route.name === 'dashboard-quick-recipe' ||
      route.name === 'dashboard-quick-recipe-results'
    )
  }
  return route.name === name
}

async function onLogout() {
  isUserMenuOpen.value = false
  await auth.logout()
  await router.replace({ name: 'login' })
}

function closeUserMenu(event: MouseEvent) {
  if (!userMenu.value?.contains(event.target as Node)) isUserMenuOpen.value = false
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') isUserMenuOpen.value = false
}

onMounted(() => {
  document.addEventListener('click', closeUserMenu)
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', closeUserMenu)
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div class="dash-shell">
    <div class="dash-atmosphere" aria-hidden="true" />
    <aside class="dash-sidebar" aria-label="Dashboard">
      <RouterLink class="brand-lockup dash-brand" :to="{ name: 'home' }" aria-label="what2cook home">
        <img src="/icon.png" alt="" width="40" height="40" />
        <span>what2cook</span>
      </RouterLink>
      <nav class="dash-nav" aria-label="Dashboard sections">
        <RouterLink
          v-for="item in navItems"
          :key="item.name"
          class="dash-nav__link"
          :class="{ 'router-link-active': isNavActive(item.name) }"
          active-class=""
          exact-active-class=""
          :to="{ name: item.name }"
        >
          {{ item.label }}
        </RouterLink>
      </nav>
      <div ref="userMenu" class="dash-sidebar__footer">
        <div v-if="isUserMenuOpen" id="user-menu" class="dash-user-menu" role="menu">
          <RouterLink class="dash-user-menu__item" :to="{ name: 'dashboard-account' }" role="menuitem" @click="isUserMenuOpen = false">
            <svg aria-hidden="true" viewBox="0 0 24 24"><circle cx="12" cy="12" r="3" /><path d="M19 15a2 2 0 0 0 .4 2.2l-2.2 2.2A2 2 0 0 0 15 19a2 2 0 0 0-1 2h-4a2 2 0 0 0-1-2 2 2 0 0 0-2.2.4l-2.2-2.2A2 2 0 0 0 5 15a2 2 0 0 0-2-1v-4a2 2 0 0 0 2-1 2 2 0 0 0-.4-2.2l2.2-2.2A2 2 0 0 0 9 5a2 2 0 0 0 1-2h4a2 2 0 0 0 1 2 2 2 0 0 0 2.2-.4l2.2 2.2A2 2 0 0 0 19 9a2 2 0 0 0 2 1v4a2 2 0 0 0-2 1Z" /></svg>
            Settings
          </RouterLink>
          <button type="button" class="dash-user-menu__item dash-user-menu__item--danger" role="menuitem" @click="onLogout">
            <svg aria-hidden="true" viewBox="0 0 24 24"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9" /></svg>
            Log out
          </button>
        </div>
        <button v-if="auth.user" type="button" class="dash-user-trigger" aria-haspopup="menu" aria-controls="user-menu" :aria-expanded="isUserMenuOpen" @click="isUserMenuOpen = !isUserMenuOpen">
          <span class="dash-user-avatar" aria-hidden="true">{{ userInitial }}</span>
          <span class="dash-user-email">{{ auth.user.email }}</span>
          <svg class="dash-user-chevron" :class="{ 'is-open': isUserMenuOpen }" aria-hidden="true" viewBox="0 0 24 24"><path d="m7 10 5 5 5-5" /></svg>
        </button>
      </div>
    </aside>
    <div class="dash-main">
      <header class="dash-mobile-header">
        <RouterLink class="brand-lockup" :to="{ name: 'home' }" aria-label="what2cook home">
          <img src="/icon.png" alt="" width="36" height="36" />
          <span>what2cook</span>
        </RouterLink>
        <button type="button" class="btn-ghost" @click="onLogout">Sign out</button>
      </header>
      <nav class="dash-mobile-nav" aria-label="Dashboard sections">
        <RouterLink
          v-for="item in navItems"
          :key="item.name"
          class="dash-nav__link"
          :class="{ 'router-link-active': isNavActive(item.name) }"
          active-class=""
          exact-active-class=""
          :to="{ name: item.name }"
        >
          {{ item.label }}
        </RouterLink>
      </nav>
      <main class="dash-content">
        <Breadcrumb />
        <RouterView />
      </main>
    </div>
  </div>
</template>
