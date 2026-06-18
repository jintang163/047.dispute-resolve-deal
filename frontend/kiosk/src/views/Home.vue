<template>
  <div class="home-page">
    <div class="header">
      <div class="logo-section">
        <div class="logo-icon">⚖️</div>
        <div class="logo-text">
          <h1>纠纷多元化解服务中心</h1>
          <p>Dispute Resolution Service Center</p>
        </div>
      </div>
      <div class="time-display">{{ currentTime }}</div>
    </div>

    <div class="content">
      <div class="welcome-section">
        <h2 class="welcome-title">欢迎您</h2>
        <p class="welcome-subtitle">请选择以下服务开始办理</p>
      </div>

      <div class="action-cards">
        <div class="action-card primary" @click="startRegistration">
          <div class="card-icon">📝</div>
          <div class="card-content">
            <h3>自助登记</h3>
            <p>纠纷信息在线登记</p>
            <p class="card-desc">快速录入您的纠纷信息，专业调解员将尽快与您联系</p>
          </div>
          <div class="card-arrow">→</div>
        </div>

        <div class="action-card secondary" @click="goToAIHelp">
          <div class="card-icon">🤖</div>
          <div class="card-content">
            <h3>AI法律咨询</h3>
            <p>智能法律助手</p>
            <p class="card-desc">7x24小时在线解答法律问题，提供专业的法律建议和指导</p>
          </div>
          <div class="card-arrow">→</div>
        </div>
      </div>

      <div class="info-section">
        <div class="info-item">
          <div class="info-icon">📞</div>
          <div>
            <div class="info-label">服务热线</div>
            <div class="info-value">12348</div>
          </div>
        </div>
        <div class="info-item">
          <div class="info-icon">⏰</div>
          <div>
            <div class="info-label">服务时间</div>
            <div class="info-value">周一至周日 8:00-20:00</div>
          </div>
        </div>
        <div class="info-item">
          <div class="info-icon">📍</div>
          <div>
            <div class="info-label">服务地址</div>
            <div class="info-value">纠纷调解服务中心大厅</div>
          </div>
        </div>
      </div>
    </div>

    <div class="footer">
      <p>请妥善保管您的案件编号，以便查询调解进度</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const currentTime = ref('')

let timer: ReturnType<typeof setInterval>

function updateTime() {
  const now = new Date()
  const options: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    weekday: 'long',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  }
  currentTime.value = now.toLocaleString('zh-CN', options)
}

function startRegistration() {
  router.push('/idcard')
}

function goToAIHelp() {
  router.push('/ai-help')
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
})

onUnmounted(() => {
  clearInterval(timer)
})
</script>

<style lang="scss" scoped>
.home-page {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 48px 64px;
  box-sizing: border-box;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 32px;
  border-bottom: 3px solid rgba(29, 108, 255, 0.1);

  .logo-section {
    display: flex;
    align-items: center;
    gap: 24px;

    .logo-icon {
      font-size: 80px;
    }

    .logo-text {
      h1 {
        font-size: 48px;
        font-weight: 700;
        color: $primary-color;
        margin: 0;
      }

      p {
        font-size: 22px;
        color: $text-color-secondary;
        margin: 8px 0 0;
        letter-spacing: 2px;
      }
    }
  }

  .time-display {
    font-size: 28px;
    color: $text-color-secondary;
    font-weight: 500;
    background: rgba(29, 108, 255, 0.08);
    padding: 16px 32px;
    border-radius: $border-radius-md;
  }
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 48px 0;

  .welcome-section {
    text-align: center;
    margin-bottom: 64px;

    .welcome-title {
      font-size: 64px;
      font-weight: 700;
      color: $text-color-primary;
      margin: 0 0 16px;
    }

    .welcome-subtitle {
      font-size: 32px;
      color: $text-color-secondary;
      margin: 0;
    }
  }

  .action-cards {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 48px;
    margin-bottom: 64px;

    .action-card {
      display: flex;
      align-items: center;
      gap: 32px;
      padding: 56px 48px;
      border-radius: $border-radius-xl;
      background: $bg-card;
      box-shadow: $shadow-card;
      cursor: pointer;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-8px);
        box-shadow: $shadow-hover;
      }

      &.primary {
        background: linear-gradient(135deg, $primary-color 0%, $primary-color-light 100%);
        color: white;

        .card-content {
          h3, p {
            color: white;
          }

          .card-desc {
            color: rgba(255, 255, 255, 0.85);
          }
        }

        .card-arrow {
          color: white;
        }
      }

      &.secondary {
        border: 3px solid rgba(29, 108, 255, 0.2);
      }

      .card-icon {
        font-size: 96px;
        flex-shrink: 0;
      }

      .card-content {
        flex: 1;

        h3 {
          font-size: 44px;
          font-weight: 700;
          margin: 0 0 8px;
          color: $text-color-primary;
        }

        p {
          font-size: 26px;
          color: $text-color-secondary;
          margin: 0 0 12px;
          font-weight: 500;
        }

        .card-desc {
          font-size: 22px;
          color: $text-color-light;
          line-height: 1.6;
          margin: 0;
        }
      }

      .card-arrow {
        font-size: 56px;
        color: $primary-color;
        font-weight: 300;
      }
    }
  }

  .info-section {
    display: flex;
    justify-content: center;
    gap: 80px;

    .info-item {
      display: flex;
      align-items: center;
      gap: 20px;

      .info-icon {
        font-size: 48px;
      }

      .info-label {
        font-size: 22px;
        color: $text-color-light;
        margin-bottom: 4px;
      }

      .info-value {
        font-size: 26px;
        color: $text-color-primary;
        font-weight: 600;
      }
    }
  }
}

.footer {
  text-align: center;
  padding-top: 32px;
  border-top: 2px solid rgba(29, 108, 255, 0.1);

  p {
    font-size: 24px;
    color: $text-color-secondary;
    margin: 0;
  }
}
</style>
