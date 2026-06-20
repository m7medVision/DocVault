"use client"

import { motion, useScroll, useSpring, type MotionProps } from "motion/react"

import { cn } from "@/lib/utils"

interface ScrollProgressProps
  extends Omit<React.HTMLAttributes<HTMLElement>, keyof MotionProps> {
  ref?: React.Ref<HTMLDivElement>
}

export function ScrollProgress({
  className,
  ref,
  ...props
}: ScrollProgressProps) {
  const { scrollYProgress } = useScroll()
  const scaleX = useSpring(scrollYProgress, {
    stiffness: 120,
    damping: 30,
    mass: 0.3,
  })

  return (
    <motion.div
      ref={ref}
      aria-hidden="true"
      className={cn(
        "fixed inset-x-0 top-0 z-[60] h-[2px] origin-left bg-primary rtl:origin-right",
        className,
      )}
      style={{
        scaleX,
      }}
      {...props}
    />
  )
}
