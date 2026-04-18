"use client";

import { Bell, Check, Info } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { useTranslations } from "next-intl";

import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useNotifications } from "@/features/notifications/useNotifications";
import { useMarkRead } from "@/features/notifications/useMarkRead";
import { Skeleton } from "@/components/ui/skeleton";

export function NotificationBell() {
  const t = useTranslations("notifications");
  const { data, isLoading } = useNotifications({ status: "unread", limit: 5 });
  const { mutate: markAsRead, isPending: isMarking } = useMarkRead();

  const unreadCount = data?.unread_count || 0;
  const notifications = data?.notifications || [];

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <Badge 
              variant="destructive" 
              className="absolute -top-1 -right-1 h-5 w-5 flex items-center justify-center p-0 text-xs"
            >
              {unreadCount > 99 ? '99+' : unreadCount}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="end">
        <div className="flex items-center justify-between px-4 py-3 border-b">
          <h4 className="text-sm font-semibold">{t("title")}</h4>
          {unreadCount > 0 && (
            <Badge variant="secondary" className="text-xs font-normal">
              {unreadCount} {t("unread")}
            </Badge>
          )}
        </div>
        
        <div className="h-max-[300px] overflow-y-auto">
          {isLoading ? (
            <div className="p-4 space-y-3">
              {[1, 2, 3].map(i => (
                <div key={i} className="flex gap-3">
                  <Skeleton className="h-8 w-8 rounded-full" />
                  <div className="space-y-2 flex-1">
                    <Skeleton className="h-4 w-full" />
                    <Skeleton className="h-3 w-2/3" />
                  </div>
                </div>
              ))}
            </div>
          ) : notifications.length === 0 ? (
            <div className="p-6 text-center text-sm text-muted-foreground flex flex-col items-center">
              <Bell className="h-8 w-8 mb-2 opacity-20" />
              <p>{t("allCaughtUp")}</p>
            </div>
          ) : (
            <div className="flex flex-col">
              {notifications.map((notif) => (
                <div 
                  key={notif.id} 
                  className="flex items-start gap-3 p-4 border-b last:border-0 hover:bg-muted/50 transition-colors"
                >
                  <div className="mt-0.5 shrink-0 bg-primary/10 p-1.5 rounded-full text-primary">
                    <Info className="h-4 w-4" />
                  </div>
                  <div className="flex-1 space-y-1">
                    <p className="text-sm font-medium leading-none">{notif.title}</p>
                    <p className="text-xs text-muted-foreground line-clamp-2">{notif.body}</p>
                    <p className="text-[10px] text-muted-foreground pt-1">
                      {formatDistanceToNow(new Date(notif.created_at), { addSuffix: true })}
                    </p>
                  </div>
                  <Button 
                    variant="ghost" 
                    size="icon" 
                    className="h-6 w-6 shrink-0 text-muted-foreground hover:text-primary"
                    onClick={() => markAsRead(notif.id)}
                    disabled={isMarking}
                  >
                    <Check className="h-4 w-4" />
                    <span className="sr-only">{t("markAsRead")}</span>
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>
        {notifications.length > 0 && (
          <div className="p-2 border-t">
            <Button variant="outline" className="w-full text-xs h-8" disabled>
              {t("viewAllNotifications")}
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
