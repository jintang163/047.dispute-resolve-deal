<template>
  <div class="app-root">
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
</template>

<script setup lang="ts">
</script>

<style lang="scss">
html, body, #app {
  margin: 0;
  padding: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  font-size: 24px;
  color: #1a1a1a;
  background: linear-gradient(135deg, #f0f7ff 0%, #e6f0ff 100%);
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  -webkit-user-select: none;
}

.app-root {
  width: 100%;
  height: 100%;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

:deep(.el-button) {
  font-size: 28px !important;
  padding: 24px 48px !important;
  min-height: 80px !important;
  border-radius: 16px !important;
}

:deep(.el-button--large) {
  font-size: 32px !important;
  padding: 32px 64px !important;
  min-height: 100px !important;
}

:deep(.el-input__wrapper) {
  font-size: 28px !important;
  padding: 20px 24px !important;
  min-height: 80px !important;
}

:deep(.el-input__inner) {
  font-size: 28px !important;
  height: 60px !important;
}

:deep(.el-textarea__inner) {
  font-size: 28px !important;
  padding: 24px !important;
  line-height: 1.8 !important;
}

:deep(.el-form-item__label) {
  font-size: 28px !important;
  line-height: 80px !important;
}

:deep(.el-card) {
  border-radius: 24px !important;
}

:deep(.el-dialog) {
  width: 90% !important;
  max-width: 900px !important;
  border-radius: 24px !important;
  --el-dialog-title-font-size: 36px;
}

:deep(.el-step__title) {
  font-size: 24px !important;
}

:deep(.el-upload) {
  font-size: 28px !important;
}

:deep(.el-upload-dragger) {
  padding: 48px !important;
}

:deep(.el-radio__label) {
  font-size: 28px !important;
  padding-left: 16px !important;
}

:deep(.el-radio__inner) {
  width: 28px !important;
  height: 28px !important;
}

:deep(.el-checkbox__label) {
  font-size: 28px !important;
  padding-left: 16px !important;
}

:deep(.el-checkbox__inner) {
  width: 28px !important;
  height: 28px !important;
}

:deep(.el-select) {
  --el-select-input-font-size: 28px;
}

:deep(.el-option) {
  font-size: 28px !important;
  padding: 20px 24px !important;
  height: auto !important;
  line-height: 1.5 !important;
}

:deep(.el-message-box) {
  width: 80% !important;
  max-width: 600px !important;
  border-radius: 20px !important;
}

:deep(.el-message-box__title) {
  font-size: 32px !important;
}

:deep(.el-message-box__message) {
  font-size: 28px !important;
}
</style>
