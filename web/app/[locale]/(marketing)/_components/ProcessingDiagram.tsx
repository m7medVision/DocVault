"use client";

import { forwardRef, useRef, type ComponentType } from "react";
import { useTranslations } from "next-intl";
import { useReducedMotion } from "motion/react";
import {
  CalendarClock,
  FileUp,
  Languages,
  ScanText,
  Search,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { AnimatedBeam } from "@/components/ui/animated-beam";

type NodeProps = {
  icon: ComponentType<{ className?: string; "aria-hidden"?: boolean }>;
  label: string;
  /** Larger central hub tile. */
  hub?: boolean;
  /** Amber tint, reserved for the dates/reminders node. */
  amber?: boolean;
};

const Node = forwardRef<HTMLDivElement, NodeProps>(function Node(
  { icon: Icon, label, hub = false, amber = false },
  ref,
) {
  return (
    <div className="flex flex-col items-center gap-2 text-center">
      <div
        ref={ref}
        className={cn(
          "z-10 flex items-center justify-center rounded-xl border border-border bg-card shadow-sm",
          hub ? "size-16" : "size-14",
        )}
      >
        <Icon
          className={cn(
            hub ? "size-7" : "size-6",
            amber ? "text-amber" : "text-foreground",
          )}
          aria-hidden={true}
        />
      </div>
      <span className="max-w-[7.5rem] text-xs font-medium leading-tight text-muted-foreground">
        {label}
      </span>
    </div>
  );
});

export function ProcessingDiagram() {
  const t = useTranslations("landing.howItWorks.nodes");
  const reduce = useReducedMotion();

  const containerRef = useRef<HTMLDivElement>(null);
  const uploadRef = useRef<HTMLDivElement>(null);
  const hubRef = useRef<HTMLDivElement>(null);
  const translateRef = useRef<HTMLDivElement>(null);
  const datesRef = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLDivElement>(null);

  const beams = [
    { from: uploadRef, to: hubRef, curvature: 0, delay: 0 },
    { from: hubRef, to: translateRef, curvature: 40, delay: 0.6 },
    { from: hubRef, to: datesRef, curvature: 0, delay: 0.9, amber: true },
    { from: hubRef, to: searchRef, curvature: -40, delay: 1.2 },
  ] as const;

  return (
    // Geometric/absolute diagram: render as an LTR island so the beam math and
    // node order stay consistent under dir="rtl" (the surrounding copy is RTL).
    <div
      ref={containerRef}
      dir="ltr"
      className="relative mx-auto flex h-[22rem] w-full max-w-2xl items-stretch justify-between gap-4 px-2 sm:gap-10 sm:px-6"
    >
      <div className="flex flex-col justify-center">
        <Node ref={uploadRef} icon={FileUp} label={t("upload")} />
      </div>

      <div className="flex flex-col justify-center">
        <Node ref={hubRef} icon={ScanText} label={t("process")} hub />
      </div>

      <div className="flex flex-col justify-center gap-8">
        <Node ref={translateRef} icon={Languages} label={t("translate")} />
        <Node ref={datesRef} icon={CalendarClock} label={t("dates")} amber />
        <Node ref={searchRef} icon={Search} label={t("search")} />
      </div>

      {beams.map((beam, i) => (
        <AnimatedBeam
          key={i}
          containerRef={containerRef}
          fromRef={beam.from}
          toRef={beam.to}
          curvature={beam.curvature}
          duration={reduce ? 0 : 3}
          delay={reduce ? 0 : beam.delay}
          repeat={reduce ? 0 : Infinity}
          pathColor="var(--border)"
          gradientStartColor="var(--primary)"
          gradientStopColor={"amber" in beam ? "var(--brand-amber)" : "var(--ring)"}
        />
      ))}
    </div>
  );
}
