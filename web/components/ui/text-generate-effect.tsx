"use client";

import { useEffect } from "react";
import {
  motion,
  stagger,
  useAnimate,
  useReducedMotion,
} from "motion/react";

import { cn } from "@/lib/utils";

export const TextGenerateEffect = ({
  words,
  className,
  filter = true,
  duration = 0.5,
  staggerAmount = 0.08,
}: {
  words: string;
  className?: string;
  filter?: boolean;
  duration?: number;
  staggerAmount?: number;
}) => {
  const [scope, animate] = useAnimate();
  const reduce = useReducedMotion();
  const wordsArray = words.split(" ");

  useEffect(() => {
    if (reduce) return;
    animate(
      ".dv-tg-word",
      {
        opacity: 1,
        filter: filter ? "blur(0px)" : "none",
      },
      {
        duration: duration ?? 0.5,
        delay: stagger(staggerAmount),
      },
    );
  }, [scope.current, reduce, animate, filter, duration, staggerAmount]);

  if (reduce) {
    return <span className={cn("inline", className)}>{words}</span>;
  }

  return (
    <motion.span ref={scope} className={cn("inline", className)}>
      {wordsArray.map((word, idx) => (
        <motion.span
          key={`${word}-${idx}`}
          aria-hidden="false"
          className="dv-tg-word inline-block opacity-0"
          style={{
            filter: filter ? "blur(6px)" : "none",
          }}
        >
          {word}
          {idx < wordsArray.length - 1 ? "\u00A0" : ""}
        </motion.span>
      ))}
    </motion.span>
  );
};
