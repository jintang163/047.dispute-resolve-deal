<template>
  <div class="dispute-type-page">
    <StepIndicator />

    <div class="content">
      <h2 class="page-title">请选择纠纷类型</h2>

      <div class="breadcrumb-bar">
        <div
          v-for="(node, index) in selectedPath"
          :key="index"
          class="breadcrumb-item"
          :class="{ clickable: index < selectedPath.length - 1, active: index === selectedPath.length - 1 }"
          @click="handleBreadcrumbClick(index)"
        >
          <span v-if="index === 0">🏠 全部类型</span>
          <span v-else>{{ node.name }}</span>
          <span v-if="index < selectedPath.length - 1" class="separator">›</span>
        </div>
        <div v-if="selectedPath.length === 0" class="breadcrumb-item active">
          🏠 全部类型
        </div>
      </div>

      <div class="category-grid" v-if="currentCategories.length > 0">
        <div
          v-for="category in currentCategories"
          :key="category.id"
          class="category-card"
          :class="{ selected: isSelected(category) }"
          @click="handleCategoryClick(category)"
        >
          <div class="category-icon">{{ getCategoryIcon(category.name) }}</div>
          <div class="category-info">
            <div class="category-name">{{ category.name }}</div>
            <div class="category-desc" v-if="category.children">
              包含 {{ category.children.length }} 个子分类
            </div>
            <div class="category-desc" v-else>
              点击选择此类型
            </div>
          </div>
          <div class="category-arrow" v-if="category.children">→</div>
          <div class="category-check" v-else>✓</div>
        </div>
      </div>

      <el-card v-if="selectedPath.length > 0 && !currentParent?.children?.length" class="selected-card" shadow="hover">
        <template #header>
          <div class="selected-header">
            <span class="selected-title">已选择的纠纷类型</span>
            <el-tag type="success" size="large">已确认</el-tag>
          </div>
        </template>
        <div class="selected-content">
          <div class="selected-path">
            <span
              v-for="(node, index) in selectedPath"
              :key="index"
              class="path-tag"
            >
              {{ node.name }}
            </span>
          </div>
          <div class="selected-description">
            <p>您选择了：<strong>{{ finalTypeName }}</strong></p>
            <p class="tip">请确认类型选择正确，点击下一步继续填写信息。如需修改，可点击上方分类重新选择。</p>
          </div>
        </div>
      </el-card>
    </div>

    <div class="footer">
      <TouchButton icon="ArrowLeft" size="large" @click="goBack">上一步</TouchButton>
      <TouchButton
        type="primary"
        icon="ArrowRight"
        size="xl"
        :disabled="!canProceed"
        @click="handleNext"
      >
        下一步：填写纠纷信息
      </TouchButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import StepIndicator from '@/components/StepIndicator.vue'
import TouchButton from '@/components/TouchButton.vue'
import { useKioskStore, type DisputeTypeNode } from '@/stores/kiosk'
import kioskApi from '@/services/kiosk'

const router = useRouter()
const store = useKioskStore()

const allTypes = ref<DisputeTypeNode[]>([])
const selectedPath = ref<DisputeTypeNode[]>([...store.caseDraft.disputeTypePath])

const mockTypes: DisputeTypeNode[] = [
  {
    id: '1',
    name: '劳动争议',
    children: [
      {
        id: '1-1',
        name: '劳动合同',
        children: [
          { id: '1-1-1', name: '合同签订纠纷' },
          { id: '1-1-2', name: '合同解除纠纷' },
          { id: '1-1-3', name: '合同续签纠纷' },
          { id: '1-1-4', name: '试用期纠纷' }
        ]
      },
      {
        id: '1-2',
        name: '工资报酬',
        children: [
          { id: '1-2-1', name: '拖欠工资' },
          { id: '1-2-2', name: '加班费争议' },
          { id: '1-2-3', name: '奖金提成纠纷' },
          { id: '1-2-4', name: '社保公积金争议' }
        ]
      },
      {
        id: '1-3',
        name: '工伤赔偿',
        children: [
          { id: '1-3-1', name: '工伤认定争议' },
          { id: '1-3-2', name: '工伤赔偿纠纷' }
        ]
      }
    ]
  },
  {
    id: '2',
    name: '婚姻家庭',
    children: [
      {
        id: '2-1',
        name: '婚姻关系',
        children: [
          { id: '2-1-1', name: '离婚纠纷' },
          { id: '2-1-2', name: '财产分割' },
          { id: '2-1-3', name: '彩礼返还' }
        ]
      },
      {
        id: '2-2',
        name: '子女抚养',
        children: [
          { id: '2-2-1', name: '抚养权争议' },
          { id: '2-2-2', name: '抚养费纠纷' },
          { id: '2-2-3', name: '探视权纠纷' }
        ]
      },
      {
        id: '2-3',
        name: '赡养继承',
        children: [
          { id: '2-3-1', name: '赡养纠纷' },
          { id: '2-3-2', name: '遗产继承' }
        ]
      }
    ]
  },
  {
    id: '3',
    name: '物业邻里',
    children: [
      {
        id: '3-1',
        name: '物业服务',
        children: [
          { id: '3-1-1', name: '物业费纠纷' },
          { id: '3-1-2', name: '服务质量争议' },
          { id: '3-1-3', name: '维修基金使用' }
        ]
      },
      {
        id: '3-2',
        name: '邻里纠纷',
        children: [
          { id: '3-2-1', name: '噪音扰民' },
          { id: '3-2-2', name: '违章搭建' },
          { id: '3-2-3', name: '采光通风纠纷' },
          { id: '3-2-4', name: '装修损害' }
        ]
      }
    ]
  },
  {
    id: '4',
    name: '消费维权',
    children: [
      {
        id: '4-1',
        name: '商品质量',
        children: [
          { id: '4-1-1', name: '假冒伪劣' },
          { id: '4-1-2', name: '虚假宣传' },
          { id: '4-1-3', name: '退换货纠纷' }
        ]
      },
      {
        id: '4-2',
        name: '服务消费',
        children: [
          { id: '4-2-1', name: '餐饮服务' },
          { id: '4-2-2', name: '美容美发' },
          { id: '4-2-3', name: '教育培训' },
          { id: '4-2-4', name: '健身服务' }
        ]
      }
    ]
  },
  {
    id: '5',
    name: '民间借贷',
    children: [
      {
        id: '5-1',
        name: '借款纠纷',
        children: [
          { id: '5-1-1', name: '欠款不还' },
          { id: '5-1-2', name: '利息争议' },
          { id: '5-1-3', name: '担保纠纷' }
        ]
      }
    ]
  },
  {
    id: '6',
    name: '交通事故',
    children: [
      {
        id: '6-1',
        name: '责任认定',
        children: [
          { id: '6-1-1', name: '责任划分争议' }
        ]
      },
      {
        id: '6-2',
        name: '赔偿纠纷',
        children: [
          { id: '6-2-1', name: '医疗费赔偿' },
          { id: '6-2-2', name: '车辆维修赔偿' },
          { id: '6-2-3', name: '误工费赔偿' }
        ]
      }
    ]
  }
]

const currentParent = computed(() => {
  if (selectedPath.value.length === 0) return null
  return selectedPath.value[selectedPath.value.length - 1]
})

const currentCategories = computed(() => {
  if (selectedPath.value.length === 0) {
    return allTypes.value
  }
  const parent = selectedPath.value[selectedPath.value.length - 1]
  return parent.children || []
})

const canProceed = computed(() => {
  return selectedPath.value.length > 0 && !currentParent.value?.children
})

const finalTypeName = computed(() => {
  return selectedPath.value.map(n => n.name).join(' - ')
})

function getCategoryIcon(name: string): string {
  const icons: Record<string, string> = {
    '劳动争议': '💼',
    '劳动合同': '📋',
    '工资报酬': '💰',
    '工伤赔偿': '🏥',
    '婚姻家庭': '👨‍👩‍👧',
    '婚姻关系': '💍',
    '子女抚养': '👶',
    '赡养继承': '👴',
    '物业邻里': '🏢',
    '物业服务': '🔧',
    '邻里纠纷': '🏘️',
    '消费维权': '🛒',
    '商品质量': '📦',
    '服务消费': '☕',
    '民间借贷': '💳',
    '借款纠纷': '📊',
    '交通事故': '🚗',
    '责任认定': '📝',
    '赔偿纠纷': '💸'
  }
  return icons[name] || '📁'
}

function isSelected(category: DisputeTypeNode): boolean {
  return selectedPath.value.some(n => n.id === category.id)
}

function handleCategoryClick(category: DisputeTypeNode) {
  if (category.children && category.children.length > 0) {
    const idx = selectedPath.value.findIndex(n => n.id === category.id)
    if (idx >= 0) {
      selectedPath.value = selectedPath.value.slice(0, idx + 1)
    } else {
      selectedPath.value.push(category)
    }
  } else {
    const parentIdx = selectedPath.value.length - 1
    if (parentIdx >= 0 && selectedPath.value[parentIdx].children) {
      selectedPath.value = [...selectedPath.value.slice(0, parentIdx + 1), category]
    } else {
      selectedPath.value.push(category)
    }
    store.setDisputeTypePath(selectedPath.value)
    ElMessage({
      message: `已选择：${category.name}`,
      type: 'success',
      duration: 2000
    })
  }
}

function handleBreadcrumbClick(index: number) {
  selectedPath.value = selectedPath.value.slice(0, index)
  if (selectedPath.value.length > 0) {
    store.setDisputeTypePath(selectedPath.value)
  }
}

async function loadTypes() {
  try {
    const data = await kioskApi.getDisputeTypes()
    if (data && data.length > 0) {
      allTypes.value = data
    } else {
      allTypes.value = mockTypes
    }
  } catch {
    allTypes.value = mockTypes
  }
}

function goBack() {
  router.push('/idcard')
}

function handleNext() {
  if (!canProceed.value) {
    ElMessage.warning('请选择具体的纠纷类型')
    return
  }
  store.setDisputeTypePath(selectedPath.value)
  router.push('/form')
}

onMounted(() => {
  loadTypes()
})
</script>

<style lang="scss" scoped>
.dispute-type-page {
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
    margin: 0 0 32px;
    text-align: center;
  }

  .breadcrumb-bar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
    padding: 20px 32px;
    background: $bg-card;
    border-radius: $border-radius-lg;
    margin-bottom: 32px;
    box-shadow: $shadow-card;

    .breadcrumb-item {
      display: flex;
      align-items: center;
      font-size: 26px;
      color: $text-color-secondary;
      padding: 8px 16px;
      border-radius: $border-radius-sm;

      &.clickable {
        cursor: pointer;
        color: $primary-color;

        &:hover {
          background: rgba(29, 108, 255, 0.1);
        }
      }

      &.active {
        color: $text-color-primary;
        font-weight: 700;
      }

      .separator {
        margin-left: 8px;
        color: $text-color-light;
        font-weight: 300;
      }
    }
  }

  .category-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 24px;
    overflow-y: auto;
    padding: 8px;
    flex: 1;

    .category-card {
      display: flex;
      align-items: center;
      gap: 24px;
      padding: 36px 32px;
      background: $bg-card;
      border-radius: $border-radius-lg;
      cursor: pointer;
      border: 3px solid transparent;
      box-shadow: $shadow-card;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-4px);
        box-shadow: $shadow-hover;
        border-color: rgba(29, 108, 255, 0.3);
      }

      &.selected {
        background: linear-gradient(135deg, rgba(29, 108, 255, 0.1) 0%, rgba(77, 140, 255, 0.1) 100%);
        border-color: $primary-color;
      }

      .category-icon {
        font-size: 64px;
        flex-shrink: 0;
      }

      .category-info {
        flex: 1;
        min-width: 0;

        .category-name {
          font-size: 30px;
          font-weight: 700;
          color: $text-color-primary;
          margin-bottom: 8px;
        }

        .category-desc {
          font-size: 22px;
          color: $text-color-secondary;
        }
      }

      .category-arrow {
        font-size: 40px;
        color: $primary-color;
        font-weight: 300;
      }

      .category-check {
        width: 56px;
        height: 56px;
        border-radius: 50%;
        background: $success-color;
        color: white;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 32px;
        font-weight: 700;
      }
    }
  }

  .selected-card {
    margin-top: 24px;

    :deep(.el-card__header) {
      padding: 24px 32px;
    }

    :deep(.el-card__body) {
      padding: 32px;
    }

    .selected-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .selected-title {
        font-size: 30px;
        font-weight: 700;
        color: $text-color-primary;
      }
    }

    .selected-content {
      .selected-path {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        margin-bottom: 24px;

        .path-tag {
          display: inline-block;
          padding: 12px 24px;
          background: rgba(29, 108, 255, 0.1);
          color: $primary-color;
          border-radius: $border-radius-md;
          font-size: 24px;
          font-weight: 600;

          &:not(:last-child)::after {
            content: '›';
            margin-left: 24px;
            color: $text-color-light;
          }
        }
      }

      .selected-description {
        p {
          font-size: 28px;
          color: $text-color-primary;
          margin: 0 0 12px;
        }

        .tip {
          font-size: 24px !important;
          color: $text-color-secondary !important;
          margin: 0 !important;
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
