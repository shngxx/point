import { useState, useEffect, useRef, useCallback } from 'react'
import './App.css'

// Класс для частиц фона
class Particle {
  constructor(canvas) {
    this.canvas = canvas
    this.x = Math.random() * canvas.width
    this.y = Math.random() * canvas.height
    this.size = Math.random() * 2 + 0.5
    this.speedX = (Math.random() - 0.5) * 0.5
    this.speedY = (Math.random() - 0.5) * 0.5
    this.opacity = Math.random() * 0.5 + 0.2
    this.hue = Math.random() * 60 + 240 // Фиолетово-синие оттенки
  }

  update() {
    this.x += this.speedX
    this.y += this.speedY
    
    if (this.x < 0 || this.x > this.canvas.width) this.speedX *= -1
    if (this.y < 0 || this.y > this.canvas.height) this.speedY *= -1
    
    this.x = Math.max(0, Math.min(this.canvas.width, this.x))
    this.y = Math.max(0, Math.min(this.canvas.height, this.y))
  }

  draw(ctx) {
    // Упрощенная отрисовка для производительности
    ctx.globalAlpha = this.opacity
    ctx.fillStyle = `rgba(150, 100, 255, 0.6)`
    ctx.beginPath()
    ctx.arc(this.x, this.y, this.size, 0, Math.PI * 2)
    ctx.fill()
    ctx.globalAlpha = 1
  }
}

// Класс для партиклов следа (оптимизированный)
class TrailParticle {
  constructor(x, y) {
    this.x = x
    this.y = y
    this.life = 1.0
    this.decay = 0.03 // Быстрее исчезают
    this.size = Math.random() * 3 + 1.5 // Меньше размер
  }

  update() {
    this.life -= this.decay
    this.size *= 0.97
    return this.life > 0
  }

  draw(ctx) {
    // Упрощенная отрисовка без градиента (быстрее)
    ctx.globalAlpha = this.life * 0.6
    ctx.fillStyle = 'rgba(255, 0, 150, 0.8)'
    ctx.beginPath()
    ctx.arc(this.x, this.y, this.size, 0, Math.PI * 2)
    ctx.fill()
    ctx.globalAlpha = 1
  }
}

function App() {
  const canvasRef = useRef(null)
  const wsRef = useRef(null)
  const animationFrameRef = useRef(null)
  const targetPointRef = useRef({ x: 400, y: 300 })
  const particlesRef = useRef([])
  const trailParticlesRef = useRef([])
  const lastPositionRef = useRef({ x: 400, y: 300 })
  const pulsePhaseRef = useRef(0)
  const [point, setPoint] = useState({ x: 400, y: 300 }) // Целевая позиция
  const [displayPoint, setDisplayPoint] = useState({ x: 400, y: 300 }) // Визуальная позиция для плавной анимации
  const [pointID, setPointID] = useState(1)
  const [isConnected, setIsConnected] = useState(false)
  const [connectionStatus, setConnectionStatus] = useState('Отключено')
  const pointSize = 3
  const animationSpeed = 0.4 // Увеличена скорость анимации для более отзывчивого движения

  // Загрузка начальной позиции точки
  const fetchPointInfo = useCallback(async () => {
    try {
      const response = await fetch(`/api/point/${pointID}`)
      if (!response.ok) {
        throw new Error('Ошибка получения информации о точке')
      }
      const data = await response.json()
      setPointID(data.id)
      const newPoint = { x: data.point.x, y: data.point.y }
      setPoint(newPoint)
      setDisplayPoint(newPoint) // Сразу устанавливаем визуальную позицию при загрузке
      console.log(`Получена информация о точке: ID=${data.id}, X=${data.point.x}, Y=${data.point.y}`)
    } catch (error) {
      console.error('Ошибка при получении информации о точке:', error)
    }
  }, [pointID])

  // Подключение WebSocket
  const connect = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws`
    
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {
      console.log('WebSocket соединение установлено')
      setIsConnected(true)
      setConnectionStatus('Подключено')
    }

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      // Обновляем целевую позицию - анимация сама доведет точку плавно
      setPoint({ x: data.x, y: data.y })
    }

    ws.onerror = (error) => {
      console.error('WebSocket ошибка:', error)
      setIsConnected(false)
      setConnectionStatus('Ошибка соединения')
    }

    ws.onclose = () => {
      console.log('WebSocket соединение закрыто')
      setIsConnected(false)
      setConnectionStatus('Отключено')
      // Попытка переподключения через 3 секунды
      setTimeout(() => connect(), 3000)
    }
  }, [])

  // Отправка команды перемещения с оптимистичным обновлением
  const sendMove = useCallback((dx, dy) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      // Оптимистичное обновление - сразу двигаем точку визуально
      setPoint(prev => ({
        x: Math.max(0, Math.min(800, prev.x + dx)),
        y: Math.max(0, Math.min(600, prev.y + dy))
      }))
      
      const message = {
        action: 'move',
        dx: dx,
        dy: dy
      }
      wsRef.current.send(JSON.stringify(message))
    }
  }, [])

  // Обработка клавиатуры с throttle для плавности
  useEffect(() => {
    let lastKeyTime = 0
    const keyThrottle = 16 // ~60 FPS
    
    const handleKeyDown = (e) => {
      if (!isConnected) {
        return
      }

      const now = Date.now()
      if (now - lastKeyTime < keyThrottle) {
        return // Пропускаем слишком частые нажатия
      }
      lastKeyTime = now

      let dx = 0
      let dy = 0

      switch(e.key) {
        case 'ArrowUp':
          dy = -10
          e.preventDefault()
          break
        case 'ArrowDown':
          dy = 10
          e.preventDefault()
          break
        case 'ArrowLeft':
          dx = -10
          e.preventDefault()
          break
        case 'ArrowRight':
          dx = 10
          e.preventDefault()
          break
        default:
          return
      }

      if (dx !== 0 || dy !== 0) {
        sendMove(dx, dy)
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isConnected, sendMove])

  // Обновляем ref при изменении целевой позиции и проверяем рассинхронизацию
  useEffect(() => {
    targetPointRef.current = point
    
    // Если рассинхронизация большая (>30px), сразу синхронизируем визуальную позицию
    const dx = Math.abs(point.x - displayPoint.x)
    const dy = Math.abs(point.y - displayPoint.y)
    if (dx > 30 || dy > 30) {
      setDisplayPoint(point) // Сразу синхронизируем при большой рассинхронизации
    }
  }, [point, displayPoint])

  // Плавная анимация перемещения точки (оптимизированная)
  useEffect(() => {
    let rafId = null
    let lastTime = 0
    
    const animate = (currentTime) => {
      if (!lastTime) lastTime = currentTime
      const deltaTime = currentTime - lastTime
      lastTime = currentTime
      
      setDisplayPoint(prev => {
        const target = targetPointRef.current
        const dx = target.x - prev.x
        const dy = target.y - prev.y
        const distance = Math.sqrt(dx * dx + dy * dy)
        
        // Если расстояние очень мало, считаем что достигли цели
        if (distance < 0.5) {
          rafId = null
          return target
        }
        
        // Адаптивная скорость: быстрее для больших расстояний
        const speed = Math.min(animationSpeed * (1 + distance / 50), 0.9)
        
        // Интерполяция с учетом времени для плавности
        const factor = Math.min(speed * (deltaTime / 16.67), 1.0) // Нормализация к 60 FPS, ограничение максимумом
        const newX = prev.x + dx * factor
        const newY = prev.y + dy * factor
        
        // Продолжаем анимацию
        rafId = requestAnimationFrame(animate)
        
        return { x: newX, y: newY }
      })
    }
    
    // Запускаем анимацию при изменении целевой позиции
    const dx = point.x - displayPoint.x
    const dy = point.y - displayPoint.y
    const distance = Math.sqrt(dx * dx + dy * dy)
    
    if (distance > 0.5) {
      if (!rafId) {
        lastTime = 0
        rafId = requestAnimationFrame(animate)
      }
    }
    
    return () => {
      if (rafId) {
        cancelAnimationFrame(rafId)
        rafId = null
      }
    }
  }, [point, displayPoint, animationSpeed])

  // Инициализация
  useEffect(() => {
    fetchPointInfo()
    connect()
  }, [fetchPointInfo, connect])

  // Инициализация частиц фона
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    
    // Создаем меньше частиц для производительности
    const particles = []
    for (let i = 0; i < 15; i++) {
      particles.push(new Particle(canvas))
    }
    particlesRef.current = particles
  }, [])

  // Отрисовка точки на canvas с плавной анимацией и эффектами (оптимизированная)
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    let animationId = null
    let lastTime = 0
    const targetFPS = 60
    const frameInterval = 1000 / targetFPS
    
    const draw = (currentTime) => {
      // Ограничение FPS для производительности
      if (currentTime - lastTime < frameInterval) {
        animationId = requestAnimationFrame(draw)
        return
      }
      lastTime = currentTime
      
      // Очищаем canvas простым цветом (быстрее чем градиент)
      ctx.fillStyle = 'rgba(20, 20, 40, 0.95)'
      ctx.fillRect(0, 0, canvas.width, canvas.height)
      
      // Обновляем и рисуем частицы фона (реже обновляем)
      particlesRef.current.forEach(particle => {
        particle.update()
        particle.draw(ctx)
      })
      
      // Ограничиваем координаты границами canvas
      const x = Math.max(pointSize, Math.min(canvas.width - pointSize, displayPoint.x))
      const y = Math.max(pointSize, Math.min(canvas.height - pointSize, displayPoint.y))
      
      // Проверяем движение и создаем партиклы следа (меньше партиклов)
      const dx = x - lastPositionRef.current.x
      const dy = y - lastPositionRef.current.y
      const distance = Math.sqrt(dx * dx + dy * dy)
      
      if (distance > 1) {
        // Создаем меньше партиклов следа
        if (trailParticlesRef.current.length < 20) {
          trailParticlesRef.current.push(new TrailParticle(x, y))
        }
        lastPositionRef.current = { x, y }
      }
      
      // Обновляем и рисуем партиклы следа
      trailParticlesRef.current = trailParticlesRef.current.filter(particle => {
        particle.update()
        particle.draw(ctx)
        return particle.life > 0
      })
      
      // Обновляем фазу пульсации (медленнее)
      pulsePhaseRef.current += 0.03
      
      // Рисуем только один пульсирующий круг (вместо трех)
      const pulseRadius = 25 + Math.sin(pulsePhaseRef.current) * 8
      const pulseOpacity = 0.2 + Math.sin(pulsePhaseRef.current) * 0.15
      
      ctx.fillStyle = `rgba(255, 0, 150, ${pulseOpacity})`
      ctx.beginPath()
      ctx.arc(x, y, pulseRadius, 0, Math.PI * 2)
      ctx.fill()
      
      // Упрощенное неоновое свечение
      ctx.fillStyle = `rgba(255, 0, 150, 0.3)`
      ctx.beginPath()
      ctx.arc(x, y, 25, 0, Math.PI * 2)
      ctx.fill()
      
      // Рисуем точку с упрощенным эффектом
      ctx.shadowColor = 'rgba(255, 0, 150, 0.8)'
      ctx.shadowBlur = 15
      ctx.fillStyle = '#ff00ff'
      ctx.beginPath()
      ctx.arc(x, y, pointSize + 1, 0, Math.PI * 2)
      ctx.fill()
      
      // Яркое ядро точки
      ctx.shadowBlur = 0
      ctx.fillStyle = '#ffffff'
      ctx.beginPath()
      ctx.arc(x, y, pointSize / 2, 0, Math.PI * 2)
      ctx.fill()
      
      // Продолжаем анимацию отрисовки
      animationId = requestAnimationFrame(draw)
    }
    
    animationId = requestAnimationFrame(draw)
    
    return () => {
      if (animationId) {
        cancelAnimationFrame(animationId)
      }
    }
  }, [displayPoint, pointSize])

  return (
    <div className="app">
      <div className="container">
        <header className="header">
          <h1>🎯 WebSocket Point Control</h1>
          <p className="subtitle">Управление точкой в реальном времени</p>
        </header>

        <div className="canvas-wrapper">
          <canvas 
            ref={canvasRef}
            width={800} 
            height={600}
            className="canvas"
          />
        </div>

        <div className="info-panel">
          <div className="status-card">
            <div className="status-header">
              <span className="status-label">Статус подключения</span>
              <div className={`status-indicator ${isConnected ? 'connected' : 'disconnected'}`}>
                <span className="status-dot"></span>
                {connectionStatus}
              </div>
            </div>
          </div>

          <div className="coordinates-card">
            <div className="coordinates-header">
              <span className="coordinates-label">Координаты</span>
            </div>
            <div className="coordinates-values">
              <div className="coordinate-item">
                <span className="coordinate-label">X:</span>
                <span className="coordinate-value">{Math.round(displayPoint.x)}</span>
              </div>
              <div className="coordinate-item">
                <span className="coordinate-label">Y:</span>
                <span className="coordinate-value">{Math.round(displayPoint.y)}</span>
              </div>
            </div>
          </div>
        </div>

        <div className="controls-card">
          <h3 className="controls-title">⌨️ Управление</h3>
          <div className="controls-list">
            <div className="control-item">
              <kbd>↑</kbd>
              <span>Переместить вверх</span>
            </div>
            <div className="control-item">
              <kbd>↓</kbd>
              <span>Переместить вниз</span>
            </div>
            <div className="control-item">
              <kbd>←</kbd>
              <span>Переместить влево</span>
            </div>
            <div className="control-item">
              <kbd>→</kbd>
              <span>Переместить вправо</span>
            </div>
          </div>
          <p className="controls-note">
            Точка также может управляться с сервера через WebSocket
          </p>
        </div>
      </div>
    </div>
  )
}

export default App

