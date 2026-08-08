<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthShell from '../layouts/AuthShell.vue'
import FormError from '../components/FormError.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const oldPassword = ref('')
const newPassword = ref('')
const confirm = ref('')
const error = ref<unknown>(null)
const success = ref('')
const submitting = ref(false)

async function onSubmit() {
  error.value = null
  success.value = ''
  if (!oldPassword.value || !newPassword.value) {
    error.value = new Error('Both passwords are required.')
    return
  }
  if (newPassword.value.length < 8) {
    error.value = new Error('New password must be at least 8 characters.')
    return
  }
  if (newPassword.value !== confirm.value) {
    error.value = new Error('New passwords do not match.')
    return
  }
  if (oldPassword.value === newPassword.value) {
    error.value = new Error('New password must be different from the current one.')
    return
  }

  submitting.value = true
  try {
    await auth.changePassword(oldPassword.value, newPassword.value)
    success.value = 'Password updated.'
    oldPassword.value = ''
    newPassword.value = ''
    confirm.value = ''
  } catch (err) {
    error.value = err
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthShell title="Change password" subtitle="Update the password for your what2cook account.">
    <form class="auth-form" @submit.prevent="onSubmit">
      <label class="field">
        <span>Current password</span>
        <input
          v-model="oldPassword"
          type="password"
          name="oldPassword"
          autocomplete="current-password"
          required
          placeholder="Current password"
        />
      </label>
      <label class="field">
        <span>New password</span>
        <input
          v-model="newPassword"
          type="password"
          name="newPassword"
          autocomplete="new-password"
          required
          minlength="8"
          placeholder="At least 8 characters"
        />
      </label>
      <label class="field">
        <span>Confirm new password</span>
        <input
          v-model="confirm"
          type="password"
          name="confirm"
          autocomplete="new-password"
          required
          minlength="8"
          placeholder="Repeat new password"
        />
      </label>
      <FormError :error="error" />
      <p v-if="success" class="form-success" role="status">{{ success }}</p>
      <button class="btn-primary" type="submit" :disabled="submitting || auth.loading">
        {{ submitting ? 'Saving…' : 'Save password' }}
      </button>
    </form>
    <p class="auth-links">
      <button type="button" class="link-button" @click="router.push({ name: 'home' })">
        Back home
      </button>
    </p>
  </AuthShell>
</template>
