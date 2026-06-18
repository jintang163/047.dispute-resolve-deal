<template>
  <button
    :type="nativeType"
    class="touch-button"
    :class="[
      `touch-button--${type}`,
      `touch-button--${size}`,
      {
        'is-disabled': disabled,
        'is-loading': loading,
        'is-plain': plain,
        'is-round': round,
        'is-circle': circle,
        'is-block': block
      }
    ]"
    :disabled="disabled || loading"
    @click="handleClick"
    @touchstart.prevent="handleTouchStart"
    @touchend.prevent="handleTouchEnd"
  >
    <el-icon v-if="loading" class="is-loading"><Loading /></el-icon>
    <el-icon v-else-if="icon" :size="iconSize"><component :is="icon" /></el-icon>
    <span v-if="$slots.default" class="touch-button__text">
      <slot />
    </span>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import type { Component } from 'vue'

interface Props {
  type?: 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'default'
  size?: 'medium' | 'large' | 'xl'
  icon?: string | Component
  nativeType?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  loading?: boolean
  plain?: boolean
  round?: boolean
  circle?: boolean
  block?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'default',
  size: 'large',
  nativeType: 'button',
  disabled: false,
  loading: false,
  plain: false,
  round: false,
  circle: false,
  block: false
})

const emit = defineEmits<{
  (e: 'click', evt: MouseEvent): void
}>()

const sizeMap = {
  medium: 28,
  large: 32,
  xl: 40
}

const iconSize = computed(() => sizeMap[props.size] || 32)

function handleClick(e: MouseEvent) {
  if (props.disabled || props.loading) return
  emit('click', e)
}

function handleTouchStart(e: TouchEvent) {
  const target = e.currentTarget as HTMLElement
  target.classList.add('is-touched')
}

function handleTouchEnd(e: TouchEvent) {
  const target = e.currentTarget as HTMLElement
  target.classList.remove('is-touched')
  if (props.disabled || props.loading) return
  emit('click', {} as MouseEvent)
}
</script>

<style lang="scss" scoped>
.touch-button {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  font-weight: 600;
  border: 3px solid transparent;
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
  touch-action: manipulation;
  transition: all 0.2s ease;
  white-space: nowrap;
  background: $bg-card;
  color: $text-color-primary;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);

  &:not(.is-disabled):hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  }

  &:not(.is-disabled):active,
  &.is-touched {
    transform: translateY(1px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  &.is-disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  &.is-loading {
    pointer-events: none;
    .is-loading {
      animation: rotate 1s linear infinite;
    }
  }

  @keyframes rotate {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  &__text {
    display: inline-block;
  }

  &--medium {
    font-size: 24px;
    padding: 16px 32px;
    min-height: 64px;
    border-radius: 12px;
  }

  &--large {
    font-size: 28px;
    padding: 20px 40px;
    min-height: 80px;
    border-radius: 16px;
  }

  &--xl {
    font-size: 32px;
    padding: 28px 56px;
    min-height: 96px;
    border-radius: 20px;
  }

  &--primary {
    background: $primary-color;
    color: white;
    border-color: $primary-color;

    &:not(.is-disabled):hover {
      background: $primary-color-light;
      border-color: $primary-color-light;
    }

    &.is-plain {
      background: rgba(29, 108, 255, 0.1);
      color: $primary-color;
      border-color: rgba(29, 108, 255, 0.3);
    }
  }

  &--success {
    background: $success-color;
    color: white;
    border-color: $success-color;

    &.is-plain {
      background: rgba(34, 197, 94, 0.1);
      color: $success-color;
      border-color: rgba(34, 197, 94, 0.3);
    }
  }

  &--warning {
    background: $warning-color;
    color: white;
    border-color: $warning-color;

    &.is-plain {
      background: rgba(245, 158, 11, 0.1);
      color: $warning-color;
      border-color: rgba(245, 158, 11, 0.3);
    }
  }

  &--danger {
    background: $danger-color;
    color: white;
    border-color: $danger-color;

    &.is-plain {
      background: rgba(239, 68, 68, 0.1);
      color: $danger-color;
      border-color: rgba(239, 68, 68, 0.3);
    }
  }

  &--info {
    background: $info-color;
    color: white;
    border-color: $info-color;

    &.is-plain {
      background: rgba(99, 102, 241, 0.1);
      color: $info-color;
      border-color: rgba(99, 102, 241, 0.3);
    }
  }

  &--default {
    background: $bg-card;
    color: $text-color-primary;
    border-color: #e5e7eb;

    &:not(.is-disabled):hover {
      border-color: $primary-color;
      color: $primary-color;
    }
  }

  &.is-round {
    border-radius: 999px;
  }

  &.is-circle {
    padding: 0;
    border-radius: 50%;

    &.touch-button--medium {
      width: 64px;
      height: 64px;
      min-height: 64px;
    }

    &.touch-button--large {
      width: 80px;
      height: 80px;
      min-height: 80px;
    }

    &.touch-button--xl {
      width: 96px;
      height: 96px;
      min-height: 96px;
    }
  }

  &.is-block {
    display: flex;
    width: 100%;
  }
}
</style>
