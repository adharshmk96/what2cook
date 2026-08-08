<script setup lang="ts">
import { ref } from 'vue'
import AuthShell from '../layouts/AuthShell.vue'
import FormError from '../components/FormError.vue'
import { forgotPassword } from '../api/auth'

const email = ref('')
const error = ref<unknown>(null)
const success = ref('')
const submitting = ref(false)

async function onSubmit() {
  error.value = null
  success.value = ''
  const trimmedEmail = email.value.trim()
  if (!trimmedEmail) {
    error.value = new Error('Email is required.')
    return
  }

  submitting.value = true
  try {
    await forgotPassword(trimmedEmail)
    success.value =
      'If an account exists for that email, you will receive reset instructions shortly.'
  } catch (err) {
    error.value = err
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthShell
    title="Reset your password"
    subtitle="Enter your email and we will send a reset link if the account exists."
  >
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
      <FormError :error="error" />
      <p v-if="success" class="form-success" role="status">{{ success }}</p>
      <button class="btn-primary" type="submit" :disabled="submitting">
        {{ submitting ? 'Sending…' : 'Send reset link' }}
      </button>
    </form>
    <p class="auth-links">
      <RouterLink :to="{ name: 'login' }">Back to sign in</RouterLink>
    </p>
  </AuthShell>
</template>
