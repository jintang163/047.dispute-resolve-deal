<template>
  <div class="evidence-page">
    <StepIndicator />

    <div class="content">
      <h2 class="page-title">上传证据材料</h2>
      <p class="page-tip">请上传与纠纷相关的证据材料，支持图片、文档等格式（可选）</p>

      <div class="upload-section">
        <el-upload
          ref="uploadRef"
          class="upload-area"
          drag
          :auto-upload="false"
          :on-change="handleFileChange"
          :on-remove="handleFileRemove"
          :file-list="fileList"
          accept="image/*,.pdf,.doc,.docx,.xls,.xlsx,.txt"
          multiple
        >
          <div class="upload-content">
            <div class="upload-icon">📁</div>
            <div class="upload-text">
              <p class="upload-title">点击或拖拽文件到此处上传</p>
              <p class="upload-desc">支持 JPG、PNG、PDF、Word、Excel 等格式</p>
              <p class="upload-limit">单个文件不超过 50MB，最多可上传 20 个文件</p>
            </div>
          </div>
        </el-upload>

        <div class="quick-buttons">
          <TouchButton type="primary" icon="Camera" size="large" @click="handleTakePhoto">
            现场拍照
          </TouchButton>
          <TouchButton icon="Picture" size="large" @click="triggerFileInput">
            选择图片
          </TouchButton>
          <TouchButton icon="Document" size="large" @click="triggerDocumentInput">
            选择文档
          </TouchButton>
        </div>

        <input
          ref="photoInputRef"
          type="file"
          accept="image/*"
          capture="environment"
          style="display: none"
          @change="handleCapturePhoto"
        />
        <input
          ref="imageInputRef"
          type="file"
          accept="image/*"
          multiple
          style="display: none"
          @change="handleSelectFiles"
        />
        <input
          ref="documentInputRef"
          type="file"
          accept=".pdf,.doc,.docx,.xls,.xlsx,.txt"
          multiple
          style="display: none"
          @change="handleSelectFiles"
        />
      </div>

      <div v-if="evidenceList.length > 0" class="evidence-list-section">
        <div class="list-header">
          <h3 class="list-title">
            已上传材料
            <el-tag type="info" size="large" effect="light">{{ evidenceList.length }} 个文件</el-tag>
          </h3>
          <TouchButton type="danger" icon="Delete" size="medium" @click="clearAll">
            清空全部
          </TouchButton>
        </div>

        <div class="evidence-grid">
          <div
            v-for="item in evidenceList"
            :key="item.id"
            class="evidence-card"
          >
            <div class="evidence-preview">
              <template v-if="item.type === 'image'">
                <img :src="item.url" alt="" class="preview-image" />
              </template>
              <template v-else>
                <div class="file-placeholder">
                  <span class="file-icon">{{ getFileIcon(item.type) }}</span>
                  <span class="file-extension">{{ getFileExtension(item.name) }}</span>
                </div>
              </template>
            </div>
            <div class="evidence-info">
              <div class="evidence-name" :title="item.name">{{ item.name }}</div>
              <div class="evidence-meta">
                <span>{{ formatSize(item.size) }}</span>
                <span>·</span>
                <span>{{ item.uploadTime }}</span>
              </div>
            </div>
            <div class="evidence-actions">
              <el-button size="large" circle @click="previewItem(item)">
                <el-icon><ZoomIn /></el-icon>
              </el-button>
              <el-button size="large" circle type="danger" @click="removeItem(item)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <div v-if="uploadingMap[item.id]" class="uploading-overlay">
              <el-progress type="circle" :percentage="uploadingMap[item.id]" :width="100" />
              <p>上传中...</p>
            </div>
          </div>
        </div>
      </div>

      <el-empty v-else description="暂无证据材料，您也可以跳过此步骤" class="empty-state" />

      <el-dialog v-model="previewVisible" title="预览" width="90%" top="5vh">
        <div class="preview-container" v-if="currentPreview">
          <img v-if="currentPreview.type === 'image'" :src="currentPreview.url" class="preview-full" />
          <div v-else class="doc-preview">
            <div class="doc-preview-icon">{{ getFileIcon(currentPreview.type) }}</div>
            <div class="doc-preview-name">{{ currentPreview.name }}</div>
            <div class="doc-preview-tip">文档文件请下载后查看</div>
            <TouchButton type="primary" icon="Download" size="large" @click="downloadCurrent">
              下载文件
            </TouchButton>
          </div>
        </div>
      </el-dialog>
    </div>

    <div class="footer">
      <TouchButton icon="ArrowLeft" size="large" @click="goBack">上一步</TouchButton>
      <div class="footer-right">
        <TouchButton icon="DArrowRight" size="large" @click="skipStep">跳过此步</TouchButton>
        <TouchButton type="primary" icon="ArrowRight" size="xl" @click="handleNext">下一步：确认信息</TouchButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type UploadFile, type UploadFiles } from 'element-plus'
import { ZoomIn, Delete } from '@element-plus/icons-vue'
import StepIndicator from '@/components/StepIndicator.vue'
import TouchButton from '@/components/TouchButton.vue'
import { useKioskStore, type EvidenceItem } from '@/stores/kiosk'
import { formatFileSize, generateId } from '@/utils/kiosk'
import kioskApi from '@/services/kiosk'

const router = useRouter()
const store = useKioskStore()

const uploadRef = ref()
const photoInputRef = ref<HTMLInputElement>()
const imageInputRef = ref<HTMLInputElement>()
const documentInputRef = ref<HTMLInputElement>()

const fileList = ref<UploadFile[]>([])
const evidenceList = ref<EvidenceItem[]>([...store.caseDraft.evidenceList])
const uploadingMap = ref<Record<string, number>>({})

const previewVisible = ref(false)
const currentPreview = ref<EvidenceItem | null>(null)

function getFileIcon(type: string): string {
  const icons: Record<string, string> = {
    image: '🖼️',
    document: '📄',
    pdf: '📕',
    video: '🎬',
    audio: '🎵'
  }
  return icons[type] || '📁'
}

function getFileExtension(name: string): string {
  const idx = name.lastIndexOf('.')
  return idx >= 0 ? name.substring(idx + 1).toUpperCase() : ''
}

function getFileType(name: string): EvidenceItem['type'] {
  const ext = name.substring(name.lastIndexOf('.') + 1).toLowerCase()
  const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'heic']
  const videoExts = ['mp4', 'avi', 'mov', 'wmv', 'flv', 'mkv']
  const audioExts = ['mp3', 'wav', 'wma', 'aac', 'flac']
  if (imageExts.includes(ext)) return 'image'
  if (videoExts.includes(ext)) return 'video'
  if (audioExts.includes(ext)) return 'audio'
  return 'document'
}

function formatSize(size: number): string {
  return formatFileSize(size)
}

async function processFile(file: File) {
  if (evidenceList.value.length >= 20) {
    ElMessage.warning('最多可上传20个文件')
    return
  }
  if (file.size > 50 * 1024 * 1024) {
    ElMessage.warning(`${file.name} 超过50MB限制`)
    return
  }

  const tempId = generateId()
  const type = getFileType(file.name)

  let url = ''
  if (type === 'image') {
    url = URL.createObjectURL(file)
  }

  const tempItem: EvidenceItem = {
    id: tempId,
    name: file.name,
    type,
    url,
    size: file.size,
    uploadTime: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  evidenceList.value.push(tempItem)
  uploadingMap.value[tempId] = 0

  try {
    const result = await kioskApi.uploadEvidence(file, (percent) => {
      uploadingMap.value[tempId] = percent
    })

    const idx = evidenceList.value.findIndex(e => e.id === tempId)
    if (idx >= 0) {
      evidenceList.value[idx] = { ...result, id: tempId, uploadTime: tempItem.uploadTime }
      store.addEvidence(evidenceList.value[idx])
    }
    delete uploadingMap.value[tempId]
  } catch (e) {
    delete uploadingMap.value[tempId]
    const idx = evidenceList.value.findIndex(e => e.id === tempId)
    if (idx >= 0) {
      evidenceList.value.splice(idx, 1)
    }
    ElMessage.error(`${file.name} 上传失败`)
  }
}

function handleFileChange(uploadFile: UploadFile, uploadFiles: UploadFiles) {
  if (uploadFile.raw) {
    processFile(uploadFile.raw as File)
  }
}

function handleFileRemove(uploadFile: UploadFile) {
  const idx = evidenceList.value.findIndex(e => e.name === uploadFile.name)
  if (idx >= 0) {
    store.removeEvidence(evidenceList.value[idx].id)
    evidenceList.value.splice(idx, 1)
  }
}

function handleTakePhoto() {
  photoInputRef.value?.click()
}

function triggerFileInput() {
  imageInputRef.value?.click()
}

function triggerDocumentInput() {
  documentInputRef.value?.click()
}

function handleCapturePhoto(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files) {
    Array.from(target.files).forEach(f => processFile(f))
    target.value = ''
  }
}

function handleSelectFiles(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files) {
    Array.from(target.files).forEach(f => processFile(f))
    target.value = ''
  }
}

function previewItem(item: EvidenceItem) {
  currentPreview.value = item
  previewVisible.value = true
}

function downloadCurrent() {
  if (currentPreview.value?.url) {
    window.open(currentPreview.value.url, '_blank')
  }
}

function removeItem(item: EvidenceItem) {
  ElMessageBox.confirm(`确定删除 ${item.name} 吗？`, '确认删除', {
    confirmButtonText: '删除',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    store.removeEvidence(item.id)
    evidenceList.value = evidenceList.value.filter(e => e.id !== item.id)
    ElMessage.success('已删除')
  }).catch(() => {})
}

async function clearAll() {
  if (evidenceList.value.length === 0) return
  try {
    await ElMessageBox.confirm('确定清空所有已上传的证据材料吗？', '确认清空', {
      confirmButtonText: '清空',
      cancelButtonText: '取消',
      type: 'warning'
    })
    evidenceList.value.forEach(e => store.removeEvidence(e.id))
    evidenceList.value = []
    fileList.value = []
    ElMessage.success('已清空')
  } catch {}
}

function goBack() {
  router.push('/form')
}

function skipStep() {
  router.push('/confirm')
}

function handleNext() {
  router.push('/confirm')
}

onMounted(() => {
  evidenceList.value.forEach(e => store.addEvidence(e))
})
</script>

<style lang="scss" scoped>
.evidence-page {
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

  .upload-section {
    margin-bottom: 32px;

    .upload-area {
      margin-bottom: 24px;

      :deep(.el-upload-dragger) {
        min-height: 280px;
        border: 4px dashed rgba(29, 108, 255, 0.3);
        border-radius: $border-radius-xl;
        transition: all 0.3s ease;

        &:hover {
          border-color: $primary-color;
          background: rgba(29, 108, 255, 0.03);
        }
      }

      .upload-content {
        padding: 48px;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 48px;

        .upload-icon {
          font-size: 120px;
          opacity: 0.8;
        }

        .upload-text {
          text-align: left;

          .upload-title {
            font-size: 36px;
            font-weight: 700;
            color: $text-color-primary;
            margin: 0 0 12px;
          }

          .upload-desc {
            font-size: 26px;
            color: $text-color-secondary;
            margin: 0 0 8px;
          }

          .upload-limit {
            font-size: 22px;
            color: $text-color-light;
            margin: 0;
          }
        }
      }
    }

    .quick-buttons {
      display: flex;
      justify-content: center;
      gap: 24px;
    }
  }

  .evidence-list-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;

    .list-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 24px;
      flex-shrink: 0;

      .list-title {
        font-size: 32px;
        font-weight: 700;
        color: $text-color-primary;
        margin: 0;
        display: flex;
        align-items: center;
        gap: 16px;
      }
    }

    .evidence-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
      gap: 20px;
      overflow-y: auto;
      padding: 8px;

      .evidence-card {
        position: relative;
        background: $bg-card;
        border-radius: $border-radius-lg;
        box-shadow: $shadow-card;
        overflow: hidden;
        transition: all 0.3s ease;

        &:hover {
          transform: translateY(-4px);
          box-shadow: $shadow-hover;
        }

        .evidence-preview {
          width: 100%;
          height: 200px;
          background: $bg-hover;
          display: flex;
          align-items: center;
          justify-content: center;
          overflow: hidden;

          .preview-image {
            width: 100%;
            height: 100%;
            object-fit: cover;
          }

          .file-placeholder {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 8px;

            .file-icon {
              font-size: 72px;
            }

            .file-extension {
              font-size: 20px;
              font-weight: 700;
              color: $primary-color;
              background: rgba(29, 108, 255, 0.1);
              padding: 4px 16px;
              border-radius: $border-radius-sm;
            }
          }
        }

        .evidence-info {
          padding: 16px 20px;

          .evidence-name {
            font-size: 22px;
            font-weight: 600;
            color: $text-color-primary;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            margin-bottom: 8px;
          }

          .evidence-meta {
            font-size: 18px;
            color: $text-color-light;
            display: flex;
            gap: 8px;
          }
        }

        .evidence-actions {
          position: absolute;
          top: 12px;
          right: 12px;
          display: flex;
          gap: 8px;
          opacity: 0;
          transition: opacity 0.2s ease;
        }

        &:hover .evidence-actions {
          opacity: 1;
        }

        .uploading-overlay {
          position: absolute;
          inset: 0;
          background: rgba(255, 255, 255, 0.95);
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 16px;

          p {
            font-size: 22px;
            color: $text-color-secondary;
            margin: 0;
          }
        }
      }
    }
  }

  .empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;

    :deep(.el-empty__description p) {
      font-size: 26px !important;
    }

    :deep(.el-empty__image) {
      width: 200px;
      height: 200px;
    }
  }

  .preview-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 500px;

    .preview-full {
      max-width: 100%;
      max-height: 70vh;
      border-radius: $border-radius-md;
    }

    .doc-preview {
      text-align: center;
      padding: 48px;

      .doc-preview-icon {
        font-size: 120px;
        margin-bottom: 24px;
      }

      .doc-preview-name {
        font-size: 32px;
        font-weight: 700;
        color: $text-color-primary;
        margin-bottom: 12px;
      }

      .doc-preview-tip {
        font-size: 24px;
        color: $text-color-secondary;
        margin-bottom: 32px;
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

  .footer-right {
    display: flex;
    gap: 24px;
    align-items: center;
  }
}
</style>
