<script setup lang="ts">
import { AI_PROVIDERS, openAiProvider } from '../lib/aiProviders'
import AiProviderIcon from './AiProviderIcon.vue'

const props = defineProps<{
  prompt: string
  disabled?: boolean
}>()

function onOpen(provider: (typeof AI_PROVIDERS)[number]) {
  if (props.disabled || !props.prompt.trim()) {
    return
  }
  openAiProvider(provider, props.prompt)
}
</script>

<template>
  <div class="ai-prompt-links" role="group" aria-label="Open prompt in AI chat">
    <button
      v-for="provider in AI_PROVIDERS"
      :key="provider.id"
      class="ai-prompt-links__btn"
      :data-provider="provider.id"
      type="button"
      :disabled="disabled || !prompt.trim()"
      :title="`Open in ${provider.label}`"
      :aria-label="`Open in ${provider.label}`"
      @click="onOpen(provider)"
    >
      <AiProviderIcon :provider="provider.id" />
    </button>
  </div>
</template>
