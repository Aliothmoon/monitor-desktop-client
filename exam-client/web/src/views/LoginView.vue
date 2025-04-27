<template>
  <div class="login-container">
    <a-card class="login-card">
      <a-space :size="24" direction="vertical" fill>
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

        <a-alert v-if="systemInfo" type="info">
          <template #icon>
            <icon-computer/>
          </template>
          <template #message>{{ systemInfo }}</template>
        </a-alert>
      </a-space>
    </a-card>
  </div>
</template>

<script lang="ts">
import {defineComponent, onMounted, reactive, ref} from 'vue';
import {Message} from '@arco-design/web-vue';
import ipcService from '../utils/ipc';

export default defineComponent({
  name: 'LoginView',
  emits: ['login-success'],
  setup(_, {emit}) {
    const form = reactive({
      username: '',
      password: ''
    });
    const systemInfo = ref('系统信息加载中...');
    const isLoading = ref(false);

    // 监听系统信息
    onMounted(() => {
      ipcService.on('systemInfo', (info) => {
        systemInfo.value = info;
      });

      // 监听登录结果
      ipcService.on('loginResult', (success, examInfo) => {
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

    return {
      form,
      systemInfo,
      isLoading,
      login
    };
  }
});
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
}

.login-card {
  width: min(90%, 450px);
  margin: auto;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.3s ease;
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

/* 响应式调整 */
@media screen and (max-height: 600px) {
  .login-card {
    padding: 12px;
  }
  
  .title {
    margin-bottom: 10px;
  }
  
  :deep(.arco-space) {
    gap: 16px !important;
  }
}
</style> 