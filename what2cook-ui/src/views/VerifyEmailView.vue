<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import AuthShell from '../layouts/AuthShell.vue'
import FormError from '../components/FormError.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()

const status = ref<'pending' | 'ok' | 'error'>('pending')
const error = ref<unknown>(null)

onMounted(async () => {
  const token = typeof route.query.token === 'string' ? route.query.token : ''
  if (!token) {
    status.value = 'error'
    error.value = new Error('Missing verification token.')
    return
  }

  try {
    await auth.verifyEmail(token)
    status.value = 'ok'
  } catch (err) {
    status.value = 'error'
    error.value = err
  }
})
</script>

<template>
  <AuthShell
    title="Verify email"
    :subtitle="
      status === 'ok'
        ? 'Your email is confirmed. You’re ready to cook.'
        : status === 'pending'
          ? 'Confirming your email…'
          : 'We couldn’t verify that link.'
    "
  >
    <div class="auth-form">
      <p v-if="status === 'pending'" class="dash-panel__desc" role="status">Please wait…</p>
      <p v-else-if="status === 'ok'" class="form-success" role="status">Email verified successfully.</p>
      <FormError v-else :error="error" />
      <p class="auth-links">
        <RouterLink v-if="auth.isAuthenticated" :to="{ name: 'dashboard' }">Go to dashboard</RouterLink>
        <RouterLink v-else :to="{ name: 'login' }">Sign in</RouterLink>
      </p>
    </div>
  </AuthShell>
</template>
