import { createContext } from 'react'

export type SpanceInfoType = {
  id: string
  name: string
}

export type GlobalContextType = {
  lang?: string
  setLang?: (value: string) => void
  theme?: string
  setTheme?: (value: string) => void
  spaceInfo?: SpanceInfoType
  setSpaceInfo?: (value: string) => void
  spaceId?: string
  setSpaceId?: (value: string) => void
}

export const GlobalContext = createContext<GlobalContextType>({
  lang: 'zh-CN',
  setLang: () => {},
  theme: 'light',
  setTheme: () => {},
  spaceInfo: JSON.parse(
    localStorage.getItem('spaceInfo') || `{"id": "","name": ""}`
  ) as SpanceInfoType,
  setSpaceInfo: () => {},
  spaceId: '',
  setSpaceId: () => {},
})
