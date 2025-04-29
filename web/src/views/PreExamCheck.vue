<template>
  <div class="pre-exam-check">
    <a-card class="check-card">
      <a-space direction="vertical" fill :size="20">
        <div class="header">
          <h1>{{ currentStep === 'commitment' ? '考试诚信承诺' : currentStep === 'device-check' ? '设备功能检测' : '检测完成' }}</h1>
          <a-steps :current="stepIndex" size="small">
            <a-step title="诚信承诺" />
            <a-step title="设备检测" />
            <a-step title="完成" />
          </a-steps>
        </div>

        <!-- 诚信承诺步骤 -->
        <div v-if="currentStep === 'commitment'" class="commitment-section">
          <a-typography-title :heading="5" style="text-align: center; margin-bottom: 16px;">考试诚信承诺书</a-typography-title>
          
          <!-- 使用div替代a-scrollbar，并简化结构 -->
          <div class="commitment-content">
            <div class="commitment-text">
              <p>本人承诺：</p>
              <p>1. 独立完成考试，不接受他人帮助，不给他人提供帮助；</p>
              <p>2. 不使用手机、平板等其他电子设备查找答案或与他人交流；</p>
              <p>3. 不查阅未经允许的资料或使用未经允许的工具；</p>
              <p>4. 不在考试过程中离开考试监控范围；</p>
              <p>5. 不使用虚拟机、远程桌面等作弊手段；</p>
              <p>6. 遵守考试规定的时间限制；</p>
              <p>7. 不以任何方式记录、复制或传播考试内容；</p>
              <p>8. 不故意干扰或绕过考试监控系统；</p>
              <p>9. 如被发现有违反考试诚信的行为，愿意接受相应处理。</p>
              <p class="highlight-text">我理解并同意：监考系统将记录我的屏幕、摄像头视频、麦克风音频及系统运行情况，这些信息仅用于确保考试公平公正。</p>
            </div>
          </div>

          <div class="agreement-section">
            <a-checkbox v-model="commitmentAgreed">我已阅读并同意以上承诺内容</a-checkbox>
            <a-button type="primary" :disabled="!commitmentAgreed" @click="moveToDeviceCheck">下一步</a-button>
          </div>
        </div>

        <!-- 设备检测步骤 -->
        <div v-if="currentStep === 'device-check'" class="device-check-section">
          <a-list :max-height="500" :bordered="false">
            <template #header>
              <a-typography-title :heading="6">请完成以下检测项目</a-typography-title>
            </template>
            <a-list-item v-for="(item, index) in checkItems" :key="index">
              <a-card class="check-item-card">
                <div class="check-header">
                  <div class="check-title">
                    <a-badge :status="getBadgeStatus(item.status)" />
                    <a-typography-title :heading="6" style="margin: 0;">{{ item.title }}</a-typography-title>
                  </div>
                  <a-tag :color="getTagColor(item.status)">{{ getStatusText(item.status) }}</a-tag>
                </div>
                <a-typography-paragraph class="check-description">{{ item.description }}</a-typography-paragraph>
                <div class="check-action">
                  <a-button 
                    :type="item.status === 'none' ? 'primary' : item.status === 'success' ? 'outline' : 'primary'" 
                    :status="item.status === 'failed' ? 'danger' : undefined"
                    :loading="item.status === 'loading'"
                    @click="runCheck(index)"
                  >
                    {{ item.status === 'none' ? '开始检测' : item.status === 'loading' ? '检测中...' : '重新检测' }}
                  </a-button>
                </div>
                <a-alert v-if="item.status === 'failed'" type="error" :content="item.errorMessage" />
              </a-card>
            </a-list-item>
          </a-list>

          <div class="navigation-buttons">
            <a-space>
              <a-button @click="currentStep = 'commitment'">上一步</a-button>
              <a-button 
                type="primary" 
                :disabled="!allChecksCompleted || hasFailedChecks" 
                @click="completeChecks"
              >
                完成检测
              </a-button>
            </a-space>
          </div>
        </div>

        <!-- 完成步骤 -->
        <div v-if="currentStep === 'complete'" class="complete-section">
          <a-result status="success" title="设备检测完成" sub-title="您的设备已通过所有必要检测，现在可以进入考试">
            <template #extra>
              <a-button type="primary" @click="enterExam">进入考试</a-button>
            </template>
          </a-result>
        </div>
      </a-space>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import ipcService from '../utils/ipc';
import type { IExamInfo } from '../types/ipc';

interface CheckItem {
  title: string;
  description: string;
  status: 'none' | 'loading' | 'success' | 'failed';
  errorMessage?: string;
  check: () => Promise<boolean>;
}

const props = defineProps<{
  examInfo: IExamInfo
}>();

const emit = defineEmits<{
  (e: 'check-complete'): void
}>();

const currentStep = ref<'commitment' | 'device-check' | 'complete'>('commitment');
const commitmentAgreed = ref(false);

// 获取当前步骤索引
const stepIndex = computed(() => {
  switch (currentStep.value) {
    case 'commitment': return 0;
    case 'device-check': return 1;
    case 'complete': return 2;
    default: return 0;
  }
});

// 设备检测项
const checkItems = ref<CheckItem[]>([
  {
    title: '系统信息检测',
    description: '检测系统是否能正常获取硬件信息',
    status: 'none',
    check: async () => {
      const result = await ipcService.checkSystemInfoAsync();
      if (!result.success && result.message) {
        checkItems.value[0].errorMessage = result.message;
      }
      return result.success;
    }
  },
  {
    title: '截屏功能检测',
    description: '检测系统是否能正常进行屏幕截图',
    status: 'none',
    check: async () => {
      const result = await ipcService.checkScreenshotAsync();
      if (!result.success && result.message) {
        checkItems.value[1].errorMessage = result.message;
      }
      return result.success;
    }
  },
  {
    title: '浏览器功能检测',
    description: '检测是否能正常打开浏览器',
    status: 'none',
    check: async () => {
      const result = await ipcService.checkBrowserAsync();
      if (!result.success && result.message) {
        checkItems.value[2].errorMessage = result.message;
      }
      return result.success;
    }
  },
]);

// 计算所有检测是否完成
const allChecksCompleted = computed(() => {
  return true;
  // return checkItems.value.every(item =>
  //   item.status === 'success' || item.status === 'failed'
  // );
});

// 计算是否有失败的检测
const hasFailedChecks = computed(() => {
  return checkItems.value.some(item => item.status === 'failed');
});

// 获取检测状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'none': return '未检测';
    case 'loading': return '检测中...';
    case 'success': return '通过';
    case 'failed': return '未通过';
    default: return '';
  }
};

// 获取状态对应的Badge状态
const getBadgeStatus = (status: string) => {
  switch (status) {
    case 'none': return 'default';
    case 'loading': return 'processing';
    case 'success': return 'success';
    case 'failed': return 'error';
    default: return 'default';
  }
};

// 获取状态对应的Tag颜色
const getTagColor = (status: string) => {
  switch (status) {
    case 'none': return 'gray';
    case 'loading': return 'blue';
    case 'success': return 'green';
    case 'failed': return 'red';
    default: return '';
  }
};

// 运行特定的检测
const runCheck = async (index: number) => {
  const item = checkItems.value[index];
  item.status = 'loading';
  
  try {
    const result = await item.check();
    
    if (result) {
      item.status = 'success';
    } else {
      item.status = 'failed';
      if (!item.errorMessage) {
        item.errorMessage = '检测未通过，请检查设备后重试';
      }
    }
  } catch (error) {
    item.status = 'failed';
    item.errorMessage = `检测过程出错: ${error}`;
  }
};

// 从承诺页面移动到设备检测页面
const moveToDeviceCheck = () => {
  currentStep.value = 'device-check';
};

// 完成所有检测
const completeChecks = () => {
  currentStep.value = 'complete';
};

// 进入考试
const enterExam = () => {
  emit('check-complete');
};
</script>

<style scoped>
.pre-exam-check {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
  overflow-y: auto;
}

.check-card {
  width: min(90%, 800px);
  margin: auto;
  max-height: 90vh;
  overflow-y: auto;
}

.header {
  text-align: center;
  margin-bottom: 16px;
}

.header h1 {
  margin-bottom: 20px;
  font-size: clamp(20px, 4vw, 24px);
  color: var(--color-text-1);
}

/* 诚信承诺样式 */
.commitment-section {
  display: flex;
  flex-direction: column;
}

.commitment-content {
  height: 350px;
  margin-bottom: 20px;
  border: 1px solid var(--color-border-2);
  border-radius: 4px;
  background-color: var(--color-fill-2);
  overflow-y: auto; /* 关键：启用垂直滚动 */
}

.commitment-text {
  padding: 16px;
  line-height: 1.6;
}

.commitment-text p {
  margin-bottom: 12px;
}

.highlight-text {
  color: var(--color-danger-6);
  font-weight: bold;
}

.agreement-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
}

/* 设备检测样式 */
.device-check-section {
  display: flex;
  flex-direction: column;
}

.check-item-card {
  margin-bottom: 0;
  background-color: var(--color-bg-2);
}

.check-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.check-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.check-description {
  color: var(--color-text-3);
  margin-bottom: 16px;
}

.check-action {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 8px;
}

/* 导航按钮 */
.navigation-buttons {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
}

/* 完成界面 */
.complete-section {
  padding: 20px 0;
}

/* 响应式调整 */
@media screen and (max-width: 768px) {
  .pre-exam-check {
    padding: 12px;
  }
  
  .header h1 {
    font-size: 20px;
    margin-bottom: 16px;
  }
  
  .agreement-section {
    gap: 12px;
  }
  
  .commitment-content {
    height: 300px;
  }
}

@media screen and (max-height: 700px) {
  .pre-exam-check {
    align-items: flex-start;
    padding-top: 10px;
  }
  
  .check-card {
    margin: 0 auto;
  }
  
  .commitment-content {
    height: 250px;
  }
  
  .commitment-text {
    padding: 12px;
  }
}
</style> 