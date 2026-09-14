import { useEffect, useState } from 'react'
import TextPressure from './TextPressure'

function usePrefersReducedMotion() {
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false)

  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
    const syncPreference = () => setPrefersReducedMotion(mediaQuery.matches)

    syncPreference()
    mediaQuery.addEventListener('change', syncPreference)
    return () => mediaQuery.removeEventListener('change', syncPreference)
  }, [])

  return prefersReducedMotion
}

export function LoginTextPressureBackground() {
  const prefersReducedMotion = usePrefersReducedMotion()

  return (
    <div className="login-text-pressure" aria-hidden="true">
      <div className="login-text-pressure__word">
        <TextPressure
          alpha={false}
          flex
          italic
          minFontSize={18}
          paused={prefersReducedMotion}
          scale
          stroke
          strokeColor="#8d6bff"
          text="XAdmin"
          textColor="#65a6ff"
          weight
          width
        />
      </div>
    </div>
  )
}
