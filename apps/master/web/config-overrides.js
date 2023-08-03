/* eslint-disable @typescript-eslint/no-var-requires */
const path = require('path')
const {
    override,
    addWebpackModuleRule,
    addWebpackPlugin,
    addWebpackAlias,
    overrideDevServer,
} = require('customize-cra')
const ArcoWebpackPlugin = require('@arco-plugins/webpack-react')
const addLessLoader = require('customize-cra-less-loader')
const setting = require('./src/settings.json')

const devServerConfig = () => (config) => {
    return {
        ...config,
        compress: true,
        proxy: {
            [process.env.REACT_APP_MASTER_API]: {
                target: 'http://localhost:28000/',
                changeOrigin: true,
                pathRewrite: {
                    [`^${process.env.REACT_APP_MASTER_API}`]: '',
                },
            },
        },
    }
}

module.exports = {
    devServer: overrideDevServer(devServerConfig()),

    webpack: override(
        addLessLoader({
            lessLoaderOptions: {
                lessOptions: {},
            },
        }),
        addWebpackModuleRule({
            test: /\.svg$/,
            loader: '@svgr/webpack',
        }),
        addWebpackPlugin(
            new ArcoWebpackPlugin({
                theme: '@arco-themes/react-arco-pro',
                modifyVars: {
                    'arcoblue-6': setting.themeColor,
                },
            })
        ),
        addWebpackAlias({
            '@': path.resolve(__dirname, 'src'),
        })
    ),
}
