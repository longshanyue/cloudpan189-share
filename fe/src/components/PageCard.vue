<template>
  <div class="page-section">
    <div class="page-card">
      <div class="card-header">
        <h1 class="card-title">{{ title }}</h1>
        <p class="card-subtitle">{{ subtitle }}</p>
      </div>
      
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title: string
  subtitle: string
}

defineProps<Props>()
</script>

<style scoped>
/* 页面区域 */
.page-section {
  margin-bottom: 1.5rem;
}

.page-card {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid #f1f5f9;
}

/* 卡片头部样式 */
.card-header {
  text-align: left;
  margin-bottom: 1.5rem;
}

.card-title {
  font-size: 1.75rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 0.5rem 0;
}

.card-subtitle {
  color: #6b7280;
  margin: 0;
  font-size: 0.95rem;
  line-height: 1.5;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .page-card {
    padding: 1.5rem;
    margin: 0 1rem;
  }
  
  .card-title {
    font-size: 1.5rem;
  }
  
  .card-subtitle {
    font-size: 0.875rem;
  }
}

@media (max-width: 480px) {
  .page-card {
    padding: 1rem;
  }
  
  .card-title {
    font-size: 1.25rem;
  }
}
</style>