<template>
  <div class="family-file-tree-selector">
    <div v-if="loading" class="tree-loading">
      <div class="loading-spinner"></div>
      <span>加载文件列表中...</span>
    </div>
    <div v-else-if="files.length === 0" class="tree-empty">
      暂无文件
    </div>
    <div v-else class="tree-container">
      <FileTreeNode
        v-for="file in files"
        :key="file.id"
        :node="file"
        :selected-id="selectedId"
        :expanded-ids="expandedIds"
        :loading-ids="loadingIds"
        @select="handleSelect"
        @toggle="handleToggle"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import FileTreeNode from './FileTreeNode.vue'
import { storageBridgeApi, type FileNode } from '@/api/storage_bridge'
import { toast } from '@/utils/toast'

interface TreeFileNode extends FileNode {
  children?: TreeFileNode[]
  level: number
}

interface Props {
  cloudToken?: string | number
  familyId?: string
  modelValue?: string
}

interface Emits {
  (e: 'update:modelValue', value: string): void
  (e: 'select', file: TreeFileNode): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const files = ref<TreeFileNode[]>([])
const loading = ref(false)
const selectedId = ref<string>('')
const expandedIds = ref<Set<string>>(new Set())
const loadingIds = ref<Set<string>>(new Set())

// 获取家庭文件列表
const fetchFiles = async () => {
  if (!props.cloudToken || !props.familyId) return
  
  try {
    loading.value = true
    const response = await storageBridgeApi.getFamilyNodes({
      cloudToken: Number(props.cloudToken),
      familyId: props.familyId,
      id: '' // 根目录
    })
    
    // 只显示文件夹，过滤掉文件
    files.value = (response.data || [])
      .filter(file => file.isFolder)
      .map(file => ({
        ...file,
        level: 0,
        children: undefined
      }))
  } catch (error) {
    console.error('获取家庭文件列表失败:', error)
    toast.error('获取家庭文件列表失败')
    files.value = []
  } finally {
    loading.value = false
  }
}

// 监听 modelValue 变化
watch(() => props.modelValue, (newValue) => {
  if (newValue !== selectedId.value) {
    selectedId.value = newValue || ''
  }
}, { immediate: true })

// 监听 cloudToken 和 familyId 变化
watch(() => [props.cloudToken, props.familyId], ([newToken, newFamilyId]) => {
  if (newToken && newFamilyId) {
    fetchFiles()
  } else {
    files.value = []
    selectedId.value = ''
    expandedIds.value.clear()
    loadingIds.value.clear()
  }
}, { immediate: true })

// 处理选择
const handleSelect = (file: TreeFileNode) => {
  if (!file.isFolder) return
  
  selectedId.value = file.id
  emit('update:modelValue', file.id)
  emit('select', file)
}

// 处理展开/折叠
const handleToggle = async (file: TreeFileNode) => {
  if (!file.isFolder) return
  
  const fileId = file.id
  
  if (expandedIds.value.has(fileId)) {
    // 折叠
    expandedIds.value.delete(fileId)
  } else {
    // 展开
    expandedIds.value.add(fileId)
    
    // 如果没有子文件，则加载
    if (!file.children) {
      await loadChildren(file)
    }
  }
}

// 加载子文件
const loadChildren = async (parentFile: TreeFileNode) => {
  if (!props.cloudToken || !props.familyId || loadingIds.value.has(parentFile.id)) return
  
  try {
    loadingIds.value.add(parentFile.id)
    
    const response = await storageBridgeApi.getFamilyNodes({
      cloudToken: Number(props.cloudToken),
      familyId: props.familyId,
      id: parentFile.id
    })
    
    // 只显示文件夹，过滤掉文件
    const children = (response.data || [])
      .filter(file => file.isFolder)
      .map(file => ({
        ...file,
        level: parentFile.level + 1,
        children: undefined
      }))
    
    // 更新父文件的子文件列表
    parentFile.children = children
    
    // 触发响应式更新
    files.value = [...files.value]
    
  } catch (error) {
    console.error('加载家庭子文件失败:', error)
    toast.error('加载家庭子文件失败')
  } finally {
    loadingIds.value.delete(parentFile.id)
  }
}

// 查找文件节点
const findFileNode = (files: TreeFileNode[], id: string): TreeFileNode | null => {
  for (const file of files) {
    if (file.id === id) return file
    if (file.children) {
      const found = findFileNode(file.children, id)
      if (found) return found
    }
  }
  return null
}

// 获取选中的文件
const getSelectedFile = (): TreeFileNode | null => {
  if (!selectedId.value) return null
  return findFileNode(files.value, selectedId.value)
}

// 暴露方法
defineExpose({
  getSelectedFile,
  refresh: fetchFiles
})
</script>

<style scoped>
.family-file-tree-selector {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: white;
  max-height: 300px;
  overflow-y: auto;
}

.tree-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 2rem;
  color: #6b7280;
  font-size: 0.875rem;
}

.loading-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid #e5e7eb;
  border-top: 2px solid #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.tree-empty {
  padding: 2rem;
  text-align: center;
  color: #9ca3af;
  font-size: 0.875rem;
}

.tree-container {
  padding: 0.5rem 0;
}
</style>