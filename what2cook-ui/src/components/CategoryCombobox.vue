<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { ChevronDown, Plus } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    modelValue: string
    options: string[]
    disabled?: boolean
    placeholder?: string
    maxlength?: number
    inputId?: string
    listId?: string
    ariaLabel?: string
  }>(),
  {
    disabled: false,
    placeholder: 'e.g. Meat, Vegetables, Spices',
    maxlength: 60,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const rootRef = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const open = ref(false)
const activeIndex = ref(-1)
const query = ref(props.modelValue)

watch(
  () => props.modelValue,
  (value) => {
    if (value !== query.value) query.value = value
  },
)

const normalizedQuery = computed(() => query.value.trim())

const filteredOptions = computed(() => {
  const needle = normalizedQuery.value.toLowerCase()
  if (!needle) return props.options
  return props.options.filter((option) =>
    option.toLowerCase().includes(needle),
  )
})

const exactMatch = computed(() =>
  props.options.some(
    (option) => option.toLowerCase() === normalizedQuery.value.toLowerCase(),
  ),
)

const showCreateOption = computed(
  () => normalizedQuery.value.length > 0 && !exactMatch.value,
)

const listItems = computed(() => {
  const items: Array<{ type: 'option' | 'create'; value: string }> = []

  for (const option of filteredOptions.value) {
    items.push({ type: 'option', value: option })
  }

  if (showCreateOption.value) {
    items.push({ type: 'create', value: normalizedQuery.value })
  }

  return items
})

const showList = computed(
  () => open.value && !props.disabled && listItems.value.length > 0,
)

function openList() {
  if (props.disabled) return
  open.value = true
  activeIndex.value = -1
}

function closeList() {
  open.value = false
  activeIndex.value = -1
}

function selectValue(value: string) {
  query.value = value
  emit('update:modelValue', value)
  closeList()
  inputRef.value?.focus()
}

function onInput(event: Event) {
  const value = (event.target as HTMLInputElement).value
  query.value = value
  emit('update:modelValue', value)
  openList()
}

function onFocus() {
  openList()
}

function onBlur() {
  window.setTimeout(() => {
    closeList()
  }, 150)
}

function onKeydown(event: KeyboardEvent) {
  if (!open.value && (event.key === 'ArrowDown' || event.key === 'ArrowUp')) {
    openList()
    event.preventDefault()
    return
  }

  const items = listItems.value
  if (!open.value || items.length === 0) {
    if (event.key === 'Escape') closeList()
    return
  }

  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      activeIndex.value = (activeIndex.value + 1) % items.length
      break
    case 'ArrowUp':
      event.preventDefault()
      activeIndex.value =
        activeIndex.value <= 0 ? items.length - 1 : activeIndex.value - 1
      break
    case 'Enter':
      if (activeIndex.value >= 0) {
        event.preventDefault()
        selectValue(items[activeIndex.value].value)
      }
      break
    case 'Escape':
      event.preventDefault()
      closeList()
      break
    case 'Tab':
      closeList()
      break
  }
}

function onOptionMouseDown(event: MouseEvent) {
  event.preventDefault()
}

function onClickOutside(event: MouseEvent) {
  if (rootRef.value && !rootRef.value.contains(event.target as Node)) {
    closeList()
  }
}

onMounted(() => {
  document.addEventListener('mousedown', onClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', onClickOutside)
})
</script>

<template>
  <div ref="rootRef" class="category-combobox">
    <div class="category-combobox__input-wrap">
      <input
        :id="inputId"
        ref="inputRef"
        class="category-combobox__input"
        type="text"
        role="combobox"
        autocomplete="off"
        :value="query"
        :placeholder="placeholder"
        :maxlength="maxlength"
        :disabled="disabled"
        :aria-label="ariaLabel"
        :aria-expanded="showList"
        :aria-controls="listId"
        aria-autocomplete="list"
        @input="onInput"
        @focus="onFocus"
        @blur="onBlur"
        @keydown="onKeydown"
      />
      <ChevronDown
        class="category-combobox__chevron icon icon--sm"
        :class="{ 'is-open': showList }"
        aria-hidden="true"
      />
    </div>

    <ul
      v-if="showList"
      :id="listId"
      class="category-combobox__list"
      role="listbox"
    >
      <li
        v-for="(item, index) in listItems"
        :key="`${item.type}-${item.value}`"
        class="category-combobox__option"
        :class="{
          'is-active': index === activeIndex,
          'is-create': item.type === 'create',
        }"
        role="option"
        :aria-selected="index === activeIndex"
        @mousedown="onOptionMouseDown"
        @click="selectValue(item.value)"
      >
        <Plus
          v-if="item.type === 'create'"
          class="icon icon--sm"
          aria-hidden="true"
        />
        <span v-if="item.type === 'create'">
          Create “{{ item.value }}”
        </span>
        <span v-else>{{ item.value }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped>
.category-combobox {
  position: relative;
}

.category-combobox__input-wrap {
  position: relative;
}

.category-combobox__input {
  width: 100%;
  border: 1px solid var(--field-border);
  border-radius: 10px;
  background: var(--field-bg);
  padding: 0.85rem 2.25rem 0.85rem 0.95rem;
  color: var(--ink);
  outline: none;
  transition:
    border-color 180ms ease,
    box-shadow 180ms ease,
    background 180ms ease;
}

.category-combobox__input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(15, 122, 95, 0.18);
  background: rgba(255, 255, 255, 0.92);
}

.category-combobox__input::placeholder {
  color: rgba(77, 99, 90, 0.65);
}

.category-combobox__input:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.category-combobox__chevron {
  position: absolute;
  top: 50%;
  right: 0.85rem;
  transform: translateY(-50%);
  color: var(--muted);
  pointer-events: none;
  transition: transform 160ms ease;
}

.category-combobox__chevron.is-open {
  transform: translateY(-50%) rotate(180deg);
}

.category-combobox__list {
  position: absolute;
  z-index: 20;
  top: calc(100% + 0.35rem);
  left: 0;
  right: 0;
  margin: 0;
  padding: 0.35rem;
  list-style: none;
  border: 1px solid var(--field-border);
  border-radius: 10px;
  background: #fff;
  box-shadow: 0 10px 30px rgba(20, 36, 31, 0.12);
  max-height: 12rem;
  overflow-y: auto;
}

.category-combobox__option {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.55rem 0.65rem;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.9rem;
  color: var(--ink);
}

.category-combobox__option:hover,
.category-combobox__option.is-active {
  background: rgba(15, 122, 95, 0.1);
}

.category-combobox__option.is-create {
  color: var(--accent-deep);
  font-weight: 600;
}
</style>
