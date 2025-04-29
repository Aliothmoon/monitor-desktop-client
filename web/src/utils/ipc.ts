import type {IpcCalls, IpcEvents, ICheckResult} from '../types/ipc';

// 定义全局IPC对象
declare global {
  interface Window {
    ipc: {
      on: <K extends keyof IpcEvents>(channel: K, callback: IpcEvents[K]) => void;
      emit: <K extends keyof IpcCalls>(channel: K, args: any[]) => void;
    }
  }
}

// IPC工具类
class IpcService {
  // 监听事件
  on<K extends keyof IpcEvents>(channel: K, callback: IpcEvents[K]): void {
    if (window.ipc) {
      window.ipc.on(channel, callback);
    } else {
      console.error('IPC不可用，可能在浏览器环境中运行');
    }
  }

  // 发送事件
  emit<K extends keyof IpcCalls>(channel: K, ...args: Parameters<IpcCalls[K]>): void {
    if (window.ipc) {
      window.ipc.emit(channel, [...args]);
    } else {
      console.error('IPC不可用，可能在浏览器环境中运行');
    }
  }

  // 用户登录
  login(username: string, password: string): void {
    this.emit('login', username, password);
  }

  // 打开浏览器
  openBrowser(): void {
    this.emit('openBrowser');
  }
  
  // 用户登出
  logout(): void {
    this.emit('logout');
  }
  
  // 检测系统信息
  checkSystemInfo(): void {
    this.emit('checkSystemInfo');
  }
  
  // 检测截屏功能
  checkScreenshot(): void {
    this.emit('checkScreenshot');
  }
  
  // 检测浏览器功能
  checkBrowser(): void {
    this.emit('checkBrowser');
  }

  // 检测系统信息并返回Promise
  checkSystemInfoAsync(): Promise<ICheckResult> {
    return new Promise((resolve) => {
      let resolved = false;
      
      // 发送检测请求
      this.checkSystemInfo();
      
      // 监听检测结果
      this.on('systemCheckResult', (result) => {
        if (!resolved) {
          resolved = true;
          resolve(result);
        }
      });
      
      // 超时处理
      setTimeout(() => {
        if (!resolved) {
          resolved = true;
          resolve({ success: false, message: '检测超时，请重试' });
        }
      }, 5000);
    });
  }
  
  // 检测截屏功能并返回Promise
  checkScreenshotAsync(): Promise<ICheckResult> {
    return new Promise((resolve) => {
      let resolved = false;
      
      // 发送检测请求
      this.checkScreenshot();
      
      // 监听检测结果
      this.on('screenshotCheckResult', (result) => {
        if (!resolved) {
          resolved = true;
          resolve(result);
        }
      });
      
      // 超时处理
      setTimeout(() => {
        if (!resolved) {
          resolved = true;
          resolve({ success: false, message: '检测超时，请重试' });
        }
      }, 5000);
    });
  }
  
  // 检测浏览器功能并返回Promise
  checkBrowserAsync(): Promise<ICheckResult> {
    return new Promise((resolve) => {
      let resolved = false;
      
      // 发送检测请求
      this.checkBrowser();
      
      // 监听检测结果
      this.on('browserCheckResult', (result) => {
        if (!resolved) {
          resolved = true;
          resolve(result);
        }
      });
      
      // 超时处理
      setTimeout(() => {
        if (!resolved) {
          resolved = true;
          resolve({ success: false, message: '检测超时，请重试' });
        }
      }, 5000);
    });
  }
}

export default new IpcService(); 