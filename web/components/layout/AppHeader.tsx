"use client";

import { useTranslations } from "next-intl";
import { Search } from "lucide-react";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";

import { Input } from "@/components/ui/input";
import { SidebarTrigger } from "@/components/ui/sidebar";
import LanguageSwitcher from "@/components/LanguageSwitcher";
import AuthNav from "@/components/AuthNav";
import { NotificationBell } from "@/components/layout/NotificationBell";

export function AppHeader() {
  const tSearch = useTranslations("search");
  const router = useRouter();
  const pathname = usePathname();
  const [searchQuery, setSearchQuery] = useState("");

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!searchQuery.trim()) return;
    
    // Find locale
    const segments = pathname.split('/');
    const locale = segments[1] && ['en', 'ar'].includes(segments[1]) ? segments[1] : 'en';
    
    router.push(`/${locale}/search?q=${encodeURIComponent(searchQuery)}`);
  };

  return (
    <header className="sticky top-0 z-50 flex h-16 shrink-0 items-center gap-4 border-b bg-background/95 px-6 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <SidebarTrigger />
      
      <div className="flex flex-1 items-center gap-4 md:ms-auto md:gap-2 lg:gap-4">
        <form onSubmit={handleSearch} className="ms-auto flex-1 sm:flex-initial">
          <div className="relative">
            <Search className="absolute start-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              type="search"
              placeholder={tSearch("placeholder")}
              className="ps-8 sm:w-[300px] md:w-[200px] lg:w-[300px]"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </form>
        
        <div className="flex items-center gap-2">
          <NotificationBell />
          <LanguageSwitcher />
          <AuthNav />
        </div>
      </div>
    </header>
  );
}
