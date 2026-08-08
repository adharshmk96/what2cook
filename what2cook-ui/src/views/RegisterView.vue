<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthShell from '../layouts/AuthShell.vue'
import FormError from '../components/FormError.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')
const confirm = ref('')
const error = ref<unknown>(null)
const submitting = ref(false)

async function onSubmit() {
  error.value = null
  const trimmedEmail = email.value.trim()
  if (!trimmedEmail || !password.value) {
    error.value = new Error('Email and password are required.')
    return
  }
  if (password.value.length < 8) {
    error.value = new Error('Password must be at least 8 characters.')
    return
  }
  if (password.value !== confirm.value) {
    error.value = new Error('Passwords do not match.')
    return
  }

  submitting.value = true
  try {
    await auth.register(trimmedEmail, password.value)
    await router.replace({ name: 'dashboard' })
  } catch (err) {
    error.value = err
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthShell title="Join the kitchen" subtitle="Create an account and start cooking smarter.">
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
          autocomplete="new-password"
          required
          minlength="8"
          placeholder="At least 8 characters"
        />
      </label>
      <label class="field">
        <span>Confirm password</span>
        <input
          v-model="confirm"
          type="password"
          name="confirm"
          autocomplete="new-password"
          required
          minlength="8"
          placeholder="Repeat password"
        />
      </label>
      <FormError :error="error" />
      <button class="btn-primary" type="submit" :disabled="submitting || auth.loading">
        {{ submitting ? 'Creating…' : 'Create account' }}
      </button>
    </form>
    <p class="auth-links">
      <RouterLink :to="{ name: 'login' }">Already have an account? Sign in</RouterLink>
    </p>
  </AuthShell>
</template>
