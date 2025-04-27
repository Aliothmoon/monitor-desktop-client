<template>
  <div class="exam-container">
    <h2>{{ examInfo.title }}</h2>
    <div class="countdown" id="countdown">剩余时间: {{ countdown }}</div>
    <div class="exam-info">
      <p>考试编号: <span>{{ examInfo.examId }}</span></p>
      <p>开始时间: <span>{{ examInfo.startTime }}</span></p>
      <p>结束时间: <span>{{ examInfo.endTime }}</span></p>
      <p>考生姓名: <span>{{ examInfo.studentName }}</span></p>
    </div>

    <div class="tabs">
      <div 
        :class="['tab', { active: activeTab === 'exam' }]" 
        @click="activeTab = 'exam'"
      >考试内容</div>
      <div 
        :class="['tab', { active: activeTab === 'browser' }]" 
        @click="activeTab = 'browser'"
      >浏览器</div>
      <div 
        :class="['tab', { active: activeTab === 'monitor' }]" 
        @click="activeTab = 'monitor'"
      >系统监控</div>
    </div>

    <div class="tab-content">
      <!-- 考试内容 -->
      <div v-if="activeTab === 'exam'" class="tab-pane">
        <div class="content-area">
          <h3>考试说明</h3>
          <p>这里是考试说明和内容。实际应用中这里可以根据需要加载考试题目。</p>
        </div>
      </div>

      <!-- 浏览器 -->
      <div v-if="activeTab === 'browser'" class="tab-pane">
        <div class="browser-container">
          <div class="url-input">
            <input 
              type="text" 
              v-model="browserUrl" 
              placeholder="输入网址"
              @keyup.enter="openBrowser"
            />
            <button @click="openBrowser">打开浏览器</button>
          </div>
          <div class="https-info">
            <h3>浏览器访问记录</h3>
            <div class="https-visits">
              <div v-if="browserVisits.length === 0">无访问记录</div>
              <div v-for="(visit, index) in browserVisits" :key="index">
                {{ visit.time }} - {{ visit.url }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 系统监控 -->
      <div v-if="activeTab === 'monitor'" class="tab-pane">
        <div class="process-info">
          <h3>系统进程信息</h3>
          <table class="process-table">
            <thead>
              <tr>
                <th>PID</th>
                <th>进程名</th>
                <th>内存使用(MB)</th>
                <th>CPU使用(%)</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="processes.length === 0">
                <td colspan="4" style="text-align: center;">加载中...</td>
              </tr>
              <tr v-for="proc in processes" :key="proc.pid">
                <td>{{ proc.pid }}</td>
                <td>{{ proc.name }}</td>
                <td>{{ proc.memory }}</td>
                <td>{{ proc.cpu.toFixed(2) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, onUnmounted } from 'vue';
import type { PropType } from 'vue';
import ipcService from '../utils/ipc';
import type { IExamInfo, IProcessInfo } from '../types/ipc';

interface BrowserVisit {
  time: string;
  url: string;
}

export default defineComponent({
  name: 'ExamView',
  props: {
    examInfo: {
      type: Object as PropType<IExamInfo>,
      required: true
    }
  },
  setup(props) {
    const activeTab = ref('exam');
    const browserUrl = ref('');
    const browserVisits = ref<BrowserVisit[]>([]);
    const processes = ref<IProcessInfo[]>([]);
    const visitedUrls = new Set<string>();
    
    // 计算剩余时间
    const calculateRemainingTime = () => {
      const now = new Date();
      const endTime = new Date(props.examInfo.endTime);
      
      if (now >= endTime) {
        return '00:00:00';
      }
      
      const diff = Math.floor((endTime.getTime() - now.getTime()) / 1000);
      const hours = Math.floor(diff / 3600).toString().padStart(2, '0');
      const minutes = Math.floor((diff % 3600) / 60).toString().padStart(2, '0');
      const seconds = Math.floor(diff % 60).toString().padStart(2, '0');
      
      return `${hours}:${minutes}:${seconds}`;
    };
    
    // 倒计时
    const countdown = ref(calculateRemainingTime());
    let timerId: number;
    
    onMounted(() => {
      // 设置倒计时
      timerId = window.setInterval(() => {
        countdown.value = calculateRemainingTime();
      }, 1000);
      
      // 监听浏览器访问
      ipcService.on('browserVisit', (url) => {
        if (!visitedUrls.has(url)) {
          visitedUrls.add(url);
          const time = new Date().toLocaleTimeString();
          browserVisits.value.unshift({ time, url });
        }
      });
      
      // 监听进程信息
      ipcService.on('processInfo', (processList) => {
        processes.value = processList;
      });
    });
    
    // 组件销毁时清除定时器
    onUnmounted(() => {
      clearInterval(timerId);
    });
    
    // 打开浏览器
    const openBrowser = () => {
      let url = browserUrl.value;
      if (!url) return;
      
      // 确保URL格式正确
      if (!url.startsWith('http://') && !url.startsWith('https://')) {
        url = 'https://' + url;
      }
      
      ipcService.emit('openBrowser', url);
    };
    
    return {
      activeTab,
      browserUrl,
      browserVisits,
      processes,
      countdown,
      openBrowser
    };
  }
});
</script>

<style scoped>
.exam-container {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  padding: 30px;
  margin-top: 20px;
  max-width: 1000px;
  margin-left: auto;
  margin-right: auto;
}

h2 {
  text-align: center;
  margin-bottom: 20px;
  color: #4A90E2;
}

.countdown {
  font-size: 24px;
  font-weight: bold;
  color: #e74c3c;
  text-align: center;
  margin: 20px 0;
}

.exam-info {
  margin-bottom: 20px;
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
}

.exam-info p {
  margin: 5px 0;
  flex-basis: calc(50% - 10px);
}

.tabs {
  display: flex;
  border-bottom: 1px solid #ddd;
  margin-bottom: 20px;
}

.tab {
  padding: 10px 20px;
  cursor: pointer;
  margin-right: 5px;
  border-radius: 4px 4px 0 0;
}

.tab.active {
  background-color: #4A90E2;
  color: white;
}

.tab-content {
  min-height: 300px;
}

.tab-pane {
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 4px;
}

.process-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
  font-size: 14px;
}

.process-table th, .process-table td {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
}

.process-table th {
  background-color: #f2f2f2;
}

.url-input {
  display: flex;
  margin-bottom: 20px;
}

.url-input input {
  flex: 1;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px 0 0 4px;
  font-size: 16px;
}

.url-input button {
  background-color: #4A90E2;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 0 4px 4px 0;
  cursor: pointer;
  font-size: 16px;
}

.https-visits {
  max-height: 150px;
  overflow-y: auto;
  margin-top: 10px;
  font-size: 14px;
  border: 1px solid #ddd;
  padding: 10px;
  border-radius: 4px;
}

.content-area {
  min-height: 300px;
  padding: 20px;
  background-color: white;
  border-radius: 4px;
  box-shadow: 0 0 5px rgba(0, 0, 0, 0.05);
}
</style> 