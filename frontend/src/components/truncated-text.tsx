import { useRef, useState, useCallback } from 'react'
import { cn } from '@/lib/utils'

interface TruncatedTextProps {
  children: string
  className?: string
}

/**
 * 智能截断文本组件
 * 仅当文本被截断时才显示 title 悬浮提示
 */
export function TruncatedText({ children, className }: TruncatedTextProps) {
  const textRef = useRef<HTMLSpanElement>(null)
  const [title, setTitle] = useState<string | undefined>(undefined)

  // 在鼠标进入时检测是否被截断
  const handleMouseEnter = useCallback(() => {
    const element = textRef.current
    if (!element) return

    // scrollWidth > clientWidth 表示文本被截断了
    if (element.scrollWidth > element.clientWidth) {
      setTitle(children)
    }
  }, [children])

  // 鼠标离开时清除 title
  const handleMouseLeave = useCallback(() => {
    setTitle(undefined)
  }, [])

  return (
    <span
      ref={textRef}
      className={cn('truncate', className)}
      title={title}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {children}
    </span>
  )
}
