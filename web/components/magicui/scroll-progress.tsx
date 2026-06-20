"use client";

import { motion, useScroll, useSpring } from "framer-motion";

/**
 * Fixed top scroll-progress indicator (magicui-style).
 * Thin ink-blue bar tracking page scroll. origin-left in LTR,
 * flipped to origin-right in RTL via the rtl: variant.
 */
export function ScrollProgress() {
  const { scrollYProgress } = useScroll();
  const scaleX = useSpring(scrollYProgress, {
    stiffness: 120,
    damping: 30,
    mass: 0.3,
  });

  return (
    <motion.div
      aria-hidden="true"
      style={{ scaleX }}
      className="fixed inset-x-0 top-0 z-[60] h-[2px] origin-left bg-primary rtl:origin-right"
    />
  );
}
