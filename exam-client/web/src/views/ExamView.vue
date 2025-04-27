<template>
  <div class="exam-container">
    <a-card class="exam-card">
      <div class="exam-header">
        <div class="exam-title-container">
          <a-badge status="processing" text="">
            <div class="exam-title-content">
              <a-typography-title :heading="3" :ellipsis="{ rows: 1, showTooltip: true }">
                {{ examInfo.title }}
              </a-typography-title>
            </div>
          </a-badge>
        </div>

        <div class="exam-timer">
          <a-space direction="vertical" size="mini" align="end">
            <a-typography-text type="secondary">剩余时间</a-typography-text>
            <a-space>
              <icon-clock-circle style="color: var(--color-danger);"/>
              <a-typography-text type="danger" bold style="font-size: 18px">
                {{ countdown }}
              </a-typography-text>
            </a-space>
          </a-space>
        </div>
      </div>

      <div class="main-content">
        <a-descriptions :data="examInfoData" layout="inline-vertical" :column="{ xs: 1, sm: 2 }" bordered size="medium"
                        title="考试信息"/>

        <a-tabs v-model:active-key="activeTab" class="exam-tabs">
          <a-tab-pane key="exam" title="考试内容">
            <div class="tab-content">
              <a-card title="考试说明" class="content-card">
                <a-scrollbar style="max-height: calc(100vh - 350px); min-height: 200px">
                  <p>这里是考试说明和内容。实际应用中这里可以根据需要加载考试题目。</p>
                  <!-- 可以在此处添加更多考试内容 -->
                </a-scrollbar>
              </a-card>
            </div>
          </a-tab-pane>

          <a-tab-pane key="browser" title="浏览器">
            <div class="tab-content">
              <a-space direction="vertical" fill size="medium">
                <a-input-group>
                  <a-input v-model="browserUrl" placeholder="输入网址" allow-clear @press-enter="openBrowser"/>
                  <a-button type="primary" @click="openBrowser">
                    <template #icon>
                      <icon-link/>
                    </template>
                    打开浏览器
                  </a-button>
                </a-input-group>

                <a-card title="浏览器访问记录" class="records-card">
                  <a-empty v-if="browserVisits.length === 0" description="无访问记录"/>
                  <a-scrollbar v-else outer-class="scrollbar-container" :style="{height: `${tableHeight}px`}">
                    <a-list :bordered="false">
                      <a-list-item v-for="(visit, index) in browserVisits" :key="index">
                        <a-space>
                          <a-avatar shape="square">
                            <icon-globe/>
                          </a-avatar>
                          <div>
                            <div>{{ visit.url }}</div>
                            <a-typography-text type="secondary">{{ visit.time }}</a-typography-text>
                          </div>
                        </a-space>
                      </a-list-item>
                    </a-list>
                  </a-scrollbar>
                </a-card>
              </a-space>
            </div>
          </a-tab-pane>

          <a-tab-pane key="monitor" title="系统监控">
            <div class="tab-content">
              <a-card title="系统进程信息" class="process-card">
                <a-table
                    :columns="processColumns"
                    :data="processes"
                    :bordered="true"
                    :pagination="false"
                    :scroll="{ y: tableHeight }"
                    :row-class="() => 'table-row'"
                >
                  <template #loading>
                    <a-empty description="加载中..."/>
                  </template>
                </a-table>
              </a-card>
            </div>
          </a-tab-pane>
        </a-tabs>
      </div>
    </a-card>
  </div>
</template>

<script lang="ts">
import type {PropType} from 'vue';
import {computed, defineComponent, onMounted, onUnmounted, ref} from 'vue';
import type {TableColumnData} from '@arco-design/web-vue';
import ipcService from '../utils/ipc';
import type {IExamInfo, IProcessInfo} from '../types/ipc';

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
    // 当前激活的标签页
    const activeTab = ref('exam');

    // 浏览器相关数据
    const browserUrl = ref('');
    const browserVisits = ref<BrowserVisit[]>([]);
    const visitedUrls = new Set<string>();

    // 进程信息相关数据
    const processes = ref<IProcessInfo[]>([]);

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

    // 添加表格高度计算
    const tableHeight = ref(200);

    // 动态调整表格高度函数
    const updateTableHeight = () => {
      const viewportHeight = window.innerHeight;
      if (viewportHeight < 650) {
        tableHeight.value = 200;
      } else if (viewportHeight < 800) {
        tableHeight.value = 300;
      } else {
        tableHeight.value = 400;
      }
    };

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
          browserVisits.value.unshift({time, url});
        }
      });

      // 监听进程信息
      ipcService.on('processInfo', (processList) => {
        processes.value = processList;
      });

      // 添加窗口尺寸变化监听
      updateTableHeight();
      window.addEventListener('resize', updateTableHeight);
    });

    // 组件销毁时清除定时器
    onUnmounted(() => {
      clearInterval(timerId);

      // 移除事件监听
      window.removeEventListener('resize', updateTableHeight);
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
      examInfoData,
      processColumns,
      openBrowser,
      tableHeight
    };
  }
});
</script>

<style scoped>
.exam-container {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
}

.exam-card {
  width: min(96%, 1200px);
  height: 96%;
  margin: auto;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
}

.main-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.exam-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid var(--color-border-2);
  background-color: var(--color-bg-2);
}

.exam-title-container {
  flex: 1;
  margin-right: 16px;
  overflow: hidden;
}

.exam-title-content {
  display: flex;
  align-items: center;
}

.exam-title-content :deep(.arco-typography) {
  margin-bottom: 0;
}

.exam-timer {
  flex-shrink: 0;
  text-align: right;
}

.exam-tabs {
  flex: 1;
  margin-top: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

:deep(.arco-tabs-content) {
  flex: 1;
  overflow: hidden;
}

:deep(.arco-tabs-content-inner) {
  height: 100%;
}

:deep(.arco-tab-pane) {
  height: 100%;
}

.tab-content {
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.content-card, .records-card, .process-card {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

:deep(.arco-card-body) {
  flex: 1;
  overflow: hidden;
}

.scrollbar-container {
  height: 100%;
}

:deep(.arco-descriptions-title) {
  font-size: 15px;
  margin-bottom: 12px;
}

@media screen and (max-height: 650px) {
  .exam-header {
    padding: 8px 12px;
  }

  :deep(.arco-descriptions-title) {
    font-size: 14px;
    margin-bottom: 8px;
  }

  .exam-tabs {
    margin-top: 8px;
  }
}

/* 添加表格行样式 */
:deep(.table-row) {
  font-size: 14px;
}

:deep(.arco-table-container) {
  border-radius: 4px;
}

:deep(.arco-table-header) {
  background-color: var(--color-fill-2);
  font-weight: 500;
}

:deep(.arco-table-tr:hover) {
  background-color: var(--color-fill-1);
}

/* 响应式表格高度 */
@media screen and (max-height: 768px) {
  :deep(.arco-table) {
    --table-max-height: 300px;
  }
}

@media screen and (max-height: 650px) {
  :deep(.arco-table) {
    --table-max-height: 200px;
  }
}
</style> 