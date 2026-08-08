<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<template>
  <div class="auth-shell">
    <div class="auth-atmosphere" aria-hidden="true" />
    <div class="auth-compose">
      <p class="brand">what2cook</p>
      <h1 class="auth-title">{{ title }}</h1>
      <p v-if="subtitle" class="auth-subtitle">{{ subtitle }}</p>
      <div class="auth-body">
        <slot />
      </div>
    </div>
  </div>
</template>
