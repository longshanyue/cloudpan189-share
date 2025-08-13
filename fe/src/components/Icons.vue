<template>
  <svg v-if="name === 'dashboard'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z"/>
  </svg>

  <svg v-else-if="name === 'users'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M16 7c0-2.21-1.79-4-4-4S8 4.79 8 7s1.79 4 4 4 4-1.79 4-4zm-4 6c-2.67 0-8 1.34-8 4v3h16v-3c0-2.66-5.33-4-8-4z"/>
  </svg>

  <svg v-else-if="name === 'group'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M9 12c1.66 0 3-1.34 3-3s-1.34-3-3-3-3 1.34-3 3 1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-1.5c0-2.33-4.67-3.5-7-3.5z"/>
    <path d="M16.5 12c1.38 0 2.5-1.12 2.5-2.5S17.88 7 16.5 7 14 8.12 14 9.5s1.12 2.5 2.5 2.5zm0 2c-1.83 0-5.5.92-5.5 2.75V18h11v-1.25c0-1.83-3.67-2.75-5.5-2.75z"/>
    <circle cx="9" cy="9" r="1.5" fill="white"/>
    <circle cx="16.5" cy="9.5" r="1.2" fill="white"/>
  </svg>

  <svg v-else-if="name === 'storage'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M2 20h20v-4H2v4zm2-3h2v2H4v-2zM2 4v4h20V4H2zm4 3H4V5h2v2zm-4 7h20v-4H2v4zm2-3h2v2H4v-2z"/>
  </svg>

  <svg v-else-if="name === 'tokens'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M12.65 10C11.83 7.67 9.61 6 7 6c-3.31 0-6 2.69-6 6s2.69 6 6 6c2.61 0 4.83-1.67 5.65-4H17v4h4v-4h2v-4H12.65zM7 14c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2z"/>
  </svg>

  <svg v-else-if="name === 'settings'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M19.14,12.94c0.04-0.3,0.06-0.61,0.06-0.94c0-0.32-0.02-0.64-0.07-0.94l2.03-1.58c0.18-0.14,0.23-0.41,0.12-0.61 l-1.92-3.32c-0.12-0.22-0.37-0.29-0.59-0.22l-2.39,0.96c-0.5-0.38-1.03-0.7-1.62-0.94L14.4,2.81c-0.04-0.24-0.24-0.41-0.48-0.41 h-3.84c-0.24,0-0.43,0.17-0.47,0.41L9.25,5.35C8.66,5.59,8.12,5.92,7.63,6.29L5.24,5.33c-0.22-0.08-0.47,0-0.59,0.22L2.74,8.87 C2.62,9.08,2.66,9.34,2.86,9.48l2.03,1.58C4.84,11.36,4.8,11.69,4.8,12s0.02,0.64,0.07,0.94l-2.03,1.58 c-0.18,0.14-0.23,0.41-0.12,0.61l1.92,3.32c0.12,0.22,0.37,0.29,0.59,0.22l2.39-0.96c0.5,0.38,1.03,0.7,1.62,0.94l0.36,2.54 c0.05,0.24,0.24,0.41,0.48,0.41h3.84c0.24,0,0.44-0.17,0.47-0.41l0.36-2.54c0.59-0.24,1.13-0.56,1.62-0.94l2.39,0.96 c0.22,0.08,0.47,0,0.59-0.22l1.92-3.32c0.12-0.22,0.07-0.47-0.12-0.61L19.14,12.94z M12,15.6c-1.98,0-3.6-1.62-3.6-3.6 s1.62-3.6,3.6-3.6s3.6,1.62,3.6,3.6S13.98,15.6,12,15.6z"/>
  </svg>

  <svg v-else-if="name === 'cloud'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96z"/>
  </svg>

  <svg v-else-if="name === 'logout'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.58L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z"/>
  </svg>

  <svg v-else-if="name === 'add'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
  </svg>

  <svg v-else-if="name === 'search'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
  </svg>

  <svg v-else-if="name === 'edit'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
  </svg>

  <svg v-else-if="name === 'delete'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
  </svg>

  <svg v-else-if="name === 'key'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M12.65 10C11.83 7.67 9.61 6 7 6c-3.31 0-6 2.69-6 6s2.69 6 6 6c2.61 0 4.83-1.67 5.65-4H17v4h4v-4h2v-4H12.65zM7 14c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2z"/>
  </svg>

  <svg v-else-if="name === 'password'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/>
  </svg>

  <svg v-else-if="name === 'user'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
  </svg>

  <svg v-else-if="name === 'info'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/>
  </svg>

  <svg v-else-if="name === 'warning'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
  </svg>

  <!-- 文件浏览器相关图标 -->
  <svg v-else-if="name === 'folder'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M10 4H4c-1.11 0-2 .89-2 2v12c0 1.11.89 2 2 2h16c1.11 0 2-.89 2-2V8c0-1.11-.89-2-2-2h-8l-2-2z"/>
  </svg>

  <svg v-else-if="name === 'file'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
  </svg>

  <svg v-else-if="name === 'file-text'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M14,2H6A2,2 0 0,0 4,4V20A2,2 0 0,0 6,22H18A2,2 0 0,0 20,20V8L14,2M18,20H6V4H13V9H18V20Z"/>
  </svg>

  <svg v-else-if="name === 'video'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M17,10.5V7A1,1 0 0,0 16,6H4A1,1 0 0,0 3,7V17A1,1 0 0,0 4,18H16A1,1 0 0,0 17,17V13.5L21,17.5V6.5L17,10.5Z"/>
  </svg>

  <svg v-else-if="name === 'music'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M12,3V13.55C11.41,13.21 10.73,13 10,13A3,3 0 0,0 7,16A3,3 0 0,0 10,19A3,3 0 0,0 13,16V7H18V5H12Z"/>
  </svg>

  <svg v-else-if="name === 'image'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M8.5,13.5L11,16.5L14.5,12L19,18H5M21,19V5C21,3.89 20.1,3 19,3H5A2,2 0 0,0 3,5V19A2,2 0 0,0 5,21H19A2,2 0 0,0 21,19Z"/>
  </svg>

  <svg v-else-if="name === 'archive'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M14,17H12V15H14M14,13H12V11H14M20,9H4V7H20M20,5H4V3H20M4,19H20V21H4V19Z"/>
  </svg>

  <svg v-else-if="name === 'download'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M5,20H19V18H5M19,9H15V3H9V9H5L12,16L19,9Z"/>
  </svg>

  <svg v-else-if="name === 'refresh'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M17.65,6.35C16.2,4.9 14.21,4 12,4A8,8 0 0,0 4,12A8,8 0 0,0 12,20C15.73,20 18.84,17.45 19.73,14H17.65C16.83,16.33 14.61,18 12,18A6,6 0 0,1 6,12A6,6 0 0,1 12,6C13.66,6 15.14,6.69 16.22,7.78L13,11H20V4L17.65,6.35Z"/>
  </svg>

  <svg v-else-if="name === 'home'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M10,20V14H14V20H19V12H22L12,3L2,12H5V20H10Z"/>
  </svg>

  <svg v-else-if="name === 'arrow-up'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M7.41,15.41L12,10.83L16.59,15.41L18,14L12,8L6,14L7.41,15.41Z"/>
  </svg>

  <svg v-else-if="name === 'alert'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M13,13H11V7H13M13,17H11V15H13M12,2A10,10 0 0,0 2,12A10,10 0 0,0 12,22A10,10 0 0,0 22,12A10,10 0 0,0 12,2Z"/>
  </svg>

  <svg v-else-if="name === 'qr-code'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M3,11H5V13H3V11M11,5H13V9H11V5M9,11H13V15H9V11M15,11H17V13H15V11M19,11H21V13H19V11M12,15H14V17H12V15M3,5H9V9H3V5M5,7V7H7V7H5M3,15H9V19H3V15M5,17V17H7V17H5M15,5H21V9H15V5M17,7V7H19V7H17M15,15H21V19H15V15M17,17V17H19V17H17Z"/>
  </svg>

  <svg v-else-if="name === 'link'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M3.9,12C3.9,10.29 5.29,8.9 7,8.9H11V7H7A5,5 0 0,0 2,12A5,5 0 0,0 7,17H11V15.1H7C5.29,15.1 3.9,13.71 3.9,12M8,13H16V11H8V13M17,7H13V8.9H17C18.71,8.9 20.1,10.29 20.1,12C20.1,13.71 18.71,15.1 17,15.1H13V17H17A5,5 0 0,0 22,12A5,5 0 0,0 17,7Z"/>
  </svg>

  <svg v-else-if="name === 'arrow-left'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M20,11V13H8L13.5,18.5L12.08,19.92L4.16,12L12.08,4.08L13.5,5.5L8,11H20Z"/>
  </svg>

  <svg v-else-if="name === 'check-circle'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M12,2A10,10 0 0,1 22,12A10,10 0 0,1 12,22A10,10 0 0,1 2,12A10,10 0 0,1 12,2M11,16.5L18,9.5L16.59,8.09L11,13.67L7.91,10.59L6.5,12L11,16.5Z"/>
  </svg>

  <svg v-else-if="name === 'close'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M19,6.41L17.59,5L12,10.59L6.41,5L5,6.41L10.59,12L5,17.59L6.41,19L12,13.41L17.59,19L19,17.59L13.41,12L19,6.41Z"/>
  </svg>

  <svg v-else-if="name === 'chevron-right'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M8.59,16.58L13.17,12L8.59,7.41L10,6L16,12L10,18L8.59,16.58Z"/>
  </svg>

  <svg v-else-if="name === 'zoom-in'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
    <path d="M12 10h-2v2H9v-2H7V9h2V7h1v2h2v1z"/>
  </svg>

  <svg v-else-if="name === 'zoom-out'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
    <path d="M7 9h5v1H7z"/>
  </svg>

  <!-- 新增删除相关图标 -->
  <svg v-else-if="name === 'trash-2'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M3 6v18h18v-18h-18zm5 14c0 .552-.448 1-1 1s-1-.448-1-1v-10c0-.552.448-1 1-1s1 .448 1 1v10zm5 0c0 .552-.448 1-1 1s-1-.448-1-1v-10c0-.552.448-1 1-1s1 .448 1 1v10zm5 0c0 .552-.448 1-1 1s-1-.448-1-1v-10c0-.552.448-1 1-1s1 .448 1 1v10zm4-18v2h-20v-2h5.711c.9 0 1.631-1.099 1.631-2h5.315c0 .901.73 2 1.631 2h5.712z"/>
  </svg>

  <svg v-else-if="name === 'x'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M18 6L6 18M6 6l12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
  </svg>

  <svg v-else-if="name === 'alert-triangle'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
    <line x1="12" y1="9" x2="12" y2="13" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
    <circle cx="12" cy="17" r="1" fill="currentColor"/>
  </svg>

  <svg v-else-if="name === 'loader'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M12,4a8,8,0,0,1,7.89,6.7A1.53,1.53,0,0,0,21.38,12h0a1.5,1.5,0,0,0,1.48-1.75,11,11,0,0,0-21.72,0A1.5,1.5,0,0,0,2.62,12h0a1.53,1.53,0,0,0,1.49-1.3A8,8,0,0,1,12,4Z">
      <animateTransform attributeName="transform" dur="0.75s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/>
    </path>
  </svg>

  <svg v-else-if="name === 'copy'" viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
  </svg>

  <!-- 默认图标 -->
  <svg v-else viewBox="0 0 24 24" fill="currentColor" class="icon">
    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
  </svg>
</template>

<script setup lang="ts">
interface Props {
  name: string
  size?: string | number
}

withDefaults(defineProps<Props>(), {
  size: '1.25rem'
})
</script>

<style scoped>
.icon {
  width: v-bind(size);
  height: v-bind(size);
  display: inline-block;
  vertical-align: middle;
}
</style>
