import type { IpcCalls, IpcEvents } from '../types/ipc';

// 定义全局IPC对象
declare global {
  interface Window {
    ipc: {
      on: <K extends keyof IpcEvents>(channel: K, callback: IpcEvents[K]) => void;
      emit: <K extends keyof IpcCalls>(channel: K, args: Array<any>) => void;
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
}

export default new IpcService(); 