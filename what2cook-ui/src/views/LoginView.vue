<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AuthShell from '../layouts/AuthShell.vue'
import FormError from '../components/FormError.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const error = ref<unknown>(null)
const submitting = ref(false)

async function onSubmit() {
  error.value = null
  const trimmedEmail = email.value.trim()
  if (!trimmedEmail || !password.value) {
    error.value = new Error('Email and password are required.')
    return
  }

  submitting.value = true
  try {
    await auth.login(trimmedEmail, password.value)
    const redirect =
      typeof route.query.redirect === 'string' && route.query.redirect
        ? route.query.redirect
        : '/app/dashboard'
    await router.replace(redirect)
  } catch (err) {
    error.value = err
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthShell title="Welcome back" subtitle="Sign in to find your next meal.">
    <form class="auth-form" @submit.prevent="onSubmit">
      <label class="field">
        <span>Email</span>
        <input
          v-model="email"
          type="email"
          name="email"
          autocomplete="email"
          required
          placeholder="you@example.com"
        />
      </label>
      <label class="field">
        <span>Password</span>
        <input
          v-model="password"
          type="password"
          name="password"
          autocomplete="current-password"
          required
          minlength="8"
          placeholder="Your password"
        />
      </label>
      <FormError :error="error" />
      <button class="btn-primary" type="submit" :disabled="submitting || auth.loading">
        {{ submitting ? 'Signing in…' : 'Sign in' }}
      </button>
    </form>
    <p class="auth-links">
      <RouterLink :to="{ name: 'forgot-password' }">Forgot password?</RouterLink>
      <RouterLink :to="{ name: 'register' }">Create an account</RouterLink>
    </p>
  </AuthShell>
</template>
