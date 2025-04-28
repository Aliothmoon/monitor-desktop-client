import {createApp} from 'vue'
import App from './App.vue'
import './style.css'

import ArcoVue from '@arco-design/web-vue'
import '@arco-design/web-vue/dist/arco.css'
import { enableAllSecurity } from './utils/security'

// 启用安全措施，阻止开发者工具和右键菜单
enableAllSecurity();

const app = createApp(App)
app.use(ArcoVue)
app.mount('#app')
