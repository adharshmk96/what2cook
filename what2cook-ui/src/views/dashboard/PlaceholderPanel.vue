<script setup lang="ts">
defineProps<{
  title: string
  description: string
}>()
</script>

<template>
  <section class="dash-panel" aria-labelledby="panel-title">
    <h1 id="panel-title" class="dash-panel__title">{{ title }}</h1>
    <p class="dash-panel__desc">{{ description }}</p>
    <p class="dash-panel__soon" role="status">Coming soon</p>
  </section>
</template>
