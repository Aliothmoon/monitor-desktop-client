// IPC接口类型定义

export interface IExamInfo {
  examId: string;
  title: string;
  startTime: string;
  endTime: string;
  duration: number;
  studentName: string;
}

export interface IProcessInfo {
  pid: number;
  name: string;
  memory: number;
  cpu: number;
}

// 定义IPC监听事件类型
export interface IpcEvents {
  'systemInfo': (info: string) => void;
  'loginResult': (success: boolean, examInfo: IExamInfo) => void;
  'browserVisit': (url: string) => void;
  'processInfo': (processes: IProcessInfo[]) => void;
}

// 定义IPC调用类型
export interface IpcCalls {
  'login': (username: string, password: string) => void;
  'openBrowser': (url: string) => void;
} 