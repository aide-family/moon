import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

const securityApi = 'https://security.api.bitrecx.com/'
// const assetApi = 'http://192.168.10.2:8002/'
const assetApi = 'http://localhost:8002/'

// https://vitejs.dev/config/
/** @type {import('vite').UserConfig} */
export default defineConfig({
    plugins: [react()],
    define: {
        'process.env': {
            REACT_APP_ASSET_API: assetApi,
            REACT_APP_SECURITY_API: securityApi
        }
    },
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src')
        }
    },
    css: {
        // 预处理器配置项
        preprocessorOptions: {
            less: {
                math: 'always'
            }
        }
    }
})
