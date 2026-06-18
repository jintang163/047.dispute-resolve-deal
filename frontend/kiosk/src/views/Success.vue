<template>
  <div class="success-page">
    <div class="content">
      <div class="success-animation">
        <div class="success-icon">
          <div class="checkmark-circle">
            <div class="background"></div>
            <div class="checkmark">
              <div class="checkmark-draw"></div>
            </div>
          </div>
        </div>
        <h1 class="success-title">登记提交成功！</h1>
        <p class="success-subtitle">您的纠纷信息已成功录入，工作人员将尽快与您联系</p>
      </div>

      <div class="case-number-card">
        <div class="case-label">您的案件编号</div>
        <div class="case-number">{{ store.caseNumber }}</div>
        <div class="case-tip">请妥善保存此编号，用于查询调解进度</div>
      </div>

      <div class="info-cards">
        <div class="info-row">
          <div class="info-block">
            <div class="info-icon">📅</div>
            <div>
              <div class="info-label">登记时间</div>
              <div class="info-value">{{ store.createdAt }}</div>
            </div>
          </div>
          <div class="info-block">
            <div class="info-icon">⏱️</div>
            <div>
              <div class="info-label">预计办理周期</div>
              <div class="info-value">3 - 7 个工作日</div>
            </div>
          </div>
        </div>

        <div class="info-block full">
          <div class="info-icon">📞</div>
          <div>
            <div class="info-label">如有疑问请联系</div>
            <div class="info-value large">调解服务热线：12348（工作日 8:00-20:00）</div>
          </div>
        </div>
      </div>

      <div class="qr-section">
        <div class="qr-card">
          <div class="qr-header">
            <h3>扫码查询进度 / 下载回执</h3>
          </div>
          <div class="qr-container">
            <img v-if="qrCodeUrl" :src="qrCodeUrl" alt="查询二维码" class="qr-image" />
            <div v-else class="qr-placeholder">
              <div class="spinner"></div>
              <p>生成中...</p>
            </div>
          </div>
          <div class="qr-tip">使用微信扫描二维码</div>
        </div>
      </div>

      <div class="next-steps">
        <h3 class="steps-title">办理流程说明</h3>
        <div class="steps-list">
          <div class="step-item">
            <div class="step-num">1</div>
            <div class="step-content">
              <div class="step-name">材料审核</div>
              <div class="step-desc">工作人员审核登记信息，1-2个工作日</div>
            </div>
          </div>
          <div class="step-item">
            <div class="step-num">2</div>
            <div class="step-content">
              <div class="step-name">分配调解员</div>
              <div class="step-desc">根据纠纷类型分配专业调解员</div>
            </div>
          </div>
          <div class="step-item">
            <div class="step-num">3</div>
            <div class="step-content">
              <div class="step-name">联系沟通</div>
              <div class="step-desc">调解员与双方电话沟通了解情况</div>
            </div>
          </div>
          <div class="step-item">
            <div class="step-num">4</div>
            <div class="step-content">
              <div class="step-name">组织调解</div>
              <div class="step-desc">安排调解时间，开展调解工作</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="footer">
      <TouchButton icon="Printer" size="large" :loading="printing" @click="handlePrint">
        {{ printing ? '打印中...' : '打印登记回执' }}
      </TouchButton>
      <TouchButton icon="Service" size="large" @click="goToAIHelp">
        AI法律咨询
      </TouchButton>
      <TouchButton type="primary" icon="HomeFilled" size="xl" @click="resetAndHome">
        返回首页
      </TouchButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import TouchButton from '@/components/TouchButton.vue'
import { useKioskStore } from '@/stores/kiosk'
import { generateQRCode, printReceipt } from '@/utils/kiosk'

const router = useRouter()
const store = useKioskStore()

const qrCodeUrl = ref('')
const printing = ref(false)

async function generateQR() {
  const queryUrl = `${window.location.origin}/query?case=${store.caseNumber}`
  qrCodeUrl.value = await generateQRCode(queryUrl, { width: 320 })
}

async function handlePrint() {
  printing.value = true
  try {
    const success = await printReceipt({
      caseNumber: store.caseNumber,
      title: '纠纷登记回执',
      items: [
        { label: '案件编号', value: store.caseNumber },
        { label: '登记时间', value: store.createdAt },
        { label: '申请人', value: store.idCardInfo.name },
        { label: '纠纷类型', value: store.caseDraft.disputeTypePath.map(n => n.name).join(' > ') },
        { label: '服务热线', value: '12348' }
      ],
      qrCodeData: `${window.location.origin}/query?case=${store.caseNumber}`,
      footerText: '感谢您的信任，我们将竭诚为您服务！'
    })

    if (success) {
      ElMessage.success('回执打印任务已发送')
    } else {
      ElMessage.warning('打印设备未连接，请截图保存案件编号')
    }
  } catch (e) {
    ElMessage.error('打印失败，请稍后重试')
  } finally {
    printing.value = false
  }
}

function goToAIHelp() {
  router.push('/ai-help')
}

function resetAndHome() {
  store.resetAll()
  router.push('/')
}

onMounted(() => {
  if (!store.caseNumber) {
    router.push('/')
    return
  }
  generateQR()
})
</script>

<style lang="scss" scoped>
.success-page {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 32px 64px;
  box-sizing: border-box;
  background: linear-gradient(180deg, #e6f0ff 0%, #f0f7ff 40%, #ffffff 100%);
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 0;
  overflow-y: auto;

  .success-animation {
    text-align: center;
    margin-bottom: 40px;

    .success-icon {
      margin-bottom: 24px;
    }

    .checkmark-circle {
      position: relative;
      width: 160px;
      height: 160px;
      display: inline-block;

      .background {
        position: absolute;
        inset: 0;
        border-radius: 50%;
        background: $success-color;
        animation: popIn 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275) forwards;
      }

      .checkmark {
        position: absolute;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 2;

        .checkmark-draw {
          width: 64px;
          height: 112px;
          border-right: 12px solid white;
          border-bottom: 12px solid white;
          transform: rotate(45deg) scale(0);
          margin-top: -20px;
          animation: checkDraw 0.5s 0.3s ease forwards;
        }
      }
    }

    @keyframes popIn {
      0% { transform: scale(0); opacity: 0; }
      80% { transform: scale(1.1); }
      100% { transform: scale(1); opacity: 1; }
    }

    @keyframes checkDraw {
      0% { transform: rotate(45deg) scale(0); }
      100% { transform: rotate(45deg) scale(1); }
    }

    .success-title {
      font-size: 56px;
      font-weight: 700;
      color: $success-color;
      margin: 0 0 16px;
    }

    .success-subtitle {
      font-size: 28px;
      color: $text-color-secondary;
      margin: 0;
    }
  }

  .case-number-card {
    background: linear-gradient(135deg, $primary-color 0%, $primary-color-light 100%);
    color: white;
    padding: 48px 80px;
    border-radius: $border-radius-xl;
    text-align: center;
    margin-bottom: 32px;
    box-shadow: 0 12px 40px rgba(29, 108, 255, 0.3);
    min-width: 600px;

    .case-label {
      font-size: 26px;
      opacity: 0.9;
      margin-bottom: 12px;
    }

    .case-number {
      font-size: 64px;
      font-weight: 800;
      letter-spacing: 8px;
      margin-bottom: 12px;
      font-family: 'Courier New', monospace;
      text-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
    }

    .case-tip {
      font-size: 22px;
      opacity: 0.85;
    }
  }

  .info-cards {
    width: 100%;
    max-width: 900px;
    background: $bg-card;
    border-radius: $border-radius-xl;
    padding: 32px;
    margin-bottom: 32px;
    box-shadow: $shadow-card;

    .info-row {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 24px;
      margin-bottom: 24px;
    }

    .info-block {
      display: flex;
      align-items: center;
      gap: 20px;
      padding: 20px;
      background: $bg-hover;
      border-radius: $border-radius-md;

      &.full {
        width: 100%;
      }

      .info-icon {
        font-size: 44px;
        flex-shrink: 0;
      }

      .info-label {
        font-size: 22px;
        color: $text-color-secondary;
        margin-bottom: 4px;
      }

      .info-value {
        font-size: 28px;
        font-weight: 700;
        color: $text-color-primary;

        &.large {
          font-size: 30px;
          color: $primary-color;
        }
      }
    }
  }

  .qr-section {
    margin-bottom: 32px;

    .qr-card {
      background: $bg-card;
      border-radius: $border-radius-xl;
      padding: 32px 48px;
      text-align: center;
      box-shadow: $shadow-card;

      .qr-header h3 {
        font-size: 28px;
        color: $text-color-primary;
        margin: 0 0 24px;
      }

      .qr-container {
        width: 320px;
        height: 320px;
        margin: 0 auto 16px;
        background: white;
        padding: 16px;
        border-radius: $border-radius-md;
        display: flex;
        align-items: center;
        justify-content: center;
        border: 2px solid rgba(0, 0, 0, 0.1);

        .qr-image {
          width: 100%;
          height: 100%;
        }

        .qr-placeholder {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 16px;
          color: $text-color-light;

          .spinner {
            width: 60px;
            height: 60px;
            border: 6px solid rgba(29, 108, 255, 0.2);
            border-top-color: $primary-color;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
          }

          @keyframes spin {
            to { transform: rotate(360deg); }
          }

          p {
            font-size: 22px;
            margin: 0;
          }
        }
      }

      .qr-tip {
        font-size: 22px;
        color: $text-color-secondary;
      }
    }
  }

  .next-steps {
    width: 100%;
    max-width: 900px;
    background: $bg-card;
    border-radius: $border-radius-xl;
    padding: 32px;
    box-shadow: $shadow-card;

    .steps-title {
      font-size: 32px;
      font-weight: 700;
      color: $text-color-primary;
      margin: 0 0 24px;
      text-align: center;
    }

    .steps-list {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 20px;

      .step-item {
        display: flex;
        align-items: flex-start;
        gap: 12px;
        padding: 20px;
        background: $bg-hover;
        border-radius: $border-radius-md;
        position: relative;

        .step-num {
          width: 48px;
          height: 48px;
          background: $primary-color;
          color: white;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 24px;
          font-weight: 700;
          flex-shrink: 0;
        }

        .step-content {
          flex: 1;
          min-width: 0;

          .step-name {
            font-size: 24px;
            font-weight: 700;
            color: $text-color-primary;
            margin-bottom: 6px;
          }

          .step-desc {
            font-size: 18px;
            color: $text-color-secondary;
            line-height: 1.5;
          }
        }
      }
    }
  }
}

.footer {
  display: flex;
  justify-content: center;
  gap: 24px;
  align-items: center;
  padding-top: 32px;
  border-top: 2px solid rgba(29, 108, 255, 0.1);
}
</style>
