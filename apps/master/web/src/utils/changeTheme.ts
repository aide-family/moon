export type ThemeType = 'light' | 'dark'

function changeTheme(theme: ThemeType) {
  // 从document获取micro-app-body
  const body = document.querySelector('micro-app-body') || document.body
  if (theme === 'dark') {
    body.setAttribute('arco-theme', 'dark')
  } else {
    body.removeAttribute('arco-theme')
  }

  localStorage.setItem('theme', theme)
}

export default changeTheme
