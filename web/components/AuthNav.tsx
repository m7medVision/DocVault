"use client";

import Link from "next/link";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/useAuth";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";

export default function AuthNav() {
  const locale = useLocale();
  const router = useRouter();
  const tAuth = useTranslations("auth");
  const { isAuthenticated, isLoading, logout, user } = useAuth();

  const handleLogout = async () => {
    await logout();
    router.replace(`/${locale}/auth/login`);
    router.refresh();
  };

  if (isLoading) {
    return (
      <div className="h-10 w-24 animate-pulse rounded-full bg-muted" />
    );
  }

  if (!isAuthenticated || !user) {
    return (
      <div className="flex gap-2">
        <Button variant="outline" size="sm" asChild>
          <Link href={`/${locale}/auth/login`}>{tAuth("signIn")}</Link>
        </Button>
        <Button size="sm" asChild>
          <Link href={`/${locale}/auth/register`}>{tAuth("createAccount")}</Link>
        </Button>
      </div>
    );
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <span className="font-semibold">{user.displayName}</span>
          {user.role && (
            <Badge variant="secondary" className="text-xs">
              {user.role}
            </Badge>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>{tAuth("profile")}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleLogout} className="text-destructive">
          {tAuth("signOut")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
