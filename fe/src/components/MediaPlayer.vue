<template>
  <div class="media-player">
    <!-- 视频播放器 -->
    <div v-if="type === 'video'" class="video-container">
      <video
        ref="videoRef"
        class="plyr-video"
        :src="src"
        :poster="poster"
        controls
        playsinline
      >
        您的浏览器不支持视频播放。
      </video>
    </div>

    <!-- 音频播放器 -->
    <div v-else-if="type === 'audio'" class="audio-container">
      <div class="audio-visual">
        <div class="audio-cover">
          <Icons name="music" size="4rem" class="audio-icon" />
        </div>
        <div class="audio-info">
          <h3 class="audio-title">{{ title }}</h3>
          <p class="audio-artist">{{ artist || '未知艺术家' }}</p>
        </div>
      </div>
      <video
        ref="audioRef"
        class="plyr-audio"
        :src="src"
        controls
        style="width: 100%; height: 60px;"
      >
        您的浏览器不支持音频播放。
      </video>
    </div>

    <!-- 图片查看器 -->
    <div v-else-if="type === 'image'" class="image-container">
      <div class="image-viewer">
        <img
          :src="src"
          :alt="title"
          class="preview-image"
          @load="onImageLoad"
          @error="onImageError"
        />
        <div class="image-controls">
          <button @click="zoomIn" class="control-btn" title="放大">
            <Icons name="zoom-in" size="1rem" />
          </button>
          <button @click="zoomOut" class="control-btn" title="缩小">
            <Icons name="zoom-out" size="1rem" />
          </button>
          <button @click="resetZoom" class="control-btn" title="重置">
            <Icons name="refresh" size="1rem" />
          </button>
          <button @click="downloadImage" class="control-btn" title="下载">
            <Icons name="download" size="1rem" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import Icons from './Icons.vue'
import Plyr from 'plyr'
import 'plyr/dist/plyr.css'

interface Props {
  type: 'video' | 'audio' | 'image'
  src: string
  title?: string
  poster?: string
  artist?: string
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  poster: '',
  artist: ''
})

const emit = defineEmits<{
  loaded: []
  error: [error: Error]
}>()

const videoRef = ref<HTMLVideoElement>()
const audioRef = ref<HTMLVideoElement>()
let player: Plyr | undefined = undefined

// 图片缩放相关
const imageScale = ref(1)
const imageTransform = ref('scale(1)')

// 初始化播放器
const initPlayer = () => {
  if (props.type === 'video' && videoRef.value) {
    player = new Plyr(videoRef.value, {
      controls: [
        'play-large',
        'play',
        'progress',
        'current-time',
        'duration',
        'mute',
        'volume',
        'settings',
        'fullscreen'
      ],
      settings: ['quality', 'speed'],
      quality: {
        default: 720,
        options: [1080, 720, 480, 360]
      },
      speed: {
        selected: 1,
        options: [0.5, 0.75, 1, 1.25, 1.5, 2]
      },
      fullscreen: {
        enabled: true,
        fallback: true,
        iosNative: true,
        container: undefined
      }
    })
  } else if (props.type === 'audio' && audioRef.value) {
    player = new Plyr(audioRef.value, {
      controls: [
        'play',
        'progress',
        'current-time',
        'duration',
        'mute',
        'volume',
        'speed'
      ],
      speed: {
        selected: 1,
        options: [0.5, 0.75, 1, 1.25, 1.5, 2]
      },
      hideControls: false
    })
  }

  if (player) {
    player.on('ready', () => {
      emit('loaded')
    })

    player.on('error', (event) => {
      emit('error', new Error('播放器错误'))
    })
  }
}

// 图片控制函数
const zoomIn = () => {
  imageScale.value = Math.min(imageScale.value * 1.2, 3)
  updateImageTransform()
}

const zoomOut = () => {
  imageScale.value = Math.max(imageScale.value / 1.2, 0.5)
  updateImageTransform()
}

const resetZoom = () => {
  imageScale.value = 1
  updateImageTransform()
}

const updateImageTransform = () => {
  imageTransform.value = `scale(${imageScale.value})`
}

const onImageLoad = () => {
  emit('loaded')
}

const onImageError = () => {
  emit('error', new Error('图片加载失败'))
}

const downloadImage = () => {
  const link = document.createElement('a')
  link.href = props.src
  link.download = props.title || 'image'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// 监听src变化，重新初始化播放器
watch(() => props.src, () => {
  if (player) {
    player.destroy()
    player = undefined
  }
  if (props.type !== 'image') {
    setTimeout(initPlayer, 100)
  }
})

onMounted(() => {
  if (props.type !== 'image') {
    initPlayer()
  }
})

onUnmounted(() => {
  if (player) {
    player.destroy()
  }
})
</script>

<style scoped>
.media-player {
  width: 100%;
  background: #000;
  border-radius: 0.5rem;
  overflow: hidden;
}

.video-container,
.audio-container {
  width: 100%;
}

.plyr-video {
  width: 100%;
  height: auto;
  max-height: 50vh;
}
.plyr:fullscreen video{
   max-height: 100vh;
}

:deep(.plyr:fullscreen),
:deep(.plyr:-webkit-full-screen),
:deep(.plyr:-moz-full-screen),
:deep(.plyr:-ms-fullscreen) {
  max-height: none !important;
  height: 100vh !important;
}

.audio-container {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 2rem;
}

.audio-visual {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  margin-bottom: 1.5rem;
  color: white;
}

.audio-cover {
  width: 4rem;
  height: 4rem;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.audio-icon {
  color: rgba(255, 255, 255, 0.8);
}

.audio-info {
  flex: 1;
}

.audio-title {
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 0.5rem 0;
  color: white;
}

.audio-artist {
  font-size: 0.875rem;
  margin: 0;
  color: rgba(255, 255, 255, 0.8);
}

.plyr-audio {
  width: 100%;
}

.image-container {
  position: relative;
  background: #000;
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-viewer {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.preview-image {
  max-width: 100%;
  max-height: 70vh;
  object-fit: contain;
  transition: transform 0.3s ease;
  cursor: grab;
}

.preview-image:active {
  cursor: grabbing;
}

.image-controls {
  position: absolute;
  bottom: 1rem;
  right: 1rem;
  display: flex;
  gap: 0.5rem;
  background: rgba(0, 0, 0, 0.7);
  border-radius: 0.5rem;
  padding: 0.5rem;
}

.control-btn {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  border-radius: 0.25rem;
  padding: 0.5rem;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease;
}

.control-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* Plyr自定义样式 */
:deep(.plyr) {
  border-radius: 0;
}

:deep(.plyr--video) {
  background: #000;
}

:deep(.plyr--audio) {
  background: transparent;
}

:deep(.plyr__control--overlaid) {
  background: rgba(59, 130, 246, 0.9);
}

:deep(.plyr__control--overlaid:hover) {
  background: rgba(37, 99, 235, 0.9);
}

:deep(.plyr__progress__buffer) {
  color: rgba(59, 130, 246, 0.3);
}

:deep(.plyr__volume--display) {
  color: #3b82f6;
}

/* 移动端适配样式 */
@media (max-width: 768px) {
  :deep(.plyr__controls) {
    padding: 8px;
  }
  
  :deep(.plyr__control) {
    min-width: 44px;
    min-height: 44px;
    padding: 8px;
  }
  
  :deep(.plyr__control svg) {
    width: 20px;
    height: 20px;
  }
  
  :deep(.plyr__control--overlaid) {
    width: 60px;
    height: 60px;
    padding: 0;
  }
  
  :deep(.plyr__control--overlaid svg) {
    width: 24px;
    height: 24px;
  }
  
  :deep(.plyr__volume) {
    display: none;
  }
  
  :deep(.plyr__menu) {
    right: 8px;
  }
  
  :deep(.plyr__tooltip) {
    display: none;
  }
}

/* 确保图标正确显示 */
:deep(.plyr__control svg) {
  display: block;
  fill: currentColor;
  pointer-events: none;
}

:deep(.plyr__control:not(:disabled):hover svg) {
  fill: currentColor;
}

/* 强制显示控制栏 */
:deep(.plyr--video .plyr__controls) {
  opacity: 1;
  visibility: visible;
}

:deep(.plyr--playing .plyr__controls) {
  transform: translateY(0);
}
</style>