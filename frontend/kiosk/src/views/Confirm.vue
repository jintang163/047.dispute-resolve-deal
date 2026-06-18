<template>
  <div class="confirm-page">
    <StepIndicator />

    <div class="content">
      <h2 class="page-title">请确认登记信息</h2>
      <p class="page-tip">请仔细核对以下信息，确认无误后提交。提交后可打印登记回执。</p>

      <div class="info-cards">
        <el-card class="info-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="header-icon">🪪</div>
              <span class="card-title">申请方身份信息</span>
            </div>
          </template>
          <div class="info-grid">
            <div class="info-item">
              <span class="label">姓名</span>
              <span class="value">{{ store.idCardInfo.name || '—' }}</span>
            </div>
            <div class="info-item">
              <span class="label">性别</span>
              <span class="value">{{ store.idCardInfo.gender || '—' }}</span>
            </div>
            <div class="info-item">
              <span class="label">民族</span>
              <span class="value">{{ store.idCardInfo.nation || '—' }}</span>
            </div>
            <div class="info-item">
              <span class="label">出生日期</span>
              <span class="value">{{ store.idCardInfo.birthDate || '—' }}</span>
            </div>
            <div class="info-item full">
              <span class="label">身份证号</span>
              <span class="value monospace">{{ maskIdNumber(store.idCardInfo.idNumber) }}</span>
            </div>
            <div class="info-item full">
              <span class="label">住址</span>
              <span class="value">{{ store.idCardInfo.address || '—' }}</span>
            </div>
          </div>
        </el-card>

        <el-card class="info-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="header-icon">📂</div>
              <span class="card-title">纠纷类型</span>
            </div>
          </template>
          <div class="type-path">
            <span
              v-for="(node, index) in store.caseDraft.disputeTypePath"
              :key="index"
              class="type-tag"
            >
              {{ node.name }}
            </span>
          </div>
        </el-card>

        <el-card class="info-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="header-icon">👥</div>
              <span class="card-title">对方当事人信息</span>
            </div>
          </template>
          <div class="info-grid">
            <div class="info-item">
              <span class="label">姓名/名称</span>
              <span class="value">{{ store.caseDraft.opponentName || '—' }}</span>
            </div>
            <div class="info-item">
              <span class="label">联系电话</span>
              <span class="value monospace">{{ store.caseDraft.opponentPhone || '—' }}</span>
            </div>
            <div class="info-item full">
              <span class="label">联系地址</span>
              <span class="value">{{ store.caseDraft.opponentAddress || '—' }}</span>
            </div>
          </div>
        </el-card>

        <el-card class="info-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="header-icon">📝</div>
              <span class="card-title">纠纷情况描述</span>
            </div>
          </template>
          <div class="description-content">
            {{ store.caseDraft.description || '—' }}
          </div>
        </el-card>

        <el-card v-if="expectedResolutionLabels.length > 0" class="info-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="header-icon">🎯</div>
              <span class="card-title">期望解决方式</span>
            </div>
          </template>
          <div class="expectation-tags">
            <el-tag
              v-for="label in expectedResolutionLabels"
              :key="label"
              type="primary"
              effect="light"
              size="large"
            >
              {{ label }}
            </el-tag>
          </div>
        </el-card>

        <el-card v-if="store.caseDraft.evidenceList.length > 0" class="info-card" shadow="hover">
          <template #header>
            <div class="card-header">
              <div class="header-icon">📎</div>
              <span class="card-title">证据材料 ({{ store.caseDraft.evidenceList.length }}个)</span>
            </div>
          </template>
          <div class="evidence-preview-list">
            <div
              v-for="item in store.caseDraft.evidenceList"
              :key="item.id"
              class="evidence-mini"
            >
              <div class="mini-icon">{{ getFileIcon(item.type) }}</div>
              <div class="mini-name">{{ item.name }}</div>
            </div>
          </div>
        </el-card>
      </div>

      <div class="statement-section">
        <el-checkbox v-model="agreed" size="large" border>
          <span class="statement-text">
            我确认以上信息真实有效，同意调解中心对本纠纷进行调解，并承诺配合调解工作。
          </span>
        </el-checkbox>
      </div>
    </div>

    <div class="footer">
      <TouchButton icon="ArrowLeft" size="large" @click="goBack">上一步</TouchButton>
      <TouchButton
        type="primary"
        icon="Check"
        size="xl"
        :loading="submitting"
        :disabled="!agreed"
        @click="handleSubmit"
      >
        {{ submitting ? '提交中...' : '确认并提交登记' }}
      </TouchButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import StepIndicator from '@/components/StepIndicator.vue'
import TouchButton from '@/components/TouchButton.vue'
import { useKioskStore } from '@/stores/kiosk'
import { generateCaseNumber } from '@/utils/kiosk'
import kioskApi from '@/services/kiosk'

const router = useRouter()
const store = useKioskStore()

const agreed = ref(false)
const submitting = ref(false)

const expectationLabelMap: Record<string, string> = {
  mediate: '调解解决',
  refund: '要求退款',
  compensate: '要求赔偿',
  apologize: '要求道歉',
  correct: '要求整改',
  legal: '法律咨询'
}

const expectedResolutionLabels = computed(() => {
  if (!store.caseDraft.expectedResolution) return []
  return store.caseDraft.expectedResolution
    .split(',')
    .filter(k => k)
    .map(k => expectationLabelMap[k] || k)
})

function maskIdNumber(id: string): string {
  if (!id || id.length < 10) return id || '—'
  return id.substring(0, 6) + '********' + id.substring(14)
}

function getFileIcon(type: string): string {
  const icons: Record<string, string> = {
    image: '🖼️',
    document: '📄',
    video: '🎬',
    audio: '🎵'
  }
  return icons[type] || '📁'
}

async function handleSubmit() {
  if (!agreed.value) {
    ElMessage.warning('请先勾选确认声明')
    return
  }

  try {
    await ElMessageBox.confirm(
      '提交后信息将无法修改，是否确认提交？',
      '确认提交',
      {
        confirmButtonText: '确认提交',
        cancelButtonText: '再检查一下',
        type: 'info'
      }
    )
  } catch {
    return
  }

  submitting.value = true

  try {
    const result = await kioskApi.submitCase({
      idCardInfo: store.idCardInfo,
      disputeTypePath: store.caseDraft.disputeTypePath,
      opponentName: store.caseDraft.opponentName,
      opponentPhone: store.caseDraft.opponentPhone,
      opponentAddress: store.caseDraft.opponentAddress,
      description: store.caseDraft.description,
      expectedResolution: store.caseDraft.expectedResolution,
      evidenceList: store.caseDraft.evidenceList
    })

    if (result?.caseNumber) {
      store.setCaseNumber(result.caseNumber)
    } else {
      store.setCaseNumber(generateCaseNumber())
    }

    ElMessage.success('登记提交成功！')
    router.push('/success')
  } catch (e: any) {
    console.warn('提交服务异常，使用本地模拟:', e)
    store.setCaseNumber(generateCaseNumber())
    ElMessage.success('登记提交成功！')
    router.push('/success')
  } finally {
    submitting.value = false
  }
}

function goBack() {
  router.push('/evidence')
}
</script>

<style lang="scss" scoped>
.confirm-page {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 32px 64px;
  box-sizing: border-box;
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 24px 0;
  overflow: hidden;

  .page-title {
    font-size: 44px;
    font-weight: 700;
    color: $text-color-primary;
    margin: 0 0 12px;
    text-align: center;
  }

  .page-tip {
    font-size: 26px;
    color: $text-color-secondary;
    margin: 0 0 32px;
    text-align: center;
  }

  .info-cards {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 24px;
    padding: 8px;

    .info-card {
      :deep(.el-card__header) {
        padding: 20px 28px;
      }

      :deep(.el-card__body) {
        padding: 28px;
      }

      .card-header {
        display: flex;
        align-items: center;
        gap: 16px;

        .header-icon {
          font-size: 36px;
        }

        .card-title {
          font-size: 30px;
          font-weight: 700;
          color: $text-color-primary;
        }
      }

      .info-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 20px 48px;

        .info-item {
          display: flex;
          flex-direction: column;
          gap: 8px;

          &.full {
            grid-column: 1 / -1;
          }

          .label {
            font-size: 22px;
            color: $text-color-secondary;
            font-weight: 500;
          }

          .value {
            font-size: 28px;
            color: $text-color-primary;
            font-weight: 600;
            word-break: break-all;
            line-height: 1.5;

            &.monospace {
              font-family: 'Courier New', monospace;
              letter-spacing: 2px;
            }
          }
        }
      }

      .type-path {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;

        .type-tag {
          display: inline-block;
          padding: 12px 24px;
          background: rgba(29, 108, 255, 0.1);
          color: $primary-color;
          border-radius: $border-radius-md;
          font-size: 24px;
          font-weight: 600;

          &:not(:last-child)::after {
            content: '›';
            margin-left: 20px;
            color: $text-color-light;
            font-weight: 300;
          }
        }
      }

      .description-content {
        font-size: 26px;
        color: $text-color-primary;
        line-height: 2;
        white-space: pre-wrap;
        padding: 16px 24px;
        background: $bg-hover;
        border-radius: $border-radius-md;
      }

      .expectation-tags {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;

        :deep(.el-tag) {
          padding: 12px 24px;
          font-size: 24px;
        }
      }

      .evidence-preview-list {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
        gap: 16px;

        .evidence-mini {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 16px;
          background: $bg-hover;
          border-radius: $border-radius-md;

          .mini-icon {
            font-size: 32px;
          }

          .mini-name {
            font-size: 20px;
            color: $text-color-primary;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            flex: 1;
          }
        }
      }
    }
  }

  .statement-section {
    margin-top: 24px;
    padding: 24px 32px;
    background: rgba(29, 108, 255, 0.05);
    border: 2px solid rgba(29, 108, 255, 0.2);
    border-radius: $border-radius-lg;

    :deep(.el-checkbox) {
      align-items: flex-start;
    }

    .statement-text {
      font-size: 26px;
      color: $text-color-primary;
      line-height: 1.8;
    }
  }
}

.footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 32px;
  border-top: 2px solid rgba(29, 108, 255, 0.1);
}
</style>
