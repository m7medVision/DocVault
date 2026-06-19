"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import {
  Home,
  Files,
  Upload,
  Search,
  MessageCircleQuestion,
  Bell,
  Settings,
  BarChart3,
  Inbox,
  Tags,
  Activity,
  Shield,
  Users,
  ScrollText,
  ChevronDown,
} from "lucide-react";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarMenuSub,
  SidebarMenuSubItem,
  SidebarMenuSubButton,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
} from "@/components/ui/sidebar";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";

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
    { title: tNav("home"), url: `${localePrefix}/dashboard`, icon: Home },
    { title: tNav("documents"), url: `${localePrefix}/documents`, icon: Files },
    { title: tNav("upload"), url: `${localePrefix}/documents/upload`, icon: Upload },
    { title: tNav("search"), url: `${localePrefix}/search`, icon: Search },
    { title: tNav("ask"), url: `${localePrefix}/ask`, icon: MessageCircleQuestion },
    { title: tNav("tags"), url: `${localePrefix}/tags`, icon: Tags },
    { title: tNav("reminders"), url: `${localePrefix}/reminders`, icon: Bell },
    { title: tNav("notifications"), url: `${localePrefix}/notifications`, icon: Inbox },
    { title: tNav("activity"), url: `${localePrefix}/activity`, icon: Activity },
    { title: tNav("analytics"), url: `${localePrefix}/analytics`, icon: BarChart3 },
    { title: tNav("settings"), url: `${localePrefix}/settings`, icon: Settings },
  ];

  const isAdminActive = pathname.startsWith(`${localePrefix}/admin`);
  const [adminOpen, setAdminOpen] = useState(isAdminActive);

  useEffect(() => {
    if (isAdminActive) setAdminOpen(true);
  }, [isAdminActive]);

  const adminItems = [
    { title: tNav("adminMembers"), url: `${localePrefix}/admin/members`, icon: Users },
    { title: tNav("adminGroups"), url: `${localePrefix}/admin/groups`, icon: Users },
    { title: tNav("adminPolicies"), url: `${localePrefix}/admin/casbin`, icon: Shield },
    { title: tNav("adminAudit"), url: `${localePrefix}/admin/audit`, icon: ScrollText },
  ];

  return (
    <Sidebar side={isRtl ? 'right' : 'left'}>
      <SidebarHeader className="h-16 flex items-center border-b px-5">
        <Link href={`${localePrefix}/dashboard`} className="flex items-center gap-3 leading-none">
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
          <SidebarMenu>
            <Collapsible
              open={adminOpen}
              onOpenChange={setAdminOpen}
              className="group/collapsible"
            >
              <SidebarMenuItem>
                <CollapsibleTrigger asChild>
                  <SidebarMenuButton isActive={isAdminActive}>
                    <Shield />
                    <span>{tNav("admin")}</span>
                    <ChevronDown className="ms-auto size-4 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-180 rtl:group-data-[state=open]/collapsible:-rotate-180" />
                  </SidebarMenuButton>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <SidebarMenuSub>
                    {adminItems.map((item) => (
                      <SidebarMenuSubItem key={item.url}>
                        <SidebarMenuSubButton asChild isActive={pathname === item.url}>
                          <Link href={item.url}>
                            <item.icon />
                            <span>{item.title}</span>
                          </Link>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                    ))}
                  </SidebarMenuSub>
                </CollapsibleContent>
              </SidebarMenuItem>
            </Collapsible>
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter className="border-t p-4" />
    </Sidebar>
  );
}
