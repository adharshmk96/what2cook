<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import FormError from '../../components/FormError.vue'
import { useAuthStore } from '../../stores/auth'

const auth = useAuthStore()
const activeTab = ref<'account' | 'api-key'>('account')

const tabs = [
  { id: 'account', label: 'Account', comingSoon: false },
  { id: 'api-key', label: 'API Key', comingSoon: true },
] as const

const email = ref(auth.user?.email ?? '')
const emailError = ref<unknown>(null)
const emailSuccess = ref('')
const emailSubmitting = ref(false)

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordError = ref<unknown>(null)
const passwordSuccess = ref('')
const passwordSubmitting = ref(false)

const verifyError = ref<unknown>(null)
const verifySuccess = ref('')
const verifySubmitting = ref(false)

const isVerified = computed(() => Boolean(auth.user?.email_verified))

watch(
  () => auth.user?.email,
  (next) => {
    if (typeof next === 'string') {
      email.value = next
    }
  },
)

async function onUpdateEmail() {
  emailError.value = null
  emailSuccess.value = ''
  const trimmed = email.value.trim()
  if (!trimmed) {
    emailError.value = new Error('Email is required.')
    return
  }

  emailSubmitting.value = true
  try {
    await auth.updateEmail(trimmed)
    emailSuccess.value = 'Email updated. Check your inbox (or app logs) for a new verification link.'
  } catch (err) {
    emailError.value = err
  } finally {
    emailSubmitting.value = false
  }
}

async function onChangePassword() {
  passwordError.value = null
  passwordSuccess.value = ''
  if (!oldPassword.value || !newPassword.value) {
    passwordError.value = new Error('Both passwords are required.')
    return
  }
  if (newPassword.value.length < 8) {
    passwordError.value = new Error('New password must be at least 8 characters.')
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = new Error('New passwords do not match.')
    return
  }
  if (oldPassword.value === newPassword.value) {
    passwordError.value = new Error('New password must be different from the current one.')
    return
  }

  passwordSubmitting.value = true
  try {
    await auth.changePassword(oldPassword.value, newPassword.value)
    passwordSuccess.value = 'Password updated.'
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (err) {
    passwordError.value = err
  } finally {
    passwordSubmitting.value = false
  }
}

async function onResendVerification() {
  verifyError.value = null
  verifySuccess.value = ''
  verifySubmitting.value = true
  try {
    await auth.resendVerification()
    verifySuccess.value = 'Verification email sent. If SMTP is not configured, check the app logs for the link.'
  } catch (err) {
    verifyError.value = err
  } finally {
    verifySubmitting.value = false
  }
}
</script>

<template>
  <div class="account-settings">
    <header class="dash-panel dash-panel--flush">
      <h1 class="dash-panel__title">Settings</h1>
      <p class="dash-panel__desc">Manage your account and what2cook preferences.</p>
    </header>

    <div class="settings-tabs" role="tablist" aria-label="Settings sections">
      <button
        v-for="tab in tabs"
        :id="`settings-tab-${tab.id}`"
        :key="tab.id"
        type="button"
        class="settings-tabs__tab"
        :class="{ 'is-active': activeTab === tab.id }"
        role="tab"
        :aria-selected="activeTab === tab.id"
        :aria-controls="`settings-panel-${tab.id}`"
        @click="activeTab = tab.id"
      >
        {{ tab.label }}
        <span v-if="tab.comingSoon" class="settings-tabs__badge">Coming soon</span>
      </button>
    </div>

    <div
      v-if="activeTab !== 'account'"
      :id="`settings-panel-${activeTab}`"
      class="dash-panel settings-coming-soon"
      role="tabpanel"
      :aria-labelledby="`settings-tab-${activeTab}`"
    >
      <p class="dash-panel__soon">Coming soon</p>
      <h2 class="dash-section-title">API Key</h2>
      <p class="dash-panel__desc">
        API key creation and management will be available here.
      </p>
    </div>

    <div
      v-else
      id="settings-panel-account"
      role="tabpanel"
      aria-labelledby="settings-tab-account"
    >
      <section class="dash-panel" aria-labelledby="email-section-title">
      <h2 id="email-section-title" class="dash-section-title">Email</h2>
      <form class="auth-form" @submit.prevent="onUpdateEmail">
        <label class="field">
          <span>Email address</span>
          <input
            v-model="email"
            type="email"
            name="email"
            autocomplete="email"
            required
            placeholder="you@example.com"
          />
        </label>
        <FormError :error="emailError" />
        <p v-if="emailSuccess" class="form-success" role="status">{{ emailSuccess }}</p>
        <button class="btn-primary" type="submit" :disabled="emailSubmitting || auth.loading">
          {{ emailSubmitting ? 'Saving…' : 'Update email' }}
        </button>
      </form>
    </section>

    <section class="dash-panel" aria-labelledby="verify-section-title">
      <h2 id="verify-section-title" class="dash-section-title">Email verification</h2>
      <p v-if="isVerified" class="form-success" role="status">Your email is verified.</p>
      <template v-else>
        <p class="dash-panel__desc">
          Your email is not verified yet. Use the link from your inbox, or resend it below.
        </p>
        <FormError :error="verifyError" />
        <p v-if="verifySuccess" class="form-success" role="status">{{ verifySuccess }}</p>
        <button
          class="btn-primary"
          type="button"
          :disabled="verifySubmitting || auth.loading"
          @click="onResendVerification"
        >
          {{ verifySubmitting ? 'Sending…' : 'Resend verification' }}
        </button>
      </template>
    </section>

    <section class="dash-panel" aria-labelledby="password-section-title">
      <h2 id="password-section-title" class="dash-section-title">Change password</h2>
      <form class="auth-form" @submit.prevent="onChangePassword">
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
            v-model="confirmPassword"
            type="password"
            name="confirmPassword"
            autocomplete="new-password"
            required
            minlength="8"
            placeholder="Repeat new password"
          />
        </label>
        <FormError :error="passwordError" />
        <p v-if="passwordSuccess" class="form-success" role="status">{{ passwordSuccess }}</p>
        <button class="btn-primary" type="submit" :disabled="passwordSubmitting || auth.loading">
          {{ passwordSubmitting ? 'Saving…' : 'Save password' }}
        </button>
      </form>
      </section>
    </div>
  </div>
</template>
