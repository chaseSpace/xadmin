import { create } from 'zustand'
import { message } from 'antd'
import { useI18nStore } from './i18n'
import { translateText } from '../i18n/messages'

export type AppTabKey =
  | '/'
  | '/organization/departments'
  | '/organization/users'
  | '/organization/positions'
  | '/business/users'
  | '/business/user-punishments'
  | '/resource/files'
  | '/permission/role-permissions'
  | '/permission/menu-permissions'
  | '/system/settings'
  | '/system/audit-logs'
  | '/system/ip-blacklist'
  | '/system/warm-tips'
  | '/system/alert-bots'

export type AppTabItem = {
  key: AppTabKey
  title: string
}

const MAX_TAB_COUNT = 8

const MAX_TAB_WARNED_KEY = 'xadmin_max_tab_warned'

type PageTabsState = {
  tabs: AppTabItem[]
  touchTab: (tab: AppTabItem) => void
  closeTab: (key: AppTabKey) => void
  closeOtherTabs: (key: AppTabKey) => void
  resetTabs: () => void
}

export const usePageTabsStore = create<PageTabsState>((set) => ({
  tabs: [],
  touchTab: (tab) => {
    set((state) => {
      const existedIndex = state.tabs.findIndex((item) => item.key === tab.key)
      if (existedIndex >= 0) {
        const nextTabs = [...state.tabs]
        nextTabs[existedIndex] = {
          ...nextTabs[existedIndex],
          title: tab.title,
        }
        return { tabs: nextTabs }
      }

      const nextTabs = [...state.tabs, tab]
      if (nextTabs.length <= MAX_TAB_COUNT) {
        return { tabs: nextTabs }
      }
      if (!sessionStorage.getItem(MAX_TAB_WARNED_KEY)) {
        sessionStorage.setItem(MAX_TAB_WARNED_KEY, '1')
        const locale = useI18nStore.getState().locale
        void message.warning(translateText('最多驻留8个标签页，将移除最早打开的标签页', locale))
      }
      return { tabs: nextTabs.slice(nextTabs.length - MAX_TAB_COUNT) }
    })
  },
  closeTab: (key) => {
    set((state) => ({
      tabs: state.tabs.filter((item) => item.key !== key),
    }))
  },
  closeOtherTabs: (key) => {
    set((state) => ({
      tabs: state.tabs.filter((item) => item.key === key),
    }))
  },
  resetTabs: () => {
    set({ tabs: [] })
  },
}))
