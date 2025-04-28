import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import {vitePluginForArco} from '@arco-plugins/vite-vue'
// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(),
    vitePluginForArco({
    style: 'css'
  })],
})
