<template>
  <div class="pagination-container">
    <div class="pagination-left">
      <!-- 左侧留空 -->
    </div>
    
    <div class="pagination-right">
      <div class="pagination-info">
        共 {{ total }} 条记录
      </div>
      
      <div class="pagination-controls">
        <button 
          class="pagination-btn" 
          :disabled="currentPage <= 1"
          @click="handlePageChange(currentPage - 1)"
        >
          上一页
        </button>
        
        <div class="page-numbers">
          <button 
            v-for="page in visiblePages" 
            :key="page"
            class="page-btn"
            :class="{ 'page-btn-active': page === currentPage }"
            @click="handlePageChange(page)"
            :disabled="page === -1"
          >
            {{ page === -1 ? '...' : page }}
          </button>
        </div>
        
        <button 
          class="pagination-btn" 
          :disabled="currentPage >= totalPages"
          @click="handlePageChange(currentPage + 1)"
        >
          下一页
        </button>
      </div>
      
      <div class="page-size-selector">
        <span class="page-size-label">每页显示</span>
        <Select 
          :model-value="pageSize" 
          :options="pageSizeOptions" 
          @update:model-value="handlePageSizeChange"
        />
        <span class="page-size-label">条</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Select from './Select.vue'

interface Props {
  currentPage: number
  pageSize: number
  total: number
}

interface Emits {
  (e: 'page-change', page: number): void
  (e: 'page-size-change', size: number): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const totalPages = computed(() => {
  return Math.ceil(props.total / props.pageSize)
})

const visiblePages = computed(() => {
  const pages: number[] = []
  const total = totalPages.value
  const current = props.currentPage
  
  if (total <= 7) {
    // 如果总页数小于等于7，显示所有页码
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    // 总是显示第一页
    pages.push(1)
    
    if (current <= 4) {
      // 当前页在前面
      for (let i = 2; i <= 5; i++) {
        pages.push(i)
      }
      pages.push(-1) // 省略号
      pages.push(total)
    } else if (current >= total - 3) {
      // 当前页在后面
      pages.push(-1) // 省略号
      for (let i = total - 4; i <= total; i++) {
        pages.push(i)
      }
    } else {
      // 当前页在中间
      pages.push(-1) // 省略号
      for (let i = current - 1; i <= current + 1; i++) {
        pages.push(i)
      }
      pages.push(-1) // 省略号
      pages.push(total)
    }
  }
  
  return pages
})

const pageSizeOptions = [
  { label: '10', value: 10 },
  { label: '15', value: 15 },
  { label: '30', value: 30 },
  { label: '50', value: 50 },
  { label: '100', value: 100 }
]

const handlePageChange = (page: number) => {
  if (page >= 1 && page <= totalPages.value && page !== props.currentPage) {
    emit('page-change', page)
  }
}

const handlePageSizeChange = (size: string | number) => {
  const numSize = typeof size === 'string' ? parseInt(size) : size
  emit('page-size-change', numSize)
}
</script>

<style scoped>
.pagination-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 0;
  gap: 1rem;
  flex-wrap: wrap;
}

.pagination-left {
  flex: 1;
}

.pagination-right {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.pagination-info {
  color: #6b7280;
  font-size: 0.875rem;
  white-space: nowrap;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.pagination-btn {
  padding: 0.5rem 1rem;
  border: 1px solid #d1d5db;
  background: white;
  color: #374151;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
  white-space: nowrap;
}

.pagination-btn:hover:not(:disabled) {
  background: #f9fafb;
  border-color: #9ca3af;
}

.pagination-btn:disabled {
  background: #f9fafb;
  color: #9ca3af;
  cursor: not-allowed;
}

.page-numbers {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.page-btn {
  min-width: 2.5rem;
  height: 2.5rem;
  padding: 0;
  border: 1px solid #d1d5db;
  background: white;
  color: #374151;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.page-btn:hover {
  background: #f9fafb;
  border-color: #9ca3af;
}

.page-btn-active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.page-btn-active:hover {
  background: #2563eb;
  border-color: #2563eb;
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  white-space: nowrap;
  overflow: visible;
}

.page-size-label {
  color: #6b7280;
  font-size: 0.875rem;
}

.page-size-selector :deep(.select-container) {
  width: 80px;
  overflow: visible;
}

@media (max-width: 768px) {
  .pagination-container {
    flex-direction: column;
    align-items: stretch;
    gap: 0.75rem;
  }
  
  .pagination-controls {
    justify-content: center;
  }
  
  .page-size-selector {
    justify-content: center;
  }
}
</style>