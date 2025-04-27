<template>
  <div class="app">
    <LoginView v-if="!isLoggedIn" @login-success="handleLoginSuccess"/>
    <ExamView v-else :examInfo="examInfo"/>
  </div>
</template>

<script lang="ts">
import {defineComponent, ref} from 'vue';
import LoginView from './views/LoginView.vue';
import ExamView from './views/ExamView.vue';
import type {IExamInfo} from './types/ipc';

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

<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

html, body {
  width: 100%;
  height: 100%;
  overflow: hidden; /* 防止滚动条出现 */
}

body {
  font-family: "Microsoft YaHei", sans-serif;
  background-color: var(--color-fill-2);
  color: var(--color-text-1);
  line-height: 1.6;
}

.app {
  width: 100%;
  height: 100vh;
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
}

/* 在小屏幕上调整显示 */
@media screen and (max-width: 768px) {
  .app {
    padding: 0 12px;
  }
}
</style>
