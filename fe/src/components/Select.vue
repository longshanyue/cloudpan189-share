<template>
  <div class="select-container" ref="selectRef">
    <div 
      class="select-trigger" 
      :class="{ 'select-open': isOpen, 'select-disabled': disabled }"
      @click="toggleDropdown"
    >
      <span class="select-value" :class="{ 'select-placeholder': !selectedOption }">
        {{ selectedOption ? selectedOption.label : placeholder }}
      </span>
      <svg class="select-arrow" :class="{ 'select-arrow-up': isOpen }" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
      </svg>
    </div>
    
    <transition name="dropdown">
      <div v-if="isOpen" class="select-dropdown">
        <div 
          v-for="option in options" 
          :key="option.value"
          class="select-option"
          :class="{ 'select-option-selected': option.value == modelValue }"
          @click="selectOption(option)"
        >
          {{ option.label }}
        </div>
        <div v-if="options.length === 0" class="select-empty">
          {{ emptyText }}
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface Option {
  label: string
  value: string | number
}

interface Props {
  modelValue?: string | number
  options: Option[]
  placeholder?: string
  disabled?: boolean
  emptyText?: string
}

interface Emits {
  (e: 'update:modelValue', value: string | number): void
  (e: 'change', value: string | number): void
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '请选择',
  disabled: false,
  emptyText: '暂无数据'
})

const emit = defineEmits<Emits>()

const isOpen = ref(false)
const selectRef = ref<HTMLElement>()

const selectedOption = computed(() => {
  // 使用宽松比较来查找选中的选项
  return props.options.find(option => option.value == props.modelValue)
})

const toggleDropdown = () => {
  if (props.disabled) return
  isOpen.value = !isOpen.value
}

const selectOption = (option: Option) => {
  emit('update:modelValue', option.value)
  emit('change', option.value)
  isOpen.value = false
}

const handleClickOutside = (event: Event) => {
  if (selectRef.value && !selectRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.select-container {
  position: relative;
  width: 100%;
}

.select-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  background: white;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
  font-size: 0.875rem;
}

.select-trigger:hover:not(.select-disabled) {
  border-color: #9ca3af;
}

.select-trigger.select-open {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.select-trigger.select-disabled {
  background: #f9fafb;
  color: #9ca3af;
  cursor: not-allowed;
}

.select-value {
  flex: 1;
  text-align: left;
  color: #374151;
}

.select-placeholder {
  color: #9ca3af;
}

.select-arrow {
  width: 1.25rem;
  height: 1.25rem;
  color: #6b7280;
  transition: transform 0.2s;
  flex-shrink: 0;
}

.select-arrow-up {
  transform: rotate(180deg);
}

.select-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  z-index: 9999;
  margin-top: 0.25rem;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  max-height: 200px;
  overflow-y: auto;
}

.select-option {
  padding: 0.75rem;
  cursor: pointer;
  transition: background-color 0.2s;
  font-size: 0.875rem;
  color: #374151;
}

.select-option:hover {
  background: #f9fafb;
}

.select-option-selected {
  background: #eff6ff;
  color: #3b82f6;
  font-weight: 500;
}

.select-option-selected:hover {
  background: #dbeafe;
}

.select-empty {
  padding: 0.75rem;
  text-align: center;
  color: #9ca3af;
  font-size: 0.875rem;
}

/* 下拉动画 */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from {
  opacity: 0;
  transform: translateY(-0.5rem);
}

.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-0.5rem);
}
</style>