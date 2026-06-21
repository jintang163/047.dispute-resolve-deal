<template>
  <div class="idcard-page">
    <StepIndicator />

    <div class="content">
      <h2 class="page-title">请将身份证放置在读卡器上</h2>
      <p class="page-subtitle">或选择手动填写身份信息</p>

      <div class="card-reader-section" :class="{ reading: isReading, success: readSuccess, error: readError }">
        <div class="reader-animation">
          <div class="reader-icon">{{ readSuccess ? '✅' : readError ? '❌' : '🪪' }}</div>
          <div v-if="isReading" class="scan-line"></div>
        </div>
        <div class="reader-status">
          <template v-if="isReading">
            <div class="status-title">正在读取身份证...</div>
            <div class="status-desc">请保持身份证平稳放置</div>
          </template>
          <template v-else-if="readSuccess">
            <div class="status-title success">身份证读取成功！</div>
            <div class="status-desc">信息已自动填入，请确认后继续</div>
          </template>
          <template v-else-if="readError">
            <div class="status-title error">读取失败</div>
            <div class="status-desc">{{ errorMessage }}，请重试或手动填写</div>
          </template>
          <template v-else>
            <div class="status-title">等待读取身份证</div>
            <div class="status-desc">请将身份证正面朝上放置在读卡器区域</div>
          </template>
        </div>
      </div>

      <div class="action-buttons">
        <TouchButton
          type="primary"
          size="xl"
          :icon="isReading ? 'Loading' : 'Camera'"
          :loading="isReading"
          :disabled="readSuccess"
          @click="handleReadCard"
        >
          {{ isReading ? '读取中...' : readSuccess ? '读取成功' : '读取身份证' }}
        </TouchButton>
        <TouchButton size="xl" icon="Edit" @click="showManualForm = !showManualForm">
          {{ showManualForm ? '收起手动填写' : '手动填写信息' }}
        </TouchButton>
      </div>

      <el-card v-if="showManualForm || readSuccess" class="info-card" shadow="hover">
        <template #header>
          <div class="card-header">
            <span class="card-title">身份信息</span>
            <el-tag size="large" :type="readSuccess ? 'success' : 'info'">
              {{ readSuccess ? '已读取' : '手动填写' }}
            </el-tag>
          </div>
        </template>

        <el-form
          ref="formRef"
          :model="formData"
          :rules="formRules"
          label-width="160px"
          label-position="right"
          size="large"
        >
          <el-row :gutter="32">
            <el-col :span="12">
              <el-form-item label="姓名" prop="name">
                <el-input v-model="formData.name" placeholder="请输入姓名" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="性别" prop="gender">
                <el-radio-group v-model="formData.gender" size="large">
                  <el-radio value="男">男</el-radio>
                  <el-radio value="女">女</el-radio>
                </el-radio-group>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="32">
            <el-col :span="12">
              <el-form-item label="民族" prop="nation">
                <el-input v-model="formData.nation" placeholder="请输入民族" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="出生日期" prop="birthDate">
                <el-date-picker
                  v-model="formData.birthDate"
                  type="date"
                  placeholder="选择出生日期"
                  format="YYYY-MM-DD"
                  value-format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="身份证号" prop="idNumber">
            <div class="idcard-input-wrapper">
              <el-input
                v-model="formData.idNumber"
                placeholder="请输入18位身份证号"
                maxlength="18"
                size="large"
                @blur="handleIDCardBlur"
              />
              <el-button
                type="primary"
                size="large"
                :icon="querying ? 'Loading' : 'Search'"
                :loading="querying"
                @click="handleQueryPopulation"
                class="query-btn"
              >
                {{ querying ? '查询中' : '人口信息查询' }}
              </el-button>
            </div>
            <div v-if="queryTip" class="query-tip" :class="queryTip.type">
              {{ queryTip.icon }} {{ queryTip.message }}
            </div>
          </el-form-item>

          <el-form-item label="住址" prop="address">
            <el-input v-model="formData.address" type="textarea" :rows="2" placeholder="请输入详细住址" />
          </el-form-item>

          <el-row :gutter="32">
            <el-col :span="12">
              <el-form-item label="签发机关" prop="issuer">
                <el-input v-model="formData.issuer" placeholder="请输入签发机关" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="有效期限" prop="validPeriod">
                <el-input v-model="formData.validPeriod" placeholder="如：2020.01.01-2040.01.01" />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
      </el-card>
    </div>

    <div class="footer">
      <TouchButton icon="ArrowLeft" size="large" @click="goBack">返回首页</TouchButton>
      <TouchButton type="primary" icon="ArrowRight" size="xl" @click="handleNext">下一步：选择纠纷类型</TouchButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import StepIndicator from '@/components/StepIndicator.vue'
import TouchButton from '@/components/TouchButton.vue'
import { useKioskStore } from '@/stores/kiosk'
import { readIdCard, stopCardReading, validateIdNumber } from '@/utils/kiosk'
import { kioskApi } from '@/services/kiosk'

const router = useRouter()
const store = useKioskStore()

const isReading = ref(false)
const readSuccess = ref(false)
const readError = ref(false)
const errorMessage = ref('')
const showManualForm = ref(false)
const formRef = ref()

const querying = ref(false)
const queryTip = ref<{ type: string; icon: string; message: string } | null>(null)

const formData = reactive({
  name: store.idCardInfo.name,
  gender: store.idCardInfo.gender,
  nation: store.idCardInfo.nation,
  birthDate: store.idCardInfo.birthDate,
  address: store.idCardInfo.address,
  idNumber: store.idCardInfo.idNumber,
  issuer: store.idCardInfo.issuer,
  validPeriod: store.idCardInfo.validPeriod
})

const formRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  gender: [{ required: true, message: '请选择性别', trigger: 'change' }],
  idNumber: [
    { required: true, message: '请输入身份证号', trigger: 'blur' },
    { validator: (_rule: any, value: string, callback: any) => {
      if (!value) {
        callback(new Error('请输入身份证号'))
      } else if (!validateIdNumber(value)) {
        callback(new Error('身份证号格式不正确'))
      } else {
        callback()
      }
    }, trigger: 'blur' }
  ],
  address: [{ required: true, message: '请输入住址', trigger: 'blur' }]
}

watch(formData, (val) => {
  store.setIdCardInfo(val)
}, { deep: true })

async function handleQueryPopulation() {
  const idNumber = formData.idNumber
  if (!idNumber) {
    ElMessage({ message: '请先输入身份证号', type: 'warning', duration: 3000 })
    return
  }
  if (!validateIdNumber(idNumber)) {
    ElMessage({ message: '身份证号格式不正确', type: 'error', duration: 3000 })
    return
  }

  querying.value = true
  queryTip.value = { type: 'info', icon: 'ⓘ', message: '正在查询人口库信息，请稍候...' }

  try {
    const res = await kioskApi.queryPopulationByIDCard(idNumber)
    const data: any = (res as any)?.data ?? res

    if (data?.name) {
      if (!formData.name) formData.name = data.name
      if (!formData.gender) formData.gender = data.genderName || (data.gender === 1 ? '男' : '女')
      if (!formData.nation) formData.nation = data.nation
      if (!formData.birthDate) formData.birthDate = data.birthDate
      if (!formData.address) formData.address = data.address
      if (!formData.issuer) formData.issuer = data.issuer
      if (!formData.validPeriod) formData.validPeriod = data.validPeriod

      store.setIdCardInfo(formData)

      queryTip.value = { type: 'success', icon: '✓', message: '人口信息查询成功，已自动预填！' }
      ElMessage({ message: '人口信息查询成功，已自动预填', type: 'success', duration: 3000 })
      readSuccess.value = true
    } else {
      queryTip.value = { type: 'error', icon: '✗', message: '未查询到该身份证号的人口信息，请手动填写' }
      ElMessage({ message: '未查询到人口信息，请手动填写', type: 'warning', duration: 3000 })
    }
  } catch (error: any) {
    queryTip.value = { type: 'error', icon: '✗', message: `查询失败：${error.message || '请稍后重试'}` }
    ElMessage({ message: error.message || '人口库查询失败', type: 'error', duration: 3000 })
  } finally {
    querying.value = false
    setTimeout(() => {
      if (queryTip.value?.type === 'success') {
        queryTip.value = null
      }
    }, 5000)
  }
}

function handleIDCardBlur() {
  if (formData.idNumber && validateIdNumber(formData.idNumber) && !formData.name) {
    handleQueryPopulation()
  }
}

async function handleReadCard() {
  if (isReading.value) return
  isReading.value = true
  readError.value = false
  errorMessage.value = ''

  try {
    const result = await readIdCard()
    if (result.success && result.data) {
      Object.assign(formData, result.data)
      store.setIdCardInfo(result.data)
      readSuccess.value = true
      showManualForm.value = true
      ElMessage({
        message: '身份证读取成功',
        type: 'success',
        duration: 3000
      })
    } else {
      readError.value = true
      errorMessage.value = result.error || '请重试'
    }
  } catch (e) {
    readError.value = true
    errorMessage.value = '读取异常，请重试'
  } finally {
    isReading.value = false
  }
}

function goBack() {
  stopCardReading()
  router.push('/')
}

async function handleNext() {
  try {
    await formRef.value.validate()
    store.setIdCardInfo(formData)
    router.push('/dispute-type')
  } catch {
    ElMessage({
      message: '请完善必填信息',
      type: 'warning',
      duration: 3000
    })
  }
}

onMounted(() => {
  if (store.idCardInfo.name) {
    showManualForm.value = true
    readSuccess.value = true
  }
})
</script>

<style lang="scss" scoped>
.idcard-page {
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
  align-items: center;
  padding: 32px 0;
  overflow-y: auto;

  .page-title {
    font-size: 48px;
    font-weight: 700;
    color: $text-color-primary;
    margin: 0 0 12px;
  }

  .page-subtitle {
    font-size: 28px;
    color: $text-color-secondary;
    margin: 0 0 48px;
  }

  .card-reader-section {
    width: 600px;
    padding: 56px;
    background: $bg-card;
    border-radius: $border-radius-xl;
    box-shadow: $shadow-card;
    text-align: center;
    border: 4px dashed rgba(29, 108, 255, 0.3);
    transition: all 0.3s ease;
    margin-bottom: 48px;

    &.reading {
      border-color: $primary-color;
      background: rgba(29, 108, 255, 0.05);
    }

    &.success {
      border-color: $success-color;
      border-style: solid;
      background: rgba(34, 197, 94, 0.05);
    }

    &.error {
      border-color: $danger-color;
      border-style: solid;
      background: rgba(239, 68, 68, 0.05);
    }

    .reader-animation {
      position: relative;
      width: 280px;
      height: 180px;
      margin: 0 auto 32px;
      background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
      border-radius: 20px;
      display: flex;
      align-items: center;
      justify-content: center;
      overflow: hidden;
      box-shadow: inset 0 2px 10px rgba(0, 0, 0, 0.1);

      .reader-icon {
        font-size: 100px;
        z-index: 2;
      }

      .scan-line {
        position: absolute;
        left: 0;
        right: 0;
        height: 4px;
        background: linear-gradient(90deg, transparent, $primary-color, transparent);
        animation: scan 1.5s ease-in-out infinite;
        z-index: 1;
      }
    }

    @keyframes scan {
      0% { top: 0; }
      50% { top: calc(100% - 4px); }
      100% { top: 0; }
    }

    .reader-status {
      .status-title {
        font-size: 32px;
        font-weight: 700;
        color: $text-color-primary;
        margin-bottom: 8px;

        &.success {
          color: $success-color;
        }

        &.error {
          color: $danger-color;
        }
      }

      .status-desc {
        font-size: 24px;
        color: $text-color-secondary;
        margin: 0;
      }
    }
  }

  .action-buttons {
    display: flex;
    gap: 32px;
    margin-bottom: 48px;
  }

  .info-card {
    width: 100%;
    max-width: 1000px;
    margin-bottom: 32px;

    :deep(.el-card__header) {
      padding: 24px 32px;
    }

    :deep(.el-card__body) {
      padding: 32px;
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .card-title {
        font-size: 32px;
        font-weight: 700;
        color: $text-color-primary;
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

.idcard-input-wrapper {
  display: flex;
  gap: 16px;
  align-items: center;

  .query-btn {
    min-width: 200px;
    height: 52px;
    font-size: 20px;
    flex-shrink: 0;
  }
}

.query-tip {
  margin-top: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 22px;
  font-weight: 500;

  &.info {
    background: rgba(24, 144, 255, 0.1);
    color: #1890ff;
  }

  &.success {
    background: rgba(82, 196, 26, 0.1);
    color: #52c41a;
  }

  &.error {
    background: rgba(255, 77, 79, 0.1);
    color: #ff4d4f;
  }
}
</style>
