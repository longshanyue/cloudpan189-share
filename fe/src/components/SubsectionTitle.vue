<template>
  <h2 class="subsection-title">{{ title }}</h2>
</template>

<script setup lang="ts">
interface Props {
  title: string
}

defineProps<Props>()
</script>

<style scoped>
.subsection-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: #374151;
  margin: 0 0 1rem 0;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid #f3f4f6;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .subsection-title {
    font-size: 1.125rem;
  }
}
</style>