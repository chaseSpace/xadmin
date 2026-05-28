type UnauthorizedHandlerDeps = {
  unauthorizedRedirectingRef: { current: boolean }
  showConfirm: (config: {
    title: string
    content: string
    okText: string
    cancelText: string
    onOk: () => void | Promise<void>
    onCancel: () => void
  }) => void
  logout: () => void
  clearQueryClient: () => void
  navigateToLogin: () => Promise<void> | void
}

export function createUnauthorizedHandler(deps: UnauthorizedHandlerDeps) {
  return () => {
    if (deps.unauthorizedRedirectingRef.current) {
      return
    }

    deps.unauthorizedRedirectingRef.current = true
    const unlockTimer = globalThis.setTimeout(() => {
      if (deps.unauthorizedRedirectingRef.current) {
        deps.unauthorizedRedirectingRef.current = false
      }
    }, 15_000)

    const releaseLock = () => {
      globalThis.clearTimeout(unlockTimer)
      deps.unauthorizedRedirectingRef.current = false
    }

    try {
      deps.showConfirm({
        title: '登录状态已失效',
        content: '当前会话已过期，是否立即跳转登录页重新登录？',
        okText: '去登录',
        cancelText: '取消',
        onOk: async () => {
          try {
            deps.logout()
            await deps.navigateToLogin()
            deps.clearQueryClient()
          } finally {
            releaseLock()
          }
        },
        onCancel: () => {
          releaseLock()
        },
      })
    } catch {
      releaseLock()
    }
  }
}
