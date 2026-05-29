"use client";

import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { 
  Home, 
  Files, 
  Upload, 
  Search, 
  Bell, 
  Settings 
} from "lucide-react";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
} from "@/components/ui/sidebar";

export function AppSidebar() {
  const tNav = useTranslations("nav");
  const tCommon = useTranslations("common");
  const pathname = usePathname();
  const isRtl = pathname.split('/')[1] === 'ar';
  
  const getLocalePrefix = () => {
    const segments = pathname.split('/');
    return segments[1] && ['en', 'ar'].includes(segments[1]) ? `/${segments[1]}` : '/en';
  };
  
  const localePrefix = getLocalePrefix();

  const navItems = [
    { title: tNav("home"), url: `${localePrefix}`, icon: Home },
    { title: tNav("documents"), url: `${localePrefix}/documents`, icon: Files },
    { title: tNav("upload"), url: `${localePrefix}/documents/upload`, icon: Upload },
    { title: tNav("search"), url: `${localePrefix}/search`, icon: Search },
    { title: tNav("reminders"), url: `${localePrefix}/reminders`, icon: Bell },
    { title: tNav("settings"), url: `${localePrefix}/settings`, icon: Settings },
  ];

  return (
    <Sidebar side={isRtl ? 'right' : 'left'}>
      <SidebarHeader className="h-16 flex items-center border-b px-5">
        <Link href={`${localePrefix}`} className="flex items-center gap-3 leading-none">
          <Image src="/logo.png" alt="" aria-hidden="true" width={34} height={34} />
          <span className="inline-flex items-center text-lg font-bold leading-none tracking-tight text-sidebar-foreground">
            {tCommon("appName")}
          </span>
        </Link>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>{tNav("mainMenu")}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => {
                const isActive = pathname === item.url || (item.url !== localePrefix && pathname.startsWith(item.url));
                return (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton asChild isActive={isActive}>
                      <Link href={item.url}>
                        <item.icon />
                        <span>{item.title}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        
        <SidebarGroup className="mt-auto">
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild isActive={pathname.startsWith(`${localePrefix}/admin`)}>
                  <Link href={`${localePrefix}/admin`}>
                    <Settings />
                    <span>{tNav("admin")}</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter className="border-t p-4" />
    </Sidebar>
  );
}
