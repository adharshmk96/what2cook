<script setup lang="ts">
import { computed } from 'vue'
import { ApiError } from '../api/client'

const props = defineProps<{
  error: unknown
}>()

const message = computed(() => {
  const err = props.error
  if (!err) return ''
  if (err instanceof ApiError) return err.message
  if (err instanceof Error) return err.message
  return 'Something went wrong. Please try again.'
})
</script>

<template>
  <p v-if="message" class="form-error" role="alert">{{ message }}</p>
</template>
