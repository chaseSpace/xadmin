import { useCallback, useRef } from 'react'

export function useDebouncedCallback<T extends (...args: never[]) => void>(callback: T, waitMs = 300): T {
  const lastCalledAtRef = useRef(0)

  const debounced = useCallback(
    (...args: Parameters<T>) => {
      const now = Date.now()
      if (now - lastCalledAtRef.current < waitMs) return
      lastCalledAtRef.current = now
      callback(...args)
    },
    [callback, waitMs],
  )

  return debounced as T
}
