// React Bits TextPressure (TS-CSS), ported from https://codepen.io/JuanFuentes/full/rgXKGQ
// https://www.reactbits.dev/text-animations/text-pressure

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

interface TextPressureProps {
  text?: string
  fontFamily?: string
  fontUrl?: string
  width?: boolean
  weight?: boolean
  italic?: boolean
  alpha?: boolean
  flex?: boolean
  stroke?: boolean
  scale?: boolean
  textColor?: string
  strokeColor?: string
  className?: string
  minFontSize?: number
  paused?: boolean
}

const dist = (a: { x: number; y: number }, b: { x: number; y: number }) => {
  const dx = b.x - a.x
  const dy = b.y - a.y
  return Math.sqrt(dx * dx + dy * dy)
}

const getAttr = (distance: number, maxDist: number, minVal: number, maxVal: number) => {
  const val = maxVal - Math.abs((maxVal * distance) / maxDist)
  return Math.max(minVal, val + minVal)
}

const debounce = (callback: () => void, delay: number) => {
  let timeoutId: ReturnType<typeof setTimeout>
  return () => {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(callback, delay)
  }
}

export default function TextPressure({
  text = 'Compressa',
  fontFamily = 'Roboto Flex',
  fontUrl = 'https://fonts.googleapis.com/css2?family=Roboto+Flex:opsz,wdth,wght@8..144,25..151,100..1000&display=swap',
  width = true,
  weight = true,
  italic = true,
  alpha = false,
  flex = true,
  stroke = false,
  scale = false,
  textColor = '#FFFFFF',
  strokeColor = '#FF0000',
  className = '',
  minFontSize = 24,
  paused = false,
}: TextPressureProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const titleRef = useRef<HTMLHeadingElement | null>(null)
  const spansRef = useRef<(HTMLSpanElement | null)[]>([])
  const mouseRef = useRef({ x: 0, y: 0 })
  const cursorRef = useRef({ x: 0, y: 0 })
  const [fontSize, setFontSize] = useState(minFontSize)
  const [scaleY, setScaleY] = useState(1)
  const [lineHeight, setLineHeight] = useState(1)
  const chars = text.split('')

  useEffect(() => {
    if (paused) return

    const handleMouseMove = (event: MouseEvent) => {
      cursorRef.current.x = event.clientX
      cursorRef.current.y = event.clientY
    }
    const handleTouchMove = (event: TouchEvent) => {
      const touch = event.touches[0]
      if (!touch) return
      cursorRef.current.x = touch.clientX
      cursorRef.current.y = touch.clientY
    }

    window.addEventListener('mousemove', handleMouseMove)
    window.addEventListener('touchmove', handleTouchMove, { passive: true })

    const container = containerRef.current
    if (container) {
      const {
        left,
        top,
        width: containerWidth,
        height: containerHeight,
      } = container.getBoundingClientRect()
      mouseRef.current.x = left + containerWidth / 2
      mouseRef.current.y = top + containerHeight / 2
      cursorRef.current.x = mouseRef.current.x
      cursorRef.current.y = mouseRef.current.y
    }

    return () => {
      window.removeEventListener('mousemove', handleMouseMove)
      window.removeEventListener('touchmove', handleTouchMove)
    }
  }, [paused])

  const setSize = useCallback(() => {
    const container = containerRef.current
    const title = titleRef.current
    if (!container || !title) return

    const { width: containerWidth, height: containerHeight } = container.getBoundingClientRect()
    let nextFontSize = containerWidth / (chars.length / 2)
    nextFontSize = Math.max(nextFontSize, minFontSize)

    setFontSize(nextFontSize)
    setScaleY(1)
    setLineHeight(1)

    requestAnimationFrame(() => {
      const currentTitle = titleRef.current
      if (!currentTitle || !scale) return

      const textRect = currentTitle.getBoundingClientRect()
      if (textRect.height <= 0) return

      const yRatio = containerHeight / textRect.height
      setScaleY(yRatio)
      setLineHeight(yRatio)
    })
  }, [chars.length, minFontSize, scale])

  useEffect(() => {
    const debouncedSetSize = debounce(setSize, 100)
    debouncedSetSize()
    window.addEventListener('resize', debouncedSetSize)
    return () => window.removeEventListener('resize', debouncedSetSize)
  }, [setSize])

  useEffect(() => {
    if (paused) return

    let animationFrameId: number
    const animate = () => {
      mouseRef.current.x += (cursorRef.current.x - mouseRef.current.x) / 15
      mouseRef.current.y += (cursorRef.current.y - mouseRef.current.y) / 15

      const title = titleRef.current
      if (title) {
        const titleRect = title.getBoundingClientRect()
        const maxDist = titleRect.width / 2

        spansRef.current.forEach((span) => {
          if (!span) return

          const rect = span.getBoundingClientRect()
          const charCenter = {
            x: rect.x + rect.width / 2,
            y: rect.y + rect.height / 2,
          }
          const distance = dist(mouseRef.current, charCenter)
          const wdth = width ? Math.floor(getAttr(distance, maxDist, 5, 200)) : 100
          const wght = weight ? Math.floor(getAttr(distance, maxDist, 100, 900)) : 400
          const italVal = italic ? getAttr(distance, maxDist, 0, 1).toFixed(2) : '0'
          const alphaVal = alpha ? getAttr(distance, maxDist, 0, 1).toFixed(2) : '1'
          const fontVariationSettings = `'wght' ${wght}, 'wdth' ${wdth}, 'ital' ${italVal}`

          if (span.style.fontVariationSettings !== fontVariationSettings) {
            span.style.fontVariationSettings = fontVariationSettings
          }
          if (alpha && span.style.opacity !== alphaVal) {
            span.style.opacity = alphaVal
          }
        })
      }

      animationFrameId = requestAnimationFrame(animate)
    }

    animate()
    return () => cancelAnimationFrame(animationFrameId)
  }, [alpha, italic, paused, weight, width])

  const styleElement = useMemo(() => {
    return (
      <style>{`
        @import url('${fontUrl}');

        .text-pressure-flex {
          display: flex;
          justify-content: space-between;
        }

        .text-pressure-stroke span {
          position: relative;
          color: ${textColor};
        }

        .text-pressure-stroke span::after {
          content: attr(data-char);
          position: absolute;
          top: 0;
          left: 0;
          z-index: -1;
          color: transparent;
          -webkit-text-stroke-width: 3px;
          -webkit-text-stroke-color: ${strokeColor};
        }

        .text-pressure-title {
          color: ${textColor};
        }
      `}</style>
    )
  }, [fontUrl, strokeColor, textColor])

  const dynamicClassName = [
    className,
    flex ? 'text-pressure-flex' : '',
    stroke ? 'text-pressure-stroke' : '',
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <div
      ref={containerRef}
      style={{ position: 'relative', width: '100%', height: '100%', background: 'transparent' }}
    >
      {styleElement}
      <h1
        ref={titleRef}
        className={`text-pressure-title ${dynamicClassName}`}
        style={{
          fontFamily,
          fontSize,
          fontWeight: 100,
          lineHeight,
          margin: 0,
          textAlign: 'center',
          transform: `scale(1, ${scaleY})`,
          transformOrigin: 'center top',
          userSelect: 'none',
          whiteSpace: 'nowrap',
          width: '100%',
        }}
      >
        {chars.map((char, index) => (
          <span
            key={`${char}-${index}`}
            ref={(element) => {
              spansRef.current[index] = element
            }}
            data-char={char}
            style={{ display: 'inline-block', color: stroke ? undefined : textColor }}
          >
            {char}
          </span>
        ))}
      </h1>
    </div>
  )
}
