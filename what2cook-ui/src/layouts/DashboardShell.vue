<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import Breadcrumb from '../components/Breadcrumb.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const navItems = [
  { name: 'dashboard-quick-recipe', label: 'Quick Recipe' },
  { name: 'dashboard-inventory', label: 'Inventory' },
  { name: 'dashboard-saved-recipes', label: 'Saved Recipes' },
  { name: 'dashboard-account', label: 'Account' },
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
  await auth.logout()
  await router.replace({ name: 'login' })
}
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
      <div class="dash-sidebar__footer">
        <p v-if="auth.user" class="dash-user-email">{{ auth.user.email }}</p>
        <button type="button" class="btn-ghost" @click="onLogout">Sign out</button>
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
