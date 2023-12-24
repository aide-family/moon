module.exports = {
    env: {browser: true, es2020: true, node: true},
    extends: [
        'eslint:recommended',
        'plugin:@typescript-eslint/recommended',
        'plugin:react-hooks/recommended',
        'plugin:prettier/recommended',
        'plugin:react/jsx-runtime',
    ],
    parser: '@typescript-eslint/parser',
    parserOptions: {
        ecmaFeatures: {
            jsx: true
        },
        ecmaVersion: 'latest',
        sourceType: 'module'
    },
    settings: {
        react: {
            version: 'detect'
        },
        'html/html-extensions': ['.html', '.we'] // consider .html and .we files as HTML
    },
    plugins: ['react-refresh', 'react', '@typescript-eslint', 'html', 'react-hooks'],
    rules: {
        'react-refresh/only-export-components': 'warn',
        'prettier/prettier': 'error',
        'arrow-body-style': 'off',
        'prefer-arrow-callback': 'off',
        '@typescript-eslint/no-explicit-any': ['off']
        // '@typescript-eslint/no-var-requires': 0
    }
    // Error: Failed to load plugin 'prettier' declared in '.eslintrc.cjs': Cannot find module 'eslint-plugin-prettier'
    // 解决方案：npm install eslint-plugin-prettier --save-dev
}
