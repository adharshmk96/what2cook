<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

type Crumb = {
  label: string
  to?: { name: string }
}

const route = useRoute()

const crumbs = computed<Crumb[]>(() => {
  const trail: Crumb[] = [{ label: 'Dashboard', to: { name: 'dashboard' } }]

  const meta = route.meta.breadcrumb
  if (Array.isArray(meta)) {
    for (const item of meta) {
      if (
        item &&
        typeof item === 'object' &&
        'label' in item &&
        typeof (item as { label: unknown }).label === 'string'
      ) {
        const crumb = item as { label: string; name?: string }
        trail.push({
          label: crumb.label,
          to: crumb.name ? { name: crumb.name } : undefined,
        })
      }
    }
  } else if (typeof meta === 'string' && meta.trim()) {
    trail.push({ label: meta })
  }

  return trail
})
</script>

<template>
  <nav v-if="crumbs.length > 1" class="breadcrumb" aria-label="Breadcrumb">
    <ol class="breadcrumb__list">
      <li v-for="(crumb, index) in crumbs" :key="`${crumb.label}-${index}`" class="breadcrumb__item">
        <RouterLink
          v-if="crumb.to && index < crumbs.length - 1"
          class="breadcrumb__link"
          :to="crumb.to"
        >
          {{ crumb.label }}
        </RouterLink>
        <span v-else class="breadcrumb__current" aria-current="page">{{ crumb.label }}</span>
        <span
          v-if="index < crumbs.length - 1"
          class="breadcrumb__sep"
          aria-hidden="true"
        >/</span>
      </li>
    </ol>
  </nav>
</template>
