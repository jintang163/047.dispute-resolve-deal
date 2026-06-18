<template>
  <div class="step-indicator">
    <div class="progress-track">
      <div class="progress-fill" :style="{ width: `${progressWidth}%` }"></div>
    </div>

    <div class="steps-row">
      <div
        v-for="(label, index) in displaySteps"
        :key="index"
        class="step-item"
        :class="{
          'is-active': currentStep === index,
          'is-completed': currentStep > index,
          'is-inactive': currentStep < index
        }"
      >
        <div class="step-circle">
          <span v-if="currentStep > index" class="check-icon">✓</span>
          <span v-else class="step-number">{{ index + 1 }}</span>
        </div>
        <div class="step-label">{{ label }}</div>
      </div>
    </div>

    <div class="step-header">
      <h2 class="current-label">{{ currentStepLabel }}</h2>
      <div class="step-counter">第 {{ Math.min(currentStep + 1, store.totalSteps + 1) }} / {{ store.totalSteps + 1 }} 步</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useKioskStore } from '@/stores/kiosk'

const store = useKioskStore()

const currentStep = computed(() => store.currentStep)
const currentStepLabel = computed(() => store.currentStepLabel)

const displaySteps = computed(() => {
  return store.stepLabels.slice(1, -1)
})

const progressWidth = computed(() => {
  const max = store.totalSteps - 1
  const current = Math.max(0, currentStep.value - 1)
  return Math.min(100, Math.round((current / max) * 100))
})
</script>

<style lang="scss" scoped>
.step-indicator {
  margin-bottom: 24px;

  .progress-track {
    width: 100%;
    height: 8px;
    background: rgba(0, 0, 0, 0.08);
    border-radius: 999px;
    overflow: hidden;
    margin-bottom: 20px;

    .progress-fill {
      height: 100%;
      background: linear-gradient(90deg, $primary-color 0%, $primary-color-light 100%);
      border-radius: 999px;
      transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
    }
  }

  .steps-row {
    display: flex;
    justify-content: space-between;
    position: relative;
    margin-bottom: 20px;

    .step-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      flex: 1;
      position: relative;
      z-index: 2;

      .step-circle {
        width: 56px;
        height: 56px;
        border-radius: 50%;
        background: #e5e7eb;
        border: 4px solid transparent;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-bottom: 12px;
        transition: all 0.3s ease;

        .step-number {
          font-size: 24px;
          font-weight: 700;
          color: $text-color-secondary;
        }

        .check-icon {
          font-size: 28px;
          color: white;
          font-weight: 700;
        }
      }

      .step-label {
        font-size: 20px;
        color: $text-color-secondary;
        text-align: center;
        font-weight: 500;
        transition: all 0.3s ease;
      }

      &.is-active {
        .step-circle {
          background: $primary-color;
          border-color: rgba(29, 108, 255, 0.2);
          box-shadow: 0 0 0 6px rgba(29, 108, 255, 0.15);
          transform: scale(1.1);

          .step-number {
            color: white;
          }
        }

        .step-label {
          color: $primary-color;
          font-weight: 700;
          font-size: 22px;
        }
      }

      &.is-completed {
        .step-circle {
          background: $success-color;
        }

        .step-label {
          color: $success-color;
        }
      }

      &.is-inactive {
        .step-circle {
          opacity: 0.6;
        }
      }
    }
  }

  .step-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 28px;
    background: $bg-card;
    border-radius: $border-radius-lg;
    box-shadow: $shadow-card;

    .current-label {
      font-size: 32px;
      font-weight: 700;
      color: $text-color-primary;
      margin: 0;
    }

    .step-counter {
      font-size: 24px;
      color: $text-color-secondary;
      font-weight: 500;
      padding: 8px 20px;
      background: rgba(29, 108, 255, 0.1);
      color: $primary-color;
      border-radius: $border-radius-md;
    }
  }
}
</style>
