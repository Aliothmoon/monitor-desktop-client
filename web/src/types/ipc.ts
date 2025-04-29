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

// 设备检测相关接口
export interface ISystemInfo {
  cpu: string;
  memory: string;
  os: string;
  version: string;
}

export interface ICheckResult {
  success: boolean;
  message?: string;
}

// 定义IPC监听事件类型
export interface IpcEvents {
  'systemInfo': (info: string) => void;
  'loginResult': (success: boolean, examInfo: IExamInfo) => void;
  'browserVisit': (url: string) => void;
  'processInfo': (processes: IProcessInfo[]) => void;
  'systemCheckResult': (result: ICheckResult) => void;
  'screenshotCheckResult': (result: ICheckResult) => void;
  'browserCheckResult': (result: ICheckResult) => void;
}

// 定义IPC调用类型
export interface IpcCalls {
  'load': () => void
  'login': (username: string, password: string) => void;
  'openBrowser': () => void;
  'logout': () => void;
  'checkSystemInfo': () => void;
  'checkScreenshot': () => void;
  'checkBrowser': () => void;
} 