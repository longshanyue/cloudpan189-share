<template>
  <div class="file-tree-node">
    <!-- 当前节点 -->
    <div 
      class="node-item"
      :class="{
        'selected': selectedId === node.id,
        'disabled': !node.isFolder
      }"
      :style="{ paddingLeft: `${node.level * 16 + 8}px` }"
      @click="handleClick"
    >
      <!-- 展开图标 -->
      <div class="expand-icon" @click.stop="handleToggle">
        <template v-if="node.isFolder">
          <Icons 
            v-if="loadingIds.has(node.id)"
            name="loading" 
            size="14px"
            class="loading-icon"
          />
          <span
            v-else
            class="expand-symbol"
          >
            {{ expandedIds.has(node.id) ? '-' : '+' }}
          </span>
        </template>
      </div>
      
      <!-- 文件图标和名称 -->
      <div class="node-content">
        <Icons 
          :name="node.isFolder ? 'folder' : 'file'" 
          size="16px"
          class="file-icon"
          :class="{ 'folder-icon': node.isFolder }"
        />
        <span class="file-name">{{ node.name }}</span>
      </div>
    </div>
    
    <!-- 子节点 -->
    <template v-if="node.children && expandedIds.has(node.id)">
      <FileTreeNode
        v-for="child in node.children"
        :key="child.id"
        :node="child"
        :selected-id="selectedId"
        :expanded-ids="expandedIds"
        :loading-ids="loadingIds"
        @select="$emit('select', $event)"
        @toggle="$emit('toggle', $event)"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import Icons from './Icons.vue'
import type { FileNode } from '@/api/storage_bridge'

interface TreeFileNode extends FileNode {
  children?: TreeFileNode[]
  level: number
}

interface Props {
  node: TreeFileNode
  selectedId: string
  expandedIds: Set<string>
  loadingIds: Set<string>
}

interface Emits {
  (e: 'select', node: TreeFileNode): void
  (e: 'toggle', node: TreeFileNode): void
}

const emit = defineEmits<Emits>()

const handleClick = () => {
  emit('select', props.node)
}

const handleToggle = () => {
  if (props.node.isFolder) {
    emit('toggle', props.node)
  }
}

// 获取 props 以便在方法中使用
const props = defineProps<Props>()
</script>

<style scoped>
.file-tree-node {
  user-select: none;
}

.node-item {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border-radius: 4px;
  margin: 1px 4px;
}

.node-item:hover {
  background-color: #f8fafc;
}

.node-item.selected {
  background-color: #eff6ff;
  border: 1px solid #3b82f6;
}

.node-item.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.node-item.disabled:hover {
  background-color: transparent;
}

.expand-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 2px;
  transition: background-color 0.2s ease;
  flex-shrink: 0;
}

.expand-icon:hover {
  background-color: #e5e7eb;
}

.loading-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.expand-symbol {
  font-size: 12px;
  font-weight: bold;
  color: #6b7280;
  user-select: none;
  line-height: 1;
}

.expand-icon:hover .expand-symbol {
  color: #374151;
}

.node-content {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  min-width: 0;
}

.file-icon {
  flex-shrink: 0;
  color: #6b7280;
}

.folder-icon {
  color: #f59e0b;
}

.node-item.selected .folder-icon {
  color: #3b82f6;
}

.file-name {
  font-size: 14px;
  color: #374151;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-item.selected .file-name {
  color: #1d4ed8;
  font-weight: 500;
}

.node-item.disabled .file-name {
  color: #9ca3af;
}
</style>