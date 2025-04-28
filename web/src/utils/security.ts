/**
 * 安全相关工具函数
 * 用于提供防止用户访问开发者工具等安全功能
 */

/**
 * 阻止开发者工具的常用快捷键
 */
export function preventDevTools(): void {
  // 阻止打开开发者工具的快捷键
  document.addEventListener('keydown', (event) => {
    // 阻止 F12 键
    if (event.key === 'F12') {
      event.preventDefault();
      return false;
    }
    
    // 阻止 Ctrl+Shift+I / Command+Option+I
    if ((event.ctrlKey || event.metaKey) && event.shiftKey && (event.key === 'I' || event.key === 'i')) {
      event.preventDefault();
      return false;
    }
    
    // 阻止 Ctrl+Shift+J / Command+Option+J (JavaScript 控制台)
    if ((event.ctrlKey || event.metaKey) && event.shiftKey && (event.key === 'J' || event.key === 'j')) {
      event.preventDefault();
      return false;
    }
    
    // 阻止 Ctrl+Shift+C / Command+Option+C (检查元素)
    if ((event.ctrlKey || event.metaKey) && event.shiftKey && (event.key === 'C' || event.key === 'c')) {
      event.preventDefault();
      return false;
    }

    // 阻止 Ctrl+U (查看源代码)
    if ((event.ctrlKey || event.metaKey) && (event.key === 'U' || event.key === 'u')) {
      event.preventDefault();
      return false;
    }
  }, true);
}

/**
 * 阻止右键菜单
 */
export function preventContextMenu(): void {
  document.addEventListener('contextmenu', (event) => {
    event.preventDefault();
    return false;
  }, true);
}

/**
 * 禁用浏览器密码管理器
 * 防止浏览器弹出保存密码的提示
 */
export function disablePasswordManager(): void {
  // 为所有的密码输入框添加 autocomplete="off" 和 autocomplete="new-password" 属性
  const applyPasswordAttributes = () => {
    // 立即处理现有的密码输入框
    const passwordInputs = document.querySelectorAll('input[type="password"]');
    passwordInputs.forEach(input => {
      input.setAttribute('autocomplete', 'new-password');
      input.setAttribute('readonly', 'readonly');
      // 添加短暂的只读状态，然后移除，可以阻止密码管理器识别
      setTimeout(() => {
        input.removeAttribute('readonly');
      }, 100);
    });

    // 处理登录表单
    const forms = document.querySelectorAll('form');
    forms.forEach(form => {
      form.setAttribute('autocomplete', 'off');
    });
  };

  // 初次加载时处理
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', applyPasswordAttributes);
  } else {
    applyPasswordAttributes();
  }

  // 使用MutationObserver监听DOM变化，处理动态添加的元素
  const observer = new MutationObserver(mutations => {
    mutations.forEach(mutation => {
      if (mutation.type === 'childList' && mutation.addedNodes.length > 0) {
        applyPasswordAttributes();
      }
    });
  });

  // 观察整个文档的变化
  observer.observe(document.documentElement, {
    childList: true,
    subtree: true
  });
}

/**
 * 阻止浏览器常见快捷键，使应用表现得更像原生客户端
 */
export function preventBrowserShortcuts(): void {
  // 阻止键盘快捷键
  document.addEventListener('keydown', (event) => {
    // 阻止 Ctrl+P (打印)
    if ((event.ctrlKey || event.metaKey) && (event.key === 'p' || event.key === 'P')) {
      event.preventDefault();
      return false;
    }

    // 阻止 Ctrl+S (保存)
    if ((event.ctrlKey || event.metaKey) && (event.key === 's' || event.key === 'S')) {
      event.preventDefault();
      return false;
    }

    // 阻止 Ctrl+O (打开)
    if ((event.ctrlKey || event.metaKey) && (event.key === 'o' || event.key === 'O')) {
      event.preventDefault();
      return false;
    }

    // 阻止 Ctrl+N (新窗口)
    if ((event.ctrlKey || event.metaKey) && (event.key === 'n' || event.key === 'N')) {
      event.preventDefault();
      return false;
    }

    // 阻止 Alt+Left/Right (前进/后退)
    if (event.altKey && (event.key === 'ArrowLeft' || event.key === 'ArrowRight')) {
      event.preventDefault();
      return false;
    }

    // 阻止 F5 或 Ctrl+R (刷新页面)
    if (event.key === 'F5' || ((event.ctrlKey || event.metaKey) && (event.key === 'r' || event.key === 'R'))) {
      event.preventDefault();
      return false;
    }
  }, true);

  // 阻止鼠标滚轮缩放
  document.addEventListener('wheel', (event) => {
    if (event.ctrlKey || event.metaKey) {
      event.preventDefault();
      return false;
    }
  }, { passive: false, capture: true });

  // 阻止双指缩放 (触控板)
  document.addEventListener('gesturestart', (event) => {
    event.preventDefault();
    return false;
  }, { passive: false, capture: true });

  document.addEventListener('gesturechange', (event) => {
    event.preventDefault();
    return false;
  }, { passive: false, capture: true });

  document.addEventListener('gestureend', (event) => {
    event.preventDefault();
    return false;
  }, { passive: false, capture: true });

  // 阻止拖放文件
  document.addEventListener('dragover', (event) => {
    event.preventDefault();
    return false;
  }, true);

  document.addEventListener('drop', (event) => {
    event.preventDefault();
    return false;
  }, true);

  // 为整个页面添加CSS禁用用户选择和缩放
  const style = document.createElement('style');
  style.innerHTML = `
    html, body {
      user-select: none;
      -webkit-user-select: none;
      -moz-user-select: none;
      -ms-user-select: none;
      touch-action: manipulation;
      overscroll-behavior: none;
    }
  `;
  document.head.appendChild(style);

  // 全局添加一个meta标签，禁止页面缩放
  const metaViewport = document.createElement('meta');
  metaViewport.name = 'viewport';
  metaViewport.content = 'width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no';
  document.head.appendChild(metaViewport);
}

/**
 * 同时启用所有安全措施
 */
export function enableAllSecurity(): void {
  preventDevTools();
  // preventContextMenu();
  preventBrowserShortcuts();
  disablePasswordManager();
}

export default {
  preventDevTools,
  preventContextMenu,
  preventBrowserShortcuts,
  disablePasswordManager,
  enableAllSecurity
}; 