"use client";

import { useId } from "react";

import { cn } from "@/lib/utils";

type NoiseTextureProps = {
  className?: string;
  /** Fractal noise frequency; higher = finer grain. */
  frequency?: number;
};

/**
 * SVG fractal-noise overlay for a warm-paper / archival grain (magicui-style).
 * Set a low opacity via className (e.g. opacity-[0.04]) and keep it pointer-events-none.
 */
export function NoiseTexture({
  className,
  frequency = 0.9,
}: NoiseTextureProps) {
  const id = useId().replace(/:/g, "");

  return (
    <svg
      aria-hidden="true"
      className={cn("pointer-events-none absolute inset-0 h-full w-full", className)}
    >
      <filter id={`noise-${id}`}>
        <feTurbulence
          type="fractalNoise"
          baseFrequency={frequency}
          numOctaves={2}
          stitchTiles="stitch"
        />
        <feColorMatrix type="saturate" values="0" />
      </filter>
      <rect width="100%" height="100%" filter={`url(#noise-${id})`} />
    </svg>
  );
}
