<script lang="ts">
import { defineComponent, ref } from 'vue';
import LoginView from './views/LoginView.vue';
import ExamView from './views/ExamView.vue';
import type { IExamInfo } from './types/ipc';

export default defineComponent({
  name: 'App',
  components: {
    LoginView,
    ExamView
  },
  setup() {
    const isLoggedIn = ref(false);
    const examInfo = ref<IExamInfo>({
      examId: '',
      title: '',
      startTime: '',
      endTime: '',
      duration: 0,
      studentName: ''
    });

    const handleLoginSuccess = (info: IExamInfo) => {
      examInfo.value = info;
      isLoggedIn.value = true;
    };

    return {
      isLoggedIn,
      examInfo,
      handleLoginSuccess
    };
  }
});
</script>

<template>
  <div class="app">
    <div class="container">
      <LoginView v-if="!isLoggedIn" @login-success="handleLoginSuccess" />
      <ExamView v-else :examInfo="examInfo" />
    </div>
  </div>
</template>

<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: "Microsoft YaHei", sans-serif;
  background-color: #f5f5f5;
  color: #333;
  line-height: 1.6;
}

.container {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

.app {
  min-height: 100vh;
}
</style>
