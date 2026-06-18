<template>
  <div class="keyboard-input">
    <div class="keyboard-display">
      <div class="display-label">{{ label }}</div>
      <div class="display-value" :class="{ placeholder: !displayValue }">
        {{ displayValue || placeholder }}
        <span class="cursor" v-if="showCursor">|</span>
      </div>
    </div>

    <div class="keyboard-grid" :class="[`keyboard-grid--${mode}`]">
      <template v-if="mode === 'number'">
        <button
          v-for="key in numberKeys"
          :key="key.value"
          class="key-btn"
          :class="{ 'is-key-wide': key.wide, 'is-danger': key.type === 'danger' }"
          @click="handleKey(key)"
          @touchstart.prevent="handleTouchStart"
          @touchend.prevent="handleKey(key); handleTouchEnd"
        >
          <template v-if="key.icon">{{ key.icon }}</template>
          <template v-else>{{ key.label }}</template>
        </button>
      </template>

      <template v-else-if="mode === 'tel'">
        <button
          v-for="key in telKeys"
          :key="key.value"
          class="key-btn"
          :class="{ 'is-key-wide': key.wide, 'is-danger': key.type === 'danger' }"
          @click="handleKey(key)"
          @touchstart.prevent="handleTouchStart"
          @touchend.prevent="handleKey(key); handleTouchEnd"
        >
          <div class="key-tel-content">
            <template v-if="key.icon">{{ key.icon }}</template>
            <template v-else>
              <div class="key-tel-main">{{ key.label }}</div>
              <div class="key-tel-sub" v-if="key.sub">{{ key.sub }}</div>
            </template>
          </div>
        </button>
      </template>

      <template v-else-if="mode === 'idcard'">
        <button
          v-for="key in idCardKeys"
          :key="key.value"
          class="key-btn"
          :class="{ 'is-key-wide': key.wide, 'is-danger': key.type === 'danger' }"
          @click="handleKey(key)"
          @touchstart.prevent="handleTouchStart"
          @touchend.prevent="handleKey(key); handleTouchEnd"
        >
          <template v-if="key.icon">{{ key.icon }}</template>
          <template v-else>{{ key.label }}</template>
        </button>
      </template>
    </div>

    <div v-if="showError" class="keyboard-error">
      <el-icon><WarningFilled /></el-icon>
      <span>{{ errorMessage }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { WarningFilled } from '@element-plus/icons-vue'

interface KeyboardKey {
  label: string
  value: string
  sub?: string
  icon?: string
  action?: 'input' | 'delete' | 'clear' | 'confirm'
  type?: 'primary' | 'danger'
  wide?: boolean
}

interface Props {
  modelValue: string
  mode?: 'number' | 'tel' | 'idcard'
  label?: string
  placeholder?: string
  maxLength?: number
  confirmText?: string
}

const props = withDefaults(defineProps<Props>(), {
  mode: 'number',
  label: '',
  placeholder: '请点击下方键盘输入',
  maxLength: 30
})

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
  (e: 'confirm', val: string): void
  (e: 'change', val: string): void
}>()

const displayValue = ref(props.modelValue)
const showCursor = ref(false)
const showError = ref(false)
const errorMessage = ref('')

watch(() => props.modelValue, (val) => {
  displayValue.value = val
})

const numberKeys = computed<KeyboardKey[]>(() => [
  { label: '1', value: '1' },
  { label: '2', value: '2' },
  { label: '3', value: '3' },
  { label: '4', value: '4' },
  { label: '5', value: '5' },
  { label: '6', value: '6' },
  { label: '7', value: '7' },
  { label: '8', value: '8' },
  { label: '9', value: '9' },
  { label: '清空', value: '', action: 'clear', type: 'danger', wide: false },
  { label: '0', value: '0' },
  { label: '删除', value: '', icon: '⌫', action: 'delete', wide: false },
  { label: props.confirmText || '确认', value: '', action: 'confirm', type: 'primary', wide: true }
])

const telKeys = computed<KeyboardKey[]>(() => [
  { label: '1', value: '1' },
  { label: '2', value: '2', sub: 'ABC' },
  { label: '3', value: '3', sub: 'DEF' },
  { label: '4', value: '4', sub: 'GHI' },
  { label: '5', value: '5', sub: 'JKL' },
  { label: '6', value: '6', sub: 'MNO' },
  { label: '7', value: '7', sub: 'PQRS' },
  { label: '8', value: '8', sub: 'TUV' },
  { label: '9', value: '9', sub: 'WXYZ' },
  { label: '*', value: '*' },
  { label: '0', value: '0', sub: '+' },
  { label: '#', value: '#' },
  { label: '清空', value: '', action: 'clear', type: 'danger' },
  { label: '删除', value: '', icon: '⌫', action: 'delete' },
  { label: props.confirmText || '确认', value: '', action: 'confirm', type: 'primary' }
])

const idCardKeys = computed<KeyboardKey[]>(() => [
  { label: '1', value: '1' },
  { label: '2', value: '2' },
  { label: '3', value: '3' },
  { label: '4', value: '4' },
  { label: '5', value: '5' },
  { label: '6', value: '6' },
  { label: '7', value: '7' },
  { label: '8', value: '8' },
  { label: '9', value: '9' },
  { label: 'X', value: 'X' },
  { label: '0', value: '0' },
  { label: '删除', value: '', icon: '⌫', action: 'delete' },
  { label: '清空', value: '', action: 'clear', type: 'danger', wide: false },
  { label: props.confirmText || '确认', value: '', action: 'confirm', type: 'primary', wide: true }
])

function handleKey(key: KeyboardKey) {
  showError.value = false

  if (key.action === 'delete') {
    displayValue.value = displayValue.value.slice(0, -1)
  } else if (key.action === 'clear') {
    displayValue.value = ''
  } else if (key.action === 'confirm') {
    emit('confirm', displayValue.value)
    return
  } else {
    if (displayValue.value.length < props.maxLength) {
      displayValue.value += key.value
    }
  }

  emit('update:modelValue', displayValue.value)
  emit('change', displayValue.value)
}

function handleTouchStart(e: TouchEvent) {
  const target = e.currentTarget as HTMLElement
  target.classList.add('is-active')
  showCursor.value = true
}

function handleTouchEnd() {
  document.querySelectorAll('.key-btn.is-active').forEach(el => {
    el.classList.remove('is-active')
  })
}

function setError(msg: string) {
  errorMessage.value = msg
  showError.value = true
}

defineExpose({ setError })
</script>

<style lang="scss" scoped>
.keyboard-input {
  background: $bg-card;
  border-radius: $border-radius-xl;
  padding: 24px;
  box-shadow: $shadow-card;

  .keyboard-display {
    background: #f8fafc;
    border: 3px solid rgba(29, 108, 255, 0.2);
    border-radius: $border-radius-lg;
    padding: 24px 28px;
    margin-bottom: 20px;

    .display-label {
      font-size: 22px;
      color: $text-color-secondary;
      margin-bottom: 8px;
      font-weight: 500;
    }

    .display-value {
      font-size: 40px;
      font-weight: 700;
      color: $text-color-primary;
      font-family: 'Courier New', monospace;
      letter-spacing: 4px;
      min-height: 52px;
      word-break: break-all;

      &.placeholder {
        color: $text-color-light;
        font-weight: 500;
        letter-spacing: 2px;
      }

      .cursor {
        display: inline-block;
        margin-left: 4px;
        animation: blink 1s infinite;
        color: $primary-color;
      }

      @keyframes blink {
        0%, 50% { opacity: 1; }
        51%, 100% { opacity: 0; }
      }
    }
  }

  .keyboard-grid {
    display: grid;
    gap: 12px;

    &--number {
      grid-template-columns: repeat(3, 1fr);

      .is-key-wide {
        grid-column: span 3;
      }
    }

    &--tel {
      grid-template-columns: repeat(3, 1fr);
    }

    &--idcard {
      grid-template-columns: repeat(3, 1fr);

      .is-key-wide {
        grid-column: span 3;
      }
    }

    .key-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 80px;
      padding: 16px;
      font-size: 36px;
      font-weight: 700;
      color: $text-color-primary;
      background: white;
      border: 3px solid #e5e7eb;
      border-radius: $border-radius-md;
      cursor: pointer;
      transition: all 0.15s ease;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
      touch-action: manipulation;
      -webkit-tap-highlight-color: transparent;

      &:hover {
        border-color: $primary-color;
        color: $primary-color;
      }

      &:active,
      &.is-active {
        transform: scale(0.95);
        background: $primary-color;
        color: white;
        border-color: $primary-color;
      }

      &.is-danger {
        background: rgba(239, 68, 68, 0.08);
        border-color: rgba(239, 68, 68, 0.3);
        color: $danger-color;
        font-size: 26px;

        &:active,
        &.is-active {
          background: $danger-color;
          color: white;
        }
      }

      .key-tel-content {
        display: flex;
        flex-direction: column;
        align-items: center;

        .key-tel-main {
          font-size: 36px;
          line-height: 1;
        }

        .key-tel-sub {
          font-size: 14px;
          color: $text-color-light;
          letter-spacing: 2px;
          margin-top: 4px;
          font-weight: 500;
        }
      }
    }

    .key-btn:last-child {
      background: $primary-color;
      border-color: $primary-color;
      color: white;
      font-size: 28px;

      &:hover {
        background: $primary-color-light;
        border-color: $primary-color-light;
      }

      &:active,
      &.is-active {
        background: #1a5ce0;
      }
    }
  }

  .keyboard-error {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    margin-top: 16px;
    padding: 16px;
    background: rgba(239, 68, 68, 0.1);
    border: 2px solid rgba(239, 68, 68, 0.3);
    border-radius: $border-radius-md;
    color: $danger-color;
    font-size: 22px;
    font-weight: 500;
  }
}
</style>
