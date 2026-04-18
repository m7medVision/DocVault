"use client";

import { useEffect } from "react";

interface HtmlAttributesProps {
  locale: string;
}

export function HtmlAttributes({ locale }: HtmlAttributesProps) {
  useEffect(() => {
    document.documentElement.dir = locale === "ar" ? "rtl" : "ltr";
    document.documentElement.lang = locale;
  }, [locale]);

  return null;
}
