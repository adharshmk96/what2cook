<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AuthShell from '../layouts/AuthShell.vue'
import FormError from '../components/FormError.vue'
import { resetPassword } from '../api/auth'

const route = useRoute()
const router = useRouter()

const token = computed(() => {
  const value = route.query.token
  return typeof value === 'string' ? value : ''
})

const password = ref('')
const confirm = ref('')
const error = ref<unknown>(null)
const submitting = ref(false)

async function onSubmit() {
  error.value = null
  if (!token.value) {
    error.value = new Error('Reset token is missing. Use the link from your email.')
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
    await resetPassword(token.value, password.value)
    await router.replace({ name: 'login', query: { reset: '1' } })
  } catch (err) {
    error.value = err
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthShell title="Choose a new password" subtitle="Pick something memorable and secure.">
    <p v-if="!token" class="form-error" role="alert">
      This reset link is missing a token. Request a new link from the forgot password page.
    </p>
    <form v-else class="auth-form" @submit.prevent="onSubmit">
      <label class="field">
        <span>New password</span>
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
      <button class="btn-primary" type="submit" :disabled="submitting">
        {{ submitting ? 'Updating…' : 'Update password' }}
      </button>
    </form>
    <p class="auth-links">
      <RouterLink :to="{ name: 'forgot-password' }">Request a new link</RouterLink>
      <RouterLink :to="{ name: 'login' }">Back to sign in</RouterLink>
    </p>
  </AuthShell>
</template>
