import { create } from 'zustand'
import { clearAccessToken, hasAccessToken, setAccessToken } from '../services/auth/token'
import type { CurrentUserMenuItem, CurrentUserWarmTip } from '../services/api/account'

const USER_PROFILE_KEY = 'xadmin_user_profile'

export type AuthUserProfile = {
  uid: number
  username: string
  displayName: string
  avatar: string
  sessionId: string
  menuRoutes: string[]
  menuItems: CurrentUserMenuItem[]
  warmTip: CurrentUserWarmTip | null
  menuLoaded: boolean
  menuLoadError: string
}

function getInitialUserProfile(): AuthUserProfile | null {
  if (typeof window === 'undefined') return null
  const raw = window.localStorage.getItem(USER_PROFILE_KEY)
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw) as Partial<AuthUserProfile>
    if (!parsed || typeof parsed.uid !== 'number' || typeof parsed.username !== 'string') return null
    return {
      uid: parsed.uid,
      username: parsed.username,
      displayName: typeof parsed.displayName === 'string' ? parsed.displayName : '',
      avatar: typeof parsed.avatar === 'string' ? parsed.avatar : '',
      sessionId: typeof parsed.sessionId === 'string' ? parsed.sessionId : '',
      menuRoutes: Array.isArray(parsed.menuRoutes) ? parsed.menuRoutes.filter((item) => typeof item === 'string') : [],
      menuItems: Array.isArray(parsed.menuItems) ? parsed.menuItems : [],
      warmTip: parsed.warmTip && typeof parsed.warmTip === 'object' ? parsed.warmTip as CurrentUserWarmTip : null,
      menuLoaded: Boolean(parsed.menuLoaded),
      menuLoadError: typeof parsed.menuLoadError === 'string' ? parsed.menuLoadError : '',
    }
  } catch {
    return null
  }
}

function persistUserProfile(profile: AuthUserProfile | null): void {
  if (typeof window === 'undefined') return
  if (!profile) {
    window.localStorage.removeItem(USER_PROFILE_KEY)
    return
  }
  window.localStorage.setItem(USER_PROFILE_KEY, JSON.stringify(profile))
}

type AuthState = {
  isAuthenticated: boolean
  currentUser: AuthUserProfile | null
  login: (session: { token: string; user: AuthUserProfile }) => void
  updateCurrentUserProfile: (
    profile: Partial<Pick<AuthUserProfile, 'username' | 'displayName' | 'avatar' | 'menuRoutes' | 'menuItems' | 'warmTip' | 'menuLoaded' | 'menuLoadError'>>,
  ) => void
  logout: () => void
  syncFromStorage: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: hasAccessToken(),
  currentUser: getInitialUserProfile(),
  login: (session) => {
    setAccessToken(session.token)
    persistUserProfile(session.user)
    set({ isAuthenticated: true, currentUser: session.user })
  },
  updateCurrentUserProfile: (profile) => {
    set((state) => {
      if (!state.currentUser) {
        return state
      }
      const nextUser: AuthUserProfile = {
        ...state.currentUser,
        username: profile.username ?? state.currentUser.username,
        displayName: profile.displayName ?? state.currentUser.displayName,
        avatar: profile.avatar ?? state.currentUser.avatar,
        menuRoutes: profile.menuRoutes ?? state.currentUser.menuRoutes,
        menuItems: profile.menuItems ?? state.currentUser.menuItems,
        warmTip: profile.warmTip ?? state.currentUser.warmTip,
        menuLoaded: profile.menuLoaded ?? state.currentUser.menuLoaded,
        menuLoadError: profile.menuLoadError ?? state.currentUser.menuLoadError,
      }
      persistUserProfile(nextUser)
      return { ...state, currentUser: nextUser }
    })
  },
  logout: () => {
    clearAccessToken()
    persistUserProfile(null)
    sessionStorage.removeItem('xadmin_max_tab_warned')
    set({ isAuthenticated: false, currentUser: null })
  },
  syncFromStorage: () => {
    set({
      isAuthenticated: hasAccessToken(),
      currentUser: getInitialUserProfile(),
    })
  },
}))
