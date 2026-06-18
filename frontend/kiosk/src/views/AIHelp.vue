<template>
  <div class="aihelp-page">
    <div class="header">
      <TouchButton icon="ArrowLeft" size="large" @click="goBack">返回</TouchButton>
      <div class="header-center">
        <div class="ai-icon">🤖</div>
        <h1>AI 法律咨询助手</h1>
      </div>
      <TouchButton icon="RefreshLeft" size="large" @click="clearChat">清空对话</TouchButton>
    </div>

    <div class="chat-container" ref="chatContainerRef">
      <div v-if="messages.length === 0" class="welcome-section">
        <div class="welcome-avatar">⚖️</div>
        <h2 class="welcome-title">您好，我是法律AI助手</h2>
        <p class="welcome-desc">我可以为您解答常见法律问题，提供法律知识咨询</p>

        <div class="quick-questions">
          <div class="qq-title">常见问题：</div>
          <div class="qq-grid">
            <div
              v-for="(q, index) in quickQuestions"
              :key="index"
              class="qq-card"
              @click="sendMessage(q)"
            >
              <span class="qq-icon">💡</span>
              <span class="qq-text">{{ q }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="messages-list">
        <div
          v-for="msg in messages"
          :key="msg.id"
          class="message-item"
          :class="msg.role"
        >
          <div class="avatar" v-if="msg.role === 'assistant'">🤖</div>
          <div class="message-bubble">
            <div v-if="msg.role === 'assistant' && msg.loading" class="typing-indicator">
              <span></span>
              <span></span>
              <span></span>
            </div>
            <div v-else class="message-content" v-html="formatMessage(msg.content)"></div>
            <div class="message-time">{{ msg.timestamp }}</div>
          </div>
          <div class="avatar" v-if="msg.role === 'user'">👤</div>
        </div>
      </div>
    </div>

    <div class="input-section">
      <div class="input-wrapper">
        <el-input
          v-model="inputText"
          type="textarea"
          :rows="3"
          placeholder="请输入您的法律问题..."
          resize="none"
          size="large"
          @keydown.ctrl.enter="handleSend"
          @keydown.meta.enter="handleSend"
        />
      </div>
      <div class="input-actions">
        <div class="quick-btns">
          <el-tag
            v-for="tag in inputQuickTags"
            :key="tag"
            size="large"
            effect="plain"
            class="quick-tag"
            @click="appendTag(tag)"
          >
            {{ tag }}
          </el-tag>
        </div>
        <TouchButton
          type="primary"
          icon="Promotion"
          size="xl"
          :loading="sending"
          :disabled="!inputText.trim()"
          @click="handleSend"
        >
          发送
        </TouchButton>
      </div>
      <div class="input-tip">
        提示：按 Ctrl+Enter 可快速发送问题
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import TouchButton from '@/components/TouchButton.vue'
import kioskApi, { type AIConversationMessage } from '@/services/kiosk'

const router = useRouter()
const chatContainerRef = ref<HTMLElement>()
const messages = ref<AIConversationMessage[]>([])
const conversationId = ref<string>('')
const inputText = ref('')
const sending = ref(false)

const quickQuestions = ref([
  '公司拖欠工资怎么办？',
  '邻里噪音纠纷如何处理？',
  '合同到期不续签有赔偿吗？',
  '交通事故责任怎么认定？',
  '借贷没有借条能起诉吗？',
  '离婚财产怎么分割？'
])

const inputQuickTags = ['你好', '谢谢', '请解释一下', '有什么建议']

function formatMessage(content: string): string {
  return content
    .replace(/\n/g, '<br/>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
}

async function scrollToBottom() {
  await nextTick()
  if (chatContainerRef.value) {
    chatContainerRef.value.scrollTop = chatContainerRef.value.scrollHeight
  }
}

function getTimestamp(): string {
  return new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2, 9)
}

async function sendMessage(text: string) {
  if (!text.trim() || sending.value) return

  const userMsg: AIConversationMessage = {
    id: generateId(),
    role: 'user',
    content: text,
    timestamp: getTimestamp()
  }
  messages.value.push(userMsg)
  inputText.value = ''
  await scrollToBottom()

  const loadingMsg: AIConversationMessage = {
    id: generateId(),
    role: 'assistant',
    content: '',
    loading: true as any,
    timestamp: getTimestamp()
  }
  messages.value.push(loadingMsg)
  sending.value = true

  try {
    const result = await kioskApi.sendAIMessage(text, conversationId.value)
    conversationId.value = result.conversationId

    const idx = messages.value.findIndex(m => m.id === loadingMsg.id)
    if (idx >= 0) {
      messages.value[idx] = {
        ...loadingMsg,
        content: result.reply,
        loading: false as any,
        timestamp: getTimestamp()
      }
    }
  } catch (e) {
    const idx = messages.value.findIndex(m => m.id === loadingMsg.id)
    if (idx >= 0) {
      messages.value[idx] = {
        ...loadingMsg,
        content: getMockReply(text),
        loading: false as any,
        timestamp: getTimestamp()
      }
    }
  } finally {
    sending.value = false
    scrollToBottom()
  }
}

function getMockReply(question: string): string {
  const mockReplies: Record<string, string> = {
    '公司拖欠工资怎么办？': `您好，针对公司拖欠工资问题，建议您：

**1. 收集证据**：劳动合同、工资条、考勤记录、工作聊天记录等
**2. 协商解决**：先与公司协商，要求支付工资
**3. 向劳动监察部门投诉**：可拨打12333或到当地人力资源和社会保障局投诉
**4. 申请劳动仲裁**：携带证据到劳动争议仲裁委员会申请仲裁（免费）
**5. 诉讼途径**：对仲裁结果不满可向法院起诉

**法律依据**：《劳动合同法》第三十条、第八十五条，用人单位应当按照劳动合同约定和国家规定，向劳动者及时足额支付劳动报酬。

您需要到调解中心进行登记吗？我们可以协助您进行调解。`,
    '邻里噪音纠纷如何处理？': `您好，邻里噪音纠纷建议您按以下步骤处理：

**1. 友好沟通**：首先尝试与邻居友好协商，说明噪音对您的影响
**2. 收集证据**：录音、录像，记录噪音发生的时间、频率和持续时长
**3. 物业/居委会调解**：请物业或社区居委会出面调解
**4. 报警处理**：严重时可拨打110报警，公安机关可给予警告
**5. 法律途径**：可向法院起诉，要求停止侵害、赔偿损失

**法律依据**：《治安管理处罚法》第五十八条、《民法典》第二百八十八条（相邻关系）

建议您先到调解中心登记，由专业调解员帮助双方沟通解决。`,
    '合同到期不续签有赔偿吗？': `您好，劳动合同到期不续签是否有补偿，分情况而定：

**单位不续签**：
- 应当支付经济补偿金
- 补偿标准：每满1年支付1个月工资
- 6个月以上不满1年按1年算，不满6个月支付半个月工资

**员工不续签**：
- 单位维持或提高原条件，员工不续签：无补偿
- 单位降低原条件，员工不续签：有补偿

**注意**：
- 如果连续签订2次固定期限合同，员工有权要求签订无固定期限合同
- 建议在到期前30天确认意向

您可以在调解中心登记，进一步咨询专业律师意见。`
  }

  return mockReplies[question] || `感谢您的提问！

针对您的问题，建议：

1. 先收集与问题相关的**证据材料**
2. 可以先尝试与对方**友好协商**解决
3. 协商不成可到调解中心进行**登记调解**
4. 调解不成可通过**法律途径**维护权益

您可以将具体情况详细说明，我会为您提供更有针对性的建议。

如需进一步帮助，可返回首页进行纠纷登记，或拨打服务热线 **12348**。`
}

function handleSend() {
  sendMessage(inputText.value)
}

function appendTag(tag: string) {
  inputText.value = inputText.value ? `${inputText.value} ${tag}` : tag
}

function clearChat() {
  messages.value = []
  conversationId.value = ''
}

function goBack() {
  router.back()
}

async function loadQuickQuestions() {
  try {
    const data = await kioskApi.getAIQuickQuestions()
    if (data && data.length > 0) {
      quickQuestions.value = data
    }
  } catch {}
}

onMounted(() => {
  loadQuickQuestions()
})
</script>

<style lang="scss" scoped>
.aihelp-page {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 24px 32px;
  box-sizing: border-box;
  background: linear-gradient(180deg, #f0f7ff 0%, #e6f0ff 100%);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-shrink: 0;

  .header-center {
    display: flex;
    align-items: center;
    gap: 16px;

    .ai-icon {
      font-size: 56px;
    }

    h1 {
      font-size: 40px;
      font-weight: 700;
      color: $primary-color;
      margin: 0;
    }
  }
}

.chat-container {
  flex: 1;
  overflow-y: auto;
  background: $bg-card;
  border-radius: $border-radius-xl;
  padding: 32px;
  box-shadow: $shadow-card;
  margin-bottom: 24px;

  .welcome-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    text-align: center;

    .welcome-avatar {
      font-size: 120px;
      margin-bottom: 24px;
    }

    .welcome-title {
      font-size: 44px;
      font-weight: 700;
      color: $text-color-primary;
      margin: 0 0 12px;
    }

    .welcome-desc {
      font-size: 26px;
      color: $text-color-secondary;
      margin: 0 0 48px;
    }

    .quick-questions {
      width: 100%;
      max-width: 1000px;

      .qq-title {
        font-size: 26px;
        color: $text-color-secondary;
        margin-bottom: 20px;
        text-align: left;
      }

      .qq-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 20px;

        .qq-card {
          display: flex;
          align-items: center;
          gap: 16px;
          padding: 28px 32px;
          background: $bg-hover;
          border: 2px solid transparent;
          border-radius: $border-radius-lg;
          cursor: pointer;
          transition: all 0.2s ease;

          &:hover {
            background: rgba(29, 108, 255, 0.1);
            border-color: rgba(29, 108, 255, 0.3);
            transform: translateY(-2px);
          }

          .qq-icon {
            font-size: 36px;
            flex-shrink: 0;
          }

          .qq-text {
            font-size: 26px;
            color: $text-color-primary;
            font-weight: 500;
            text-align: left;
          }
        }
      }
    }
  }

  .messages-list {
    display: flex;
    flex-direction: column;
    gap: 32px;

    .message-item {
      display: flex;
      gap: 16px;
      align-items: flex-start;

      &.user {
        flex-direction: row-reverse;
      }

      .avatar {
        width: 64px;
        height: 64px;
        border-radius: 50%;
        background: $bg-hover;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 36px;
        flex-shrink: 0;
      }

      .message-bubble {
        max-width: 70%;
        padding: 24px 28px;
        border-radius: $border-radius-lg;
        position: relative;

        .typing-indicator {
          display: flex;
          gap: 8px;
          padding: 8px 0;

          span {
            width: 14px;
            height: 14px;
            background: $text-color-light;
            border-radius: 50%;
            animation: typing 1.4s infinite;

            &:nth-child(2) { animation-delay: 0.2s; }
            &:nth-child(3) { animation-delay: 0.4s; }
          }

          @keyframes typing {
            0%, 60%, 100% { opacity: 0.3; transform: translateY(0); }
            30% { opacity: 1; transform: translateY(-6px); }
          }
        }

        .message-content {
          font-size: 26px;
          line-height: 1.8;
          color: $text-color-primary;
          word-break: break-word;

          :deep(strong) {
            color: $primary-color;
          }
        }

        .message-time {
          font-size: 18px;
          color: $text-color-light;
          margin-top: 12px;
          text-align: right;
        }
      }

      &.assistant .message-bubble {
        background: $bg-hover;
        border-top-left-radius: 4px;
      }

      &.user .message-bubble {
        background: $primary-color;
        border-top-right-radius: 4px;

        .message-content {
          color: white;
        }

        .message-time {
          color: rgba(255, 255, 255, 0.7);
        }
      }
    }
  }
}

.input-section {
  background: $bg-card;
  border-radius: $border-radius-xl;
  padding: 24px 32px;
  box-shadow: $shadow-card;
  flex-shrink: 0;

  .input-wrapper {
    margin-bottom: 16px;

    :deep(.el-textarea__inner) {
      font-size: 28px !important;
      padding: 20px !important;
      line-height: 1.6 !important;
    }
  }

  .input-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 24px;

    .quick-btns {
      display: flex;
      gap: 12px;
      flex-wrap: wrap;
      flex: 1;

      .quick-tag {
        cursor: pointer;
        padding: 8px 20px;
        font-size: 22px;
        transition: all 0.2s ease;

        &:hover {
          background: $primary-color;
          color: white;
          border-color: $primary-color;
        }
      }
    }
  }

  .input-tip {
    font-size: 20px;
    color: $text-color-light;
    margin-top: 12px;
    text-align: right;
  }
}
</style>
