<template>
  <div class="form-page">
    <StepIndicator />

    <div class="content">
      <h2 class="page-title">请填写纠纷信息</h2>
      <div class="dispute-type-bar">
        <span class="label">当前纠纷类型：</span>
        <el-tag
          v-for="(node, index) in store.caseDraft.disputeTypePath"
          :key="index"
          type="primary"
          size="large"
          effect="light"
        >
          {{ node.name }}
        </el-tag>
      </div>

      <div class="form-container">
        <el-card class="section-card" shadow="hover">
          <template #header>
            <div class="section-header">
              <div class="section-icon">👤</div>
              <span class="section-title">对方当事人信息</span>
              <span class="section-tip">请尽可能填写完整，便于调解员联系</span>
            </div>
          </template>

          <el-form
            ref="formRef"
            :model="formData"
            :rules="formRules"
            label-width="180px"
            label-position="right"
            size="large"
          >
            <el-row :gutter="32">
              <el-col :span="12">
                <el-form-item label="对方姓名" prop="opponentName">
                  <el-input v-model="formData.opponentName" placeholder="请输入对方姓名或单位名称" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="联系电话" prop="opponentPhone">
                  <el-input v-model="formData.opponentPhone" placeholder="请输入对方联系电话" maxlength="11">
                    <template #prefix>
                      <span>📱</span>
                    </template>
                  </el-input>
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item label="联系地址" prop="opponentAddress">
              <el-input v-model="formData.opponentAddress" type="textarea" :rows="2" placeholder="请输入对方详细住址或单位地址" />
            </el-form-item>
          </el-form>
        </el-card>

        <el-card class="section-card" shadow="hover">
          <template #header>
            <div class="section-header">
              <div class="section-icon">📝</div>
              <span class="section-title">纠纷情况描述</span>
              <span class="section-tip">请详细描述纠纷的时间、经过、涉及金额等关键信息</span>
            </div>
          </template>

          <el-form :model="formData" :rules="formRules" label-width="180px" label-position="right" size="large">
            <el-form-item label="纠纷简述" prop="description">
              <el-input
                v-model="formData.description"
                type="textarea"
                :rows="8"
                maxlength="2000"
                show-word-limit
                placeholder="请详细描述纠纷情况：
1. 纠纷发生的时间、地点
2. 事情的起因和经过
3. 您受到的损失或影响
4. 您与对方的沟通情况"
                style="font-size: 28px"
              />
            </el-form-item>
          </el-form>

          <div class="quick-fill-section">
            <div class="quick-label">快速输入：</div>
            <div class="quick-tags">
              <el-tag
                v-for="(tag, index) in quickTags"
                :key="index"
                size="large"
                effect="plain"
                class="quick-tag"
                @click="appendQuickTag(tag)"
              >
                + {{ tag }}
              </el-tag>
            </div>
          </div>
        </el-card>

        <el-card class="section-card" shadow="hover">
          <template #header>
            <div class="section-header">
              <div class="section-icon">🎯</div>
              <span class="section-title">期望解决方式</span>
              <span class="section-tip">选择您最希望的解决方案，可多选</span>
            </div>
          </template>

          <el-form :model="formData" label-width="180px" label-position="right" size="large">
            <el-form-item label="期望方式" prop="expectedResolution">
              <el-checkbox-group v-model="selectedExpectations" size="large">
                <el-checkbox
                  v-for="item in expectationOptions"
                  :key="item.value"
                  :value="item.value"
                  class="expectation-checkbox"
                >
                  <div class="checkbox-content">
                    <span class="checkbox-icon">{{ item.icon }}</span>
                    <div>
                      <div class="checkbox-title">{{ item.label }}</div>
                      <div class="checkbox-desc">{{ item.desc }}</div>
                    </div>
                  </div>
                </el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-form>
        </el-card>
      </div>
    </div>

    <div class="footer">
      <TouchButton icon="ArrowLeft" size="large" @click="goBack">上一步</TouchButton>
      <TouchButton type="primary" icon="ArrowRight" size="xl" @click="handleNext">下一步：上传证据材料</TouchButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import StepIndicator from '@/components/StepIndicator.vue'
import TouchButton from '@/components/TouchButton.vue'
import { useKioskStore } from '@/stores/kiosk'
import { validatePhone } from '@/utils/kiosk'

const router = useRouter()
const store = useKioskStore()

const formRef = ref()
const selectedExpectations = ref<string[]>([])

const formData = reactive({
  opponentName: store.caseDraft.opponentName,
  opponentPhone: store.caseDraft.opponentPhone,
  opponentAddress: store.caseDraft.opponentAddress,
  description: store.caseDraft.description,
  expectedResolution: store.caseDraft.expectedResolution
})

const quickTags = [
  '对方拒绝沟通',
  '多次协商未果',
  '涉及金额较大',
  '需要立即处理',
  '对方态度恶劣',
  '证据齐全'
]

const expectationOptions = [
  { value: 'mediate', label: '调解解决', icon: '🤝', desc: '由专业调解员帮助双方协商' },
  { value: 'refund', label: '要求退款', icon: '💵', desc: '要求对方退还相关款项' },
  { value: 'compensate', label: '要求赔偿', icon: '💰', desc: '要求对方赔偿实际损失' },
  { value: 'apologize', label: '要求道歉', icon: '🙇', desc: '要求对方承认错误并道歉' },
  { value: 'correct', label: '要求整改', icon: '🔧', desc: '要求对方纠正不当行为' },
  { value: 'legal', label: '法律咨询', icon: '⚖️', desc: '希望了解相关法律途径' }
]

const formRules = {
  opponentName: [{ required: true, message: '请输入对方姓名或名称', trigger: 'blur' }],
  opponentPhone: [
    { required: true, message: '请输入联系电话', trigger: 'blur' },
    { validator: (_rule: any, value: string, callback: any) => {
      if (value && !validatePhone(value)) {
        callback(new Error('电话号码格式不正确'))
      } else {
        callback()
      }
    }, trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请填写纠纷情况描述', trigger: 'blur' },
    { min: 20, message: '请至少填写20个字，详细描述有助于调解', trigger: 'blur' }
  ]
}

watch(formData, (val) => {
  store.setCaseInfo({
    opponentName: val.opponentName,
    opponentPhone: val.opponentPhone,
    opponentAddress: val.opponentAddress,
    description: val.description,
    expectedResolution: selectedExpectations.value.join(',')
  })
}, { deep: true })

watch(selectedExpectations, (val) => {
  formData.expectedResolution = val.join(',')
  store.setCaseInfo({ expectedResolution: val.join(',') })
}, { deep: true })

function appendQuickTag(tag: string) {
  if (formData.description.length === 0) {
    formData.description = tag + '，'
  } else if (formData.description.endsWith('，') || formData.description.endsWith('。')) {
    formData.description += tag + '，'
  } else {
    formData.description += '；' + tag + '，'
  }
}

function goBack() {
  router.push('/dispute-type')
}

async function handleNext() {
  try {
    await formRef.value.validate()
    store.setCaseInfo({
      opponentName: formData.opponentName,
      opponentPhone: formData.opponentPhone,
      opponentAddress: formData.opponentAddress,
      description: formData.description,
      expectedResolution: selectedExpectations.value.join(',')
    })
    router.push('/evidence')
  } catch {
    ElMessage({
      message: '请完善必填信息',
      type: 'warning',
      duration: 3000
    })
  }
}
</script>

<style lang="scss" scoped>
.form-page {
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
    margin: 0 0 24px;
    text-align: center;
  }

  .dispute-type-bar {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-wrap: wrap;
    gap: 12px;
    padding: 20px 32px;
    background: rgba(29, 108, 255, 0.08);
    border-radius: $border-radius-lg;
    margin-bottom: 32px;

    .label {
      font-size: 26px;
      color: $text-color-secondary;
      margin-right: 8px;
    }
  }

  .form-container {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 24px;

    .section-card {
      :deep(.el-card__header) {
        padding: 24px 32px;
      }

      :deep(.el-card__body) {
        padding: 32px;
      }

      .section-header {
        display: flex;
        align-items: center;
        gap: 16px;

        .section-icon {
          font-size: 40px;
        }

        .section-title {
          font-size: 32px;
          font-weight: 700;
          color: $text-color-primary;
          flex: 1;
        }

        .section-tip {
          font-size: 22px;
          color: $text-color-light;
        }
      }

      .quick-fill-section {
        margin-top: 24px;
        padding-top: 24px;
        border-top: 2px dashed rgba(0, 0, 0, 0.1);

        .quick-label {
          font-size: 24px;
          color: $text-color-secondary;
          margin-bottom: 16px;
        }

        .quick-tags {
          display: flex;
          flex-wrap: wrap;
          gap: 12px;

          .quick-tag {
            cursor: pointer;
            padding: 12px 24px;
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

      .expectation-checkbox {
        margin-right: 0 !important;
        margin-bottom: 16px;
        width: calc(50% - 16px);

        &:nth-child(even) {
          margin-left: 32px;
        }

        .checkbox-content {
          display: flex;
          align-items: center;
          gap: 16px;
          padding: 8px 0;

          .checkbox-icon {
            font-size: 48px;
            flex-shrink: 0;
          }

          .checkbox-title {
            font-size: 28px;
            font-weight: 700;
            color: $text-color-primary;
            margin-bottom: 4px;
          }

          .checkbox-desc {
            font-size: 20px;
            color: $text-color-secondary;
          }
        }
      }
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
