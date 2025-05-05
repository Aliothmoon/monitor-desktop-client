<template>
  <div class="exam-container">
    <a-card class="exam-card">
      <div class="card-title">
        <h2 class="exam-title">{{ examInfo.title }}</h2>
        <a-typography-text class="countdown" type="danger">
          <icon-clock-circle style="margin-right: 6px;"/>
          剩余时间: {{ countdown }}
        </a-typography-text>
      </div>

      <!-- 退出登录按钮 -->
      <div class="logout-btn">
        <a-button status="danger" size="small" @click="logout">
          <template #icon>
            <icon-export/>
          </template>
          退出登录
        </a-button>
      </div>

      <a-descriptions :data="examInfoData" bordered size="large" title="考试信息"/>

      <!-- 打开浏览器按钮区域 -->
      <div class="browser-actions">
        <div class="browser-input-group">
          <a-button type="primary" @click="openBrowser">
            <template #icon>
              <icon-relation/>
            </template>
            打开浏览器
          </a-button>
        </div>

        <!-- 浏览器访问记录 -->
        <a-collapse class="browser-visits-collapse" :accordion="false">
          <a-collapse-item key="1" header="浏览器访问记录">
            <a-empty v-if="browserVisits.length === 0" description="无访问记录"/>
            <a-list v-else :bordered="false" :max-height="250">
              <a-list-item v-for="(visit, index) in browserVisits" :key="index">
                <a-space>
                  <div>
                    <div>{{ visit.url }}</div>
                    <a-typography-text type="secondary">{{ visit.time }}</a-typography-text>
                  </div>
                </a-space>
              </a-list-item>
            </a-list>
          </a-collapse-item>
        </a-collapse>
      </div>

      <a-tabs v-model:active-key="activeTab" class="exam-tabs">
        <a-tab-pane key="monitor" title="系统监控">
          <a-card title="系统进程信息">
            <a-table :bordered="true" :columns="processColumns" :data="processes" :pagination="false">
              <template #loading>
                <a-empty description="加载中..."/>
              </template>
            </a-table>
          </a-card>
        </a-tab-pane>

        <a-tab-pane key="behavior" title="行为监控">
          <div class="behavior-summary">
            <h3>行为统计</h3>
            <a-row :gutter="16" style="margin-top: 16px">
              <a-col :span="8">
                <a-statistic title="正常行为" :value="normalBehaviorCount" :value-style="{ color: '#3c9' }"/>
              </a-col>
              <a-col :span="8">
                <a-statistic title="可疑行为" :value="suspiciousBehaviorCount" :value-style="{ color: '#f90' }"/>
              </a-col>
              <a-col :span="8">
                <a-statistic title="违规行为" :value="violationBehaviorCount" :value-style="{ color: '#f00' }"/>
              </a-col>
            </a-row>
          </div>
          <a-divider style="margin: 16px 0"/>
          <a-card title="考生行为记录">
            <a-empty v-if="behaviorLogs.length === 0" description="暂无行为记录"/>
            <a-list v-else :bordered="false" :max-height="500">
              <a-list-item v-for="(log, index) in behaviorLogs" :key="index">
                <div class="behavior-log-item">
                  <div class="behavior-time">{{ log.time }}</div>
                  <div class="behavior-type" :class="getBehaviorTypeClass(log.type)">
                    <a-tag :color="getBehaviorTypeColor(log.type)">{{ getBehaviorTypeText(log.type) }}</a-tag>
                  </div>
                  <div class="behavior-content">{{ log.content }}</div>
                </div>
              </a-list-item>
            </a-list>
          </a-card>
        </a-tab-pane>
      </a-tabs>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import {computed, onMounted, onUnmounted, ref} from 'vue';
import {Modal, type TableColumnData} from '@arco-design/web-vue';
import ipcService from '../utils/ipc';
import type {IExamInfo, IProcessInfo} from '../types/ipc';

interface BrowserVisit {
  time: string;
  url: string;
}

interface BehaviorLog {
  time: string;
  type: 'normal' | 'suspicious' | 'violation';
  content: string;
}

const props = defineProps<{
  examInfo: IExamInfo
}>();

// 当前激活的标签页
const activeTab = ref('monitor');

// 浏览器相关数据
const browserVisits = ref<BrowserVisit[]>([]);
const visitedUrls = new Set<string>();

// 进程信息相关数据
const processes = ref<IProcessInfo[]>([]);

// 行为监控数据
const behaviorLogs = ref<BehaviorLog[]>([
  {
    time: new Date().toLocaleTimeString(),
    type: 'normal',
    content: '考生登录系统'
  },
  {
    time: new Date(Date.now() - 60000).toLocaleTimeString(),
    type: 'normal',
    content: '查看考试内容'
  },
  {
    time: new Date(Date.now() - 120000).toLocaleTimeString(),
    type: 'suspicious',
    content: '切换到其他应用程序'
  },
  {
    time: new Date(Date.now() - 180000).toLocaleTimeString(),
    type: 'violation',
    content: '检测到开启手机热点'
  }
]);

// 统计行为数量
const normalBehaviorCount = computed(() =>
    behaviorLogs.value.filter(log => log.type === 'normal').length
);

const suspiciousBehaviorCount = computed(() =>
    behaviorLogs.value.filter(log => log.type === 'suspicious').length
);

const violationBehaviorCount = computed(() =>
    behaviorLogs.value.filter(log => log.type === 'violation').length
);

// 获取行为类型对应的颜色
const getBehaviorTypeColor = (type: string) => {
  switch (type) {
    case 'normal':
      return 'green';
    case 'suspicious':
      return 'orange';
    case 'violation':
      return 'red';
    default:
      return '';
  }
};

// 获取行为类型对应的文本
const getBehaviorTypeText = (type: string) => {
  switch (type) {
    case 'normal':
      return '正常';
    case 'suspicious':
      return '可疑';
    case 'violation':
      return '违规';
    default:
      return '';
  }
};

// 获取行为类型对应的CSS类
const getBehaviorTypeClass = (type: string) => {
  return `behavior-type-${type}`;
};

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

// 考试信息显示数据
const examInfoData = computed(() => [
  {
    label: '考试编号',
    value: props.examInfo.examId,
  },
  {
    label: '开始时间',
    value: props.examInfo.startTime,
  },
  {
    label: '结束时间',
    value: props.examInfo.endTime,
  },
  {
    label: '考生姓名',
    value: props.examInfo.studentName,
  },
]);

// 进程表格列定义
const processColumns = [
  {
    title: 'PID',
    dataIndex: 'pid',
  },
  {
    title: '进程名',
    dataIndex: 'name',
  },
  {
    title: '内存使用(MB)',
    dataIndex: 'memory',
  },
  {
    title: 'CPU使用(%)',
    dataIndex: 'cpu',
    render: ({record}: { record: IProcessInfo }) => record.cpu.toFixed(2),
  },
] as TableColumnData[];

onMounted(() => {
  // 设置倒计时
  timerId = window.setInterval(() => {
    countdown.value = calculateRemainingTime();
  }, 1000);

  // 监听浏览器访问
  ipcService.on('browserVisit', (url) => {
    console.log(url)
    if (url.startsWith('about:blank') && !visitedUrls.has(url)) {
      visitedUrls.add(url);
      const time = new Date().toLocaleTimeString();
      browserVisits.value.unshift({time, url});

      // 同时记录到行为监控中
      behaviorLogs.value.unshift({
        time,
        type: 'normal',
        content: `访问网站: ${url}`
      });
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
  ipcService.emit('openBrowser');
};

// 登出功能
const logout = () => {
  Modal.warning({
    title: '确认退出',
    content: '确定要退出登录吗？系统将关闭当前考试会话。',
    okText: '确定',
    cancelText: '取消',
    onOk: () => {
      // 调用登出方法
      console.log('logout');
      // 调用登出方法
      ipcService.logout();
      // 通知父组件登出事件
      emit('logout');
    }
  });

};

// 定义emit事件
const emit = defineEmits<{
  (e: 'logout'): void
}>();
</script>

<style scoped>
.exam-container {
  width: 100%;
  height: 100vh;
  padding: min(20px, 3vw);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.exam-card {
  flex: 1;
  overflow: auto;
  margin-bottom: min(20px, 3vh);
}

.card-title {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: min(12px, 2vh);
  padding: min(15px, 2vw) min(20px, 3vw);
  width: 100%;
  min-height: min(90px, 15vh);
}

.exam-title {
  margin: 0;
  font-size: clamp(18px, 3vw, 24px);
  color: var(--color-text-1);
  font-weight: 500;
  text-align: center;
  width: 100%;
  overflow-wrap: break-word;
  word-break: break-word;
  line-height: 1.4;
  padding: 0 min(10px, 2vw);
  max-width: 100%;
  hyphens: auto;
}

.countdown {
  font-size: clamp(16px, 2.5vw, 18px);
  font-weight: bold;
  margin-top: min(5px, 1vh);
}

/* 退出登录按钮样式 */
.logout-btn {
  position: absolute;
  top: min(16px, 3vh);
  right: min(16px, 3vw);
  z-index: 10;
}

/* 浏览器操作区域样式 */
.browser-actions {
  margin-top: min(24px, 3vh);
  margin-bottom: min(16px, 2vh);
  padding: min(16px, 3vw);
  background-color: var(--color-fill-1);
  border-radius: 4px;
}

.browser-input-group {
  display: flex;
  gap: 8px;
  margin-bottom: min(16px, 2vh);
}

.browser-visits-collapse {
  margin-top: 8px;
}

.exam-tabs {
  margin-top: min(16px, 2vh);
}

.content-card {
  margin-top: min(16px, 2vh);
}

/* 行为监控样式 */
.behavior-log-item {
  width: 100%;
  display: flex;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 8px;
}

.behavior-time {
  color: var(--color-text-3);
  min-width: clamp(70px, 20%, 80px);
  font-size: clamp(13px, 1.5vw, 14px);
}

.behavior-type {
  min-width: clamp(50px, 15%, 60px);
  font-size: clamp(13px, 1.5vw, 14px);
}

.behavior-content {
  flex: 1;
  word-break: break-word;
  font-size: clamp(13px, 1.5vw, 14px);
}

.behavior-summary {
  margin-top: min(16px, 2vh);
}

:deep(.arco-card-header) {
  min-height: auto;
  height: auto;
  padding: 0;
}

:deep(.arco-card-header-title) {
  overflow: visible;
  white-space: normal;
  text-overflow: clip;
  font-size: clamp(15px, 2vw, 16px);
}

:deep(.arco-descriptions-title) {
  font-size: clamp(15px, 2vw, 16px);
  margin-bottom: min(16px, 2vh);
}

:deep(.arco-descriptions-item-label),
:deep(.arco-descriptions-item-value) {
  font-size: clamp(13px, 1.5vw, 14px);
  padding: min(8px, 1vw) min(12px, 1.5vw);
}

:deep(.arco-tabs-tab) {
  font-size: clamp(14px, 1.6vw, 16px);
}

:deep(.arco-card-body) {
  padding: min(16px, 3vw);
}

:deep(.arco-table-th) {
  font-size: clamp(13px, 1.5vw, 14px);
  padding: min(8px, 1vw);
}

:deep(.arco-table-td) {
  font-size: clamp(13px, 1.5vw, 14px);
  padding: min(8px, 1vw);
}

:deep(.arco-collapse-item-header-title) {
  font-size: clamp(14px, 1.6vw, 15px);
}

:deep(.arco-statistic-title) {
  font-size: clamp(13px, 1.5vw, 14px);
  margin-bottom: 4px;
}

:deep(.arco-statistic-value) {
  font-size: clamp(20px, 3vw, 24px);
}

/* 适配小屏幕 */
@media screen and (max-width: 768px) {
  .exam-container {
    padding: 10px;
    height: 100vh;
  }

  .card-title {
    gap: 8px;
    padding: 10px;
    min-height: 70px;
  }

  .exam-title {
    font-size: 18px;
    line-height: 1.3;
  }

  .countdown {
    font-size: 16px;
    margin-top: 3px;
  }

  .browser-input-group {
    flex-direction: column;
  }

  .behavior-log-item {
    flex-direction: column;
    gap: 4px;
  }

  .behavior-time, .behavior-type {
    min-width: auto;
  }

  .logout-btn {
    top: 8px;
    right: 8px;
  }
  
  :deep(.arco-tabs-tab) {
    padding: 8px 12px;
  }
  
  :deep(.arco-collapse-item-header) {
    padding: 8px 12px;
  }
  
  :deep(.arco-collapse-item-content-box) {
    padding: 8px;
  }
  
  :deep(.arco-row) {
    margin-left: -8px !important;
    margin-right: -8px !important;
  }
  
  :deep(.arco-col) {
    padding-left: 8px !important;
    padding-right: 8px !important;
  }
}

/* 超小屏幕 */
@media screen and (max-width: 480px) {
  .exam-container {
    padding: 8px;
  }
  
  .card-title {
    padding: 8px;
    min-height: 60px;
  }
  
  .exam-title {
    font-size: 16px;
  }
  
  .countdown {
    font-size: 14px;
  }
  
  .logout-btn button {
    padding: 0 6px;
    font-size: 12px;
  }
  
  :deep(.arco-tabs-nav-tab-list) {
    padding: 0 6px;
  }
  
  :deep(.arco-tabs-tab) {
    padding: 6px 8px;
    font-size: 13px;
  }
  
  :deep(.arco-statistic-value) {
    font-size: 18px;
  }
  
  :deep(.arco-descriptions-title) {
    font-size: 15px;
    margin-bottom: 10px;
  }
  
  :deep(.arco-table-th),
  :deep(.arco-table-td) {
    padding: 6px;
    font-size: 12px;
  }
  
  :deep(.arco-empty-description) {
    font-size: 12px;
  }
}

/* 宽屏显示 */
@media screen and (min-width: 1440px) {
  .exam-container {
    padding: 30px;
  }
  
  .card-title {
    gap: 16px;
    padding: 20px 30px;
  }
  
  .exam-title {
    font-size: 28px;
  }
  
  .countdown {
    font-size: 20px;
  }
}
</style> 