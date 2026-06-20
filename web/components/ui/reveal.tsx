"use client";

import * as React from "react";
import { motion, useReducedMotion } from "motion/react";

import { cn } from "@/lib/utils";

const EASE_OUT: [number, number, number, number] = [0.22, 1, 0.36, 1];

type RevealProps = {
  children: React.ReactNode;
  className?: string;
  /** Stagger delay in seconds. */
  delay?: number;
  /** Reveal duration in seconds. */
  duration?: number;
  /** Vertical travel distance in px. */
  offset?: number;
  /** Scroll-triggered (default) vs animate on mount (above-the-fold). */
  inView?: boolean;
  /** Animate only the first time it enters the viewport. */
  once?: boolean;
  /** Bottom viewport margin before the reveal fires. */
  margin?: string;
} & Omit<React.ComponentProps<typeof motion.div>, "ref" | "variants">;

export function Reveal({
  children,
  className,
  delay = 0,
  duration = 0.6,
  offset = 8,
  inView = true,
  once = true,
  margin = "0px 0px -10% 0px",
  ...rest
}: RevealProps) {
  const reduce = useReducedMotion();

  const variants = {
    hidden: reduce ? { opacity: 1 } : { opacity: 0, y: offset },
    visible: {
      opacity: 1,
      y: 0,
      transition: reduce
        ? { duration: 0 }
        : { duration, ease: EASE_OUT, delay },
    },
  };

  const trigger = inView
    ? {
        initial: "hidden",
        whileInView: "visible",
        viewport: { once, margin },
      }
    : {
        initial: "hidden",
        animate: "visible",
      };

  return (
    <motion.div
      className={cn(className)}
      variants={variants}
      {...trigger}
      {...rest}
    >
      {children}
    </motion.div>
  );
}
