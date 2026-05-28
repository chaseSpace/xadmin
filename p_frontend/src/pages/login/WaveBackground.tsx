import { useEffect, useRef } from 'react'

/**
 * 炫酷登录背景：极光渐变波浪 + 粒子连线网络 + 鼠标交互响应
 */
export function WaveBackground() {
  const canvasRef = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    let animId: number
    let time = 0
    let mouse = { x: -1000, y: -1000 }

    const resize = () => {
      const dpr = window.devicePixelRatio || 1
      canvas.width = canvas.offsetWidth * dpr
      canvas.height = canvas.offsetHeight * dpr
      ctx.scale(dpr, dpr)
    }
    resize()
    window.addEventListener('resize', resize)

    const onMouseMove = (e: MouseEvent) => {
      const rect = canvas.getBoundingClientRect()
      mouse = { x: e.clientX - rect.left, y: e.clientY - rect.top }
    }
    canvas.parentElement?.addEventListener('mousemove', onMouseMove)

    // 粒子系统
    const particleCount = 50
    const particles = Array.from({ length: particleCount }, () => ({
      x: Math.random() * canvas.offsetWidth,
      y: Math.random() * canvas.offsetHeight,
      vx: (Math.random() - 0.5) * 0.3,
      vy: (Math.random() - 0.5) * 0.3,
      radius: Math.random() * 2 + 1,
    }))

    // 极光色系
    const auroraColors = [
      ['rgba(15,109,255,0.12)', 'rgba(120,60,255,0.08)'],
      ['rgba(60,180,255,0.10)', 'rgba(15,109,255,0.06)'],
      ['rgba(120,60,255,0.08)', 'rgba(60,220,200,0.05)'],
      ['rgba(15,109,255,0.14)', 'rgba(60,180,255,0.06)'],
    ]

    const waves = [
      { amplitude: 55, wavelength: 900, speed: 0.001, yOffset: 0.72 },
      { amplitude: 40, wavelength: 700, speed: 0.0013, yOffset: 0.68 },
      { amplitude: 28, wavelength: 500, speed: 0.0016, yOffset: 0.63 },
      { amplitude: 48, wavelength: 800, speed: 0.0009, yOffset: 0.78 },
    ]

    const draw = () => {
      const w = canvas.offsetWidth
      const h = canvas.offsetHeight
      ctx.clearRect(0, 0, w, h)

      // 极光渐变波浪
      waves.forEach((wave, i) => {
        ctx.beginPath()
        const baseY = h * wave.yOffset
        for (let x = 0; x <= w; x += 3) {
          // 鼠标扰动
          const dx = x - mouse.x
          const dy = baseY - mouse.y
          const dist = Math.sqrt(dx * dx + dy * dy)
          const influence = dist < 150 ? (1 - dist / 150) * 20 : 0

          const y =
            baseY +
            Math.sin((x / wave.wavelength) * Math.PI * 2 + time * wave.speed * 60) * wave.amplitude +
            Math.sin((x / (wave.wavelength * 0.6)) * Math.PI * 2 - time * wave.speed * 25) * (wave.amplitude * 0.25) -
            influence
          if (x === 0) ctx.moveTo(x, y)
          else ctx.lineTo(x, y)
        }
        ctx.lineTo(w, h)
        ctx.lineTo(0, h)
        ctx.closePath()

        const grad = ctx.createLinearGradient(0, h * wave.yOffset - wave.amplitude, w, h)
        grad.addColorStop(0, auroraColors[i][0])
        grad.addColorStop(1, auroraColors[i][1])
        ctx.fillStyle = grad
        ctx.fill()
      })

      // 粒子更新与绘制
      for (const p of particles) {
        // 鼠标排斥
        const dx = p.x - mouse.x
        const dy = p.y - mouse.y
        const dist = Math.sqrt(dx * dx + dy * dy)
        if (dist < 120) {
          const force = (1 - dist / 120) * 0.8
          p.vx += (dx / dist) * force
          p.vy += (dy / dist) * force
        }

        p.x += p.vx
        p.y += p.vy
        p.vx *= 0.99
        p.vy *= 0.99

        // 边界回弹
        if (p.x < 0 || p.x > w) p.vx *= -1
        if (p.y < 0 || p.y > h) p.vy *= -1
        p.x = Math.max(0, Math.min(w, p.x))
        p.y = Math.max(0, Math.min(h, p.y))

        ctx.beginPath()
        ctx.arc(p.x, p.y, p.radius, 0, Math.PI * 2)
        ctx.fillStyle = 'rgba(15,109,255,0.25)'
        ctx.fill()
      }

      // 粒子连线
      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const dx = particles[i].x - particles[j].x
          const dy = particles[i].y - particles[j].y
          const dist = Math.sqrt(dx * dx + dy * dy)
          if (dist < 120) {
            ctx.beginPath()
            ctx.moveTo(particles[i].x, particles[i].y)
            ctx.lineTo(particles[j].x, particles[j].y)
            ctx.strokeStyle = `rgba(15,109,255,${0.12 * (1 - dist / 120)})`
            ctx.lineWidth = 0.5
            ctx.stroke()
          }
        }
      }

      time += 1
      animId = requestAnimationFrame(draw)
    }

    draw()

    return () => {
      cancelAnimationFrame(animId)
      window.removeEventListener('resize', resize)
      canvas.parentElement?.removeEventListener('mousemove', onMouseMove)
    }
  }, [])

  return (
    <canvas
      ref={canvasRef}
      style={{
        position: 'absolute',
        inset: 0,
        width: '100%',
        height: '100%',
        pointerEvents: 'none',
      }}
    />
  )
}
