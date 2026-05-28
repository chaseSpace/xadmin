import { forwardRef, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Button } from 'antd'
import type { ButtonProps } from 'antd'

export type UiButtonProps = ButtonProps & {
  debounceMs?: number
  disableDebounce?: boolean
  requestLoadingDelayMs?: number
  disableAutoRequestLoading?: boolean
}

function isPromiseLike(value: unknown): value is Promise<unknown> {
  return Boolean(value && typeof value === 'object' && 'then' in value)
}

export const UiButton = forwardRef<HTMLButtonElement, UiButtonProps>(function UiButton(
  {
    onClick,
    debounceMs = 300,
    disableDebounce = false,
    requestLoadingDelayMs = 1000,
    disableAutoRequestLoading = false,
    loading,
    ...rest
  },
  ref,
) {
  const lastTriggeredAtRef = useRef(0)
  const loadingTimerRef = useRef<number | null>(null)
  const [internalLoading, setInternalLoading] = useState(false)

  useEffect(() => {
    return () => {
      if (loadingTimerRef.current) {
        window.clearTimeout(loadingTimerRef.current)
      }
    }
  }, [])

  const handleClick: ButtonProps['onClick'] = useCallback(
    (event) => {
      if (!onClick) return
      if (disableDebounce || debounceMs <= 0) {
        const result = onClick(event)
        if (!disableAutoRequestLoading && isPromiseLike(result)) {
          loadingTimerRef.current = window.setTimeout(() => {
            setInternalLoading(true)
          }, requestLoadingDelayMs)
          void result.finally(() => {
            if (loadingTimerRef.current) {
              window.clearTimeout(loadingTimerRef.current)
              loadingTimerRef.current = null
            }
            setInternalLoading(false)
          })
        }
        return
      }
      const now = Date.now()
      if (now - lastTriggeredAtRef.current < debounceMs) return
      lastTriggeredAtRef.current = now
      const result = onClick(event)
      if (!disableAutoRequestLoading && isPromiseLike(result)) {
        loadingTimerRef.current = window.setTimeout(() => {
          setInternalLoading(true)
        }, requestLoadingDelayMs)
        void result.finally(() => {
          if (loadingTimerRef.current) {
            window.clearTimeout(loadingTimerRef.current)
            loadingTimerRef.current = null
          }
          setInternalLoading(false)
        })
      }
    },
    [onClick, disableDebounce, debounceMs, disableAutoRequestLoading, requestLoadingDelayMs],
  )

  const finalLoading = useMemo(() => {
    if (typeof loading === 'boolean') return loading || internalLoading
    if (loading && typeof loading === 'object') {
      return {
        ...loading,
        delay: loading.delay ?? requestLoadingDelayMs,
      }
    }
    return internalLoading
  }, [loading, internalLoading, requestLoadingDelayMs])

  return <Button {...rest} ref={ref} loading={finalLoading} onClick={handleClick} />
})
