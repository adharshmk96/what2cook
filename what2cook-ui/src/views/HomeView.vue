<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

async function onLogout() {
  await auth.logout()
  await router.replace({ name: 'login' })
}
</script>

<template>
  <div class="home-shell">
    <div class="home-atmosphere" aria-hidden="true" />
    <header class="home-header">
      <p class="brand brand--home">what2cook</p>
      <nav class="home-nav">
        <RouterLink :to="{ name: 'change-password' }">Change password</RouterLink>
        <button type="button" class="btn-ghost" @click="onLogout">Sign out</button>
      </nav>
    </header>
    <main class="home-main">
      <h1>You're in</h1>
      <p class="home-lead">
        Signed in as
        <strong>{{ auth.user?.email ?? 'your account' }}</strong>.
        Recipe discovery comes next.
      </p>
    </main>
  </div>
</template>
