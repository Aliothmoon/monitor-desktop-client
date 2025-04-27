<template>
  <div class="login-container">
    <h1>考试客户端系统</h1>
    <div class="form-group">
      <label for="username">学号/准考证号:</label>
      <input 
        type="text" 
        id="username" 
        v-model="username" 
        placeholder="请输入学号或准考证号"
      >
    </div>
    <div class="form-group">
      <label for="password">密码:</label>
      <input 
        type="password" 
        id="password" 
        v-model="password" 
        placeholder="请输入密码"
      >
    </div>
    <button 
      id="loginBtn" 
      @click="login" 
      :disabled="isLoading"
    >{{ isLoading ? '正在登录...' : '登录' }}</button>
    <div v-show="isLoading" class="loading">正在登录，请稍候...</div>
    <div class="system-info">
      <div id="systemInfo">{{ systemInfo }}</div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue';
import ipcService from '../utils/ipc';

export default defineComponent({
  name: 'LoginView',
  emits: ['login-success'],
  setup(_, { emit }) {
    const username = ref('');
    const password = ref('');
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
          emit('login-success', examInfo);
        } else {
          alert('登录失败，请检查用户名和密码');
        }
      });
    });

    // 登录处理
    const login = () => {
      if (!username.value || !password.value) {
        alert('请输入学号/准考证号和密码');
        return;
      }
      
      isLoading.value = true;
      ipcService.emit('login', username.value, password.value);
    };

    return {
      username,
      password,
      systemInfo,
      isLoading,
      login
    };
  }
});
</script>

<style scoped>
.login-container {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 30px;
  margin-top: 50px;
  max-width: 500px;
  margin-left: auto;
  margin-right: auto;
}

h1 {
  text-align: center;
  margin-bottom: 20px;
  color: #4A90E2;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

input[type="text"], input[type="password"] {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 16px;
}

button {
  background-color: #4A90E2;
  color: white;
  border: none;
  padding: 12px 20px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  width: 100%;
  transition: background-color 0.3s;
}

button:hover:not(:disabled) {
  background-color: #3A7BC8;
}

button:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.loading {
  text-align: center;
  margin-top: 10px;
  color: #4A90E2;
}

.system-info {
  margin-top: 20px;
  color: #666;
  font-size: 14px;
}
</style> 