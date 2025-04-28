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
  username: '',
  password: ''
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
  ipcService.on('loginResult', (success, examInfo) => {
    console.log('loginResult', success, examInfo);
    isLoading.value = false;
    if (success) {
      Message.success('登录成功');
      emit('login-success', examInfo);
    } else {
      Message.error('登录失败，请检查用户名和密码');
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
  padding: 20px;
  overflow-y: auto;
}

.login-card {
  width: min(90%, 500px);
  margin: auto;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  transition: all 0.3s ease;
  max-height: 90vh;
  overflow-y: auto;
}

.title {
  text-align: center;
  margin-bottom: 20px;
}

.title h1 {
  color: var(--color-text-1);
  font-size: clamp(20px, 4vw, 26px);
  font-weight: 500;
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
}

.info-label {
  flex: 0 0 80px;
  font-weight: 500;
  color: var(--color-text-2);
}

.info-value {
  flex: 1;
  word-break: break-word;
}

:deep(.arco-alert-content-message) {
  white-space: normal;
  word-break: break-word;
}

/* 响应式调整 */
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
}
</style> 