<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { X } from 'lucide-vue-next'

const props = defineProps<{
  open: boolean
  title: string
  titleId?: string
}>()

const emit = defineEmits<{
  close: []
}>()

const resolvedTitleId = props.titleId ?? 'modal-title'

function onBackdropClick(event: MouseEvent) {
  if (event.target === event.currentTarget) {
    emit('close')
  }
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && props.open) {
    emit('close')
  }
}

watch(
  () => props.open,
  (isOpen) => {
    document.body.style.overflow = isOpen ? 'hidden' : ''
  },
)

onMounted(() => {
  document.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
})
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-backdrop"
      @click="onBackdropClick"
    >
      <div
        class="modal"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="resolvedTitleId"
        @click.stop
      >
        <header class="modal__header">
          <h2 :id="resolvedTitleId" class="modal__title">{{ title }}</h2>
          <button
            class="modal__close"
            type="button"
            aria-label="Close"
            @click="emit('close')"
          >
            <X class="icon icon--sm" aria-hidden="true" />
          </button>
        </header>
        <div class="modal__body">
          <slot />
        </div>
      </div>
    </div>
  </Teleport>
</template>
