<template>
  <div class="transfer-container">
    <div class="transfer-panel">
      <div class="transfer-header">
        <h4>{{ leftTitle }}</h4>
        <span class="count">{{ availableItems.length }} 项</span>
      </div>
      <div class="transfer-body">
        <div v-if="loading" class="loading-state">
          <div class="loading-spinner"></div>
          <p>加载中...</p>
        </div>
        <div v-else-if="availableItems.length === 0" class="empty-state">
          <p>{{ leftEmptyText }}</p>
        </div>
        <div v-else class="item-list">
          <div 
            v-for="item in availableItems" 
            :key="item[itemKey]" 
            class="item"
            :class="{ selected: selectedAvailable.includes(item[itemKey]) }"
            @click="selectAvailable(item)"
          >
            <span class="item-name">{{ item[nameField] }}</span>
            <span v-if="descField" class="item-desc">{{ item[descField] }}</span>
          </div>
        </div>
      </div>
    </div>
    
    <div class="transfer-controls">
      <button 
        @click="moveToSelected" 
        :disabled="selectedAvailable.length === 0"
        class="btn btn-sm btn-primary"
      >
        >
      </button>
      <button 
        @click="moveToAvailable" 
        :disabled="selectedBound.length === 0"
        class="btn btn-sm btn-primary"
      >
        <
      </button>
    </div>
    
    <div class="transfer-panel">
      <div class="transfer-header">
        <h4>{{ rightTitle }}</h4>
        <span class="count">{{ boundItems.length }} 项</span>
      </div>
      <div class="transfer-body">
        <div v-if="boundItems.length === 0" class="empty-state">
          <p>{{ rightEmptyText }}</p>
        </div>
        <div v-else class="item-list">
          <div 
            v-for="item in boundItems" 
            :key="item[itemKey]" 
            class="item"
            :class="{ selected: selectedBound.includes(item[itemKey]) }"
            @click="selectBound(item)"
          >
            <span class="item-name">{{ item[nameField] }}</span>
            <span v-if="descField" class="item-desc">{{ item[descField] }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

interface TransferItem {
  [key: string]: any
}

interface Props {
  allItems: TransferItem[]
  boundItemIds: (string | number)[]
  itemKey: string
  nameField: string
  descField?: string
  leftTitle: string
  rightTitle: string
  leftEmptyText?: string
  rightEmptyText?: string
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  leftEmptyText: '暂无可选项',
  rightEmptyText: '暂无已选项',
  loading: false
})

const emit = defineEmits<{
  change: [boundIds: (string | number)[]]
}>()

const selectedAvailable = ref<(string | number)[]>([])
const selectedBound = ref<(string | number)[]>([])

// 计算可选和已选项目
const availableItems = computed(() => {
  return props.allItems.filter(item => !props.boundItemIds.includes(item[props.itemKey]))
})

const boundItems = computed(() => {
  return props.allItems.filter(item => props.boundItemIds.includes(item[props.itemKey]))
})

// 选择项目
const selectAvailable = (item: TransferItem) => {
  const id = item[props.itemKey]
  const index = selectedAvailable.value.indexOf(id)
  if (index > -1) {
    selectedAvailable.value.splice(index, 1)
  } else {
    selectedAvailable.value.push(id)
  }
}

const selectBound = (item: TransferItem) => {
  const id = item[props.itemKey]
  const index = selectedBound.value.indexOf(id)
  if (index > -1) {
    selectedBound.value.splice(index, 1)
  } else {
    selectedBound.value.push(id)
  }
}

// 移动项目
const moveToSelected = () => {
  const newBoundIds = [...props.boundItemIds, ...selectedAvailable.value]
  selectedAvailable.value = []
  emit('change', newBoundIds)
}

const moveToAvailable = () => {
  const newBoundIds = props.boundItemIds.filter(id => !selectedBound.value.includes(id))
  selectedBound.value = []
  emit('change', newBoundIds)
}

// 监听boundItemIds变化，清空选择
watch(() => props.boundItemIds, () => {
  selectedAvailable.value = []
  selectedBound.value = []
})
</script>

<style scoped>
/* 穿梭框样式 */
.transfer-container {
  display: flex;
  gap: 1rem;
  align-items: stretch;
  min-height: 300px;
}

.transfer-panel {
  flex: 1;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
}

.transfer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
  border-radius: 8px 8px 0 0;
}

.transfer-header h4 {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
}

.transfer-header .count {
  font-size: 0.75rem;
  color: #6b7280;
  background: #e5e7eb;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
}

.transfer-body {
  flex: 1;
  overflow-y: auto;
  max-height: 250px;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: #6b7280;
}

.loading-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid #e5e7eb;
  border-top: 2px solid #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 0.5rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: #6b7280;
  font-size: 0.875rem;
}

.item-list {
  padding: 0.5rem;
}

.item {
  display: flex;
  flex-direction: column;
  padding: 0.75rem;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  margin-bottom: 0.5rem;
  cursor: pointer;
  transition: all 0.2s;
}

.item:hover {
  background: #f9fafb;
  border-color: #3b82f6;
}

.item.selected {
  background: #eff6ff;
  border-color: #3b82f6;
}

.item-name {
  font-weight: 500;
  color: #1f2937;
  font-size: 0.875rem;
  margin-bottom: 0.25rem;
}

.item-desc {
  font-size: 0.75rem;
  color: #6b7280;
  word-break: break-all;
}

.transfer-controls {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.5rem;
  padding: 0 0.5rem;
}

.transfer-controls .btn {
  min-width: 40px;
  height: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn-sm {
  padding: 0.5rem 0.75rem;
  font-size: 0.75rem;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .transfer-container {
    flex-direction: column;
    min-height: auto;
  }
  
  .transfer-controls {
    flex-direction: row;
    justify-content: center;
    padding: 0.5rem 0;
  }
  
  .transfer-body {
    max-height: 200px;
  }
}
</style>