<template>
  <div class="login-container">
    <a-card class="login-card">
      <a-space :size="20" direction="vertical" fill>
        <div class="title">
          <h1>考试客户端系统</h1>
        </div>

        <a-form :model="form" layout="vertical" @submit="login">
          <a-form-item :rules="[{ required: true, message: '请输入学号或准考证号' }]" field="username"
                       label="学号/准考证号"
          >
            <a-input v-model="form.username" allow-clear placeholder="请输入学号或准考证号"/>
          </a-form-item>

          <a-form-item :rules="[{ required: true, message: '请输入密码' }]" field="password"
                       label="密码"
          >
            <a-input-password v-model="form.password" allow-clear placeholder="请输入密码"/>
          </a-form-item>

          <a-form-item>
            <a-button :loading="isLoading" html-type="submit" long type="primary">
              {{ isLoading ? '登录中...' : '登录' }}
            </a-button>
          </a-form-item>
        </a-form>

        <a-alert v-if="systemInfo" class="system-info" type="info">
          <template #icon>
            <icon-computer/>
          </template>
          <template #title>设备信息</template>
          <div class="device-info">
            <div class="info-item" v-if="deviceInfo.platform">
              <span class="info-label">系统:</span>
              <span class="info-value">{{ deviceInfo.platform }}</span>
            </div>
            <div class="info-item" v-if="deviceInfo.hostname">
              <span class="info-label">主机名:</span>
              <span class="info-value">{{ deviceInfo.hostname }}</span>
            </div>
            <div class="info-item" v-if="deviceInfo.kernelArch">
              <span class="info-label">架构:</span>
              <span class="info-value">{{ deviceInfo.kernelArch }}</span>
            </div>
            <div class="info-item" v-if="deviceInfo.uptime">
              <span class="info-label">运行时间:</span>
              <span class="info-value">{{ formatUptime(deviceInfo.uptime) }}</span>
            </div>
            <div class="info-item" v-if="deviceInfo.procs">
              <span class="info-label">进程数:</span>
              <span class="info-value">{{ deviceInfo.procs }}</span>
            </div>
            <div class="info-item" v-if="deviceInfo.kernelVersion">
              <span class="info-label">内核版本:</span>
              <span class="info-value">{{ deviceInfo.kernelArch }}</span>
            </div>
          </div>
        </a-alert>
      </a-space>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import {onMounted, reactive, ref} from 'vue';
import {Message} from '@arco-design/web-vue';
import ipcService from '../utils/ipc';
import type {IExamInfo} from '../types/ipc';

interface DeviceInfo {
  hostname: string;
  uptime: number;
  bootTime: number;
  procs: number;
  os: string;
  platform: string;
  platformFamily: string;
  platformVersion: string;
  kernelVersion: string;
  kernelArch: string;
  virtualizationSystem: string;
  virtualizationRole: string;
  hostId: string;
}

const emit = defineEmits<{
  (e: 'login-success', examInfo: IExamInfo): void
}>();

const form = reactive({
  username: '5120214350',
  password: '123456'
});
const systemInfo = ref<string>('系统信息加载中...');
const deviceInfo = ref<DeviceInfo>({} as DeviceInfo);
const isLoading = ref(false);

// 格式化运行时间
const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  const parts = [];
  if (days > 0) parts.push(`${days}天`);
  if (hours > 0) parts.push(`${hours}小时`);
  if (minutes > 0) parts.push(`${minutes}分钟`);

  return parts.join(' ');
};

// 监听系统信息
onMounted(() => {
  ipcService.emit('load')
  ipcService.on('systemInfo', (info) => {
    console.log('systemInfo', info);
    try {
      // 解析JSON数据
      deviceInfo.value = JSON.parse(info);
      systemInfo.value = info; // 保留原始信息
    } catch (e) {
      // 如果解析失败，直接显示原始信息
      systemInfo.value = info;
    }
  });

  // 监听登录结果
  ipcService.on('loginResult', (success, examInfo, errorMsg) => {
    console.log('loginResult', success, examInfo, errorMsg);
    isLoading.value = false;
    if (success) {
      Message.success('登录成功');
      emit('login-success', examInfo);
    } else {
      // 显示具体的错误消息
      Message.error(errorMsg || '登录失败，请检查用户名和密码');
    }
  });
});

// 登录处理
const login = () => {
  if (!form.username || !form.password) {
    Message.warning('请输入学号/准考证号和密码');
    return;
  }

  isLoading.value = true;
  ipcService.emit('login', form.username, form.password);
};
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  padding: min(20px, 3vw);
  overflow-y: auto;
}

.login-card {
  width: min(90%, 500px);
  margin: auto;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  transition: all 0.3s ease;
  max-height: min(90vh, 800px);
  overflow-y: auto;
}

.title {
  text-align: center;
  margin-bottom: min(20px, 3vh);
}

.title h1 {
  color: var(--color-text-1);
  font-size: clamp(18px, 3vw, 26px);
  font-weight: 500;
  line-height: 1.3;
}

.system-info {
  width: 100%;
  word-break: break-word;
  white-space: normal;
}

.device-info {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.info-item {
  display: flex;
  flex-wrap: wrap;
  margin-bottom: 4px;
}

.info-label {
  flex: 0 0 min(80px, 30%);
  font-weight: 500;
  color: var(--color-text-2);
  font-size: clamp(13px, 1.5vw, 14px);
}

.info-value {
  flex: 1;
  word-break: break-word;
  font-size: clamp(13px, 1.5vw, 14px);
}

:deep(.arco-alert-content-message) {
  white-space: normal;
  word-break: break-word;
}

:deep(.arco-alert) {
  font-size: clamp(13px, 1.5vw, 14px);
}

:deep(.arco-form-item-label) {
  font-size: clamp(14px, 1.6vw, 16px);
}

:deep(.arco-form-item) {
  margin-bottom: min(24px, 3vh);
}

:deep(.arco-input),
:deep(.arco-input-password) {
  height: min(36px, 5vh);
  font-size: clamp(14px, 1.6vw, 16px);
}

:deep(.arco-btn) {
  height: min(38px, 5vh);
  font-size: clamp(14px, 1.6vw, 16px);
}

/* 响应式调整 */
@media screen and (max-width: 768px) {
  .login-container {
    padding: 16px;
  }
  
  .login-card {
    padding: 16px;
  }
  
  .title h1 {
    font-size: 20px;
  }
  
  :deep(.arco-alert-title) {
    font-size: 15px;
  }
  
  :deep(.arco-form-item-label) {
    font-size: 15px;
    margin-bottom: 4px;
  }
  
  :deep(.arco-alert-icon) {
    font-size: 16px;
  }
}

@media screen and (max-height: 600px) {
  .login-container {
    align-items: flex-start;
    padding-top: 10px;
  }

  .login-card {
    padding: 12px;
    margin-top: 0;
  }

  .title {
    margin-bottom: 12px;
  }

  :deep(.arco-space) {
    gap: 12px !important;
  }
  
  :deep(.arco-form-item) {
    margin-bottom: 16px;
  }
}

/* 超小屏幕 */
@media screen and (max-width: 480px) {
  .login-container {
    padding: 10px;
  }
  
  .login-card {
    width: 95%;
    padding: 12px;
  }
  
  .title h1 {
    font-size: 18px;
  }
  
  :deep(.arco-form-item-label) {
    font-size: 14px;
  }
  
  :deep(.arco-input),
  :deep(.arco-input-password) {
    height: 34px;
    font-size: 14px;
  }
  
  :deep(.arco-btn) {
    height: 36px;
    font-size: 14px;
  }
  
  .info-item {
    margin-bottom: 6px;
  }
  
  .info-label,
  .info-value {
    font-size: 13px;
  }
}

/* 高分辨率屏幕 */
@media screen and (min-width: 1440px) and (min-height: 900px) {
  .login-card {
    width: min(80%, 600px);
    padding: 30px;
  }
  
  .title {
    margin-bottom: 30px;
  }
  
  .title h1 {
    font-size: 28px;
  }
  
  :deep(.arco-form-item) {
    margin-bottom: 28px;
  }
}
</style> 