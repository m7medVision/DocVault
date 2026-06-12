"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Loader2 } from "lucide-react";
import { useMembers } from "@/features/admin/useMembers";

function formatDate(dateString: string) {
  const date = new Date(dateString);
  return date.toLocaleDateString();
}

function getRoleVariant(role: string) {
  switch (role) {
    case "admin":
      return "destructive" as const;
    case "member":
      return "default" as const;
    case "viewer":
      return "secondary" as const;
    default:
      return "outline" as const;
  }
}

function getStatusVariant(status: string) {
  switch (status) {
    case "active":
      return "default" as const;
    case "invited":
      return "secondary" as const;
    case "disabled":
      return "outline" as const;
    default:
      return "outline" as const;
  }
}

export default function MembersPage() {
  const t = useTranslations("admin");
  const pathname = usePathname();
  const localePrefix = pathname.split("/")[1] || "en";
  const {
    members,
    loading,
    error,
    filter,
    setFilter,
    handleInvite,
    handleEditRole,
  } = useMembers();

  const filteredMembers = members.filter((member) => {
    if (filter === "all") return true;
    return member.role === filter;
  });

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg bg-destructive/10 p-4 text-center text-destructive">
        {error}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">{t("workspaceMembers")}</h1>
        <Button onClick={() => void handleInvite()}>{t("inviteMember")}</Button>
      </div>

      <div className="flex items-center gap-4">
        <Select value={filter} onValueChange={setFilter}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder={t("allRoles")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t("allRoles")}</SelectItem>
            <SelectItem value="admin">{t("admins")}</SelectItem>
            <SelectItem value="member">{t("members")}</SelectItem>
            <SelectItem value="viewer">{t("viewers")}</SelectItem>
          </SelectContent>
        </Select>

        <span className="text-sm text-muted-foreground">
          {filteredMembers.length}{" "}
          {filteredMembers.length === 1 ? t("member") : t("members")}
        </span>
      </div>

      {filteredMembers.length === 0 ? (
        <Card className="p-8 text-center">
          <p className="text-muted-foreground">{t("noMembersFound")}</p>
        </Card>
      ) : (
        <div className="space-y-3">
          {filteredMembers.map((member) => (
            <Card key={member.id} className="p-4">
              <div className="flex items-center gap-4">
                <Avatar className="h-10 w-10">
                  <AvatarFallback className="bg-primary text-primary-foreground">
                    {member.name.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>

                <div className="flex-1 space-y-0.5">
                  <p className="font-medium">{member.name}</p>
                  <p className="text-sm text-muted-foreground">{member.email}</p>
                </div>

                <div className="flex items-center gap-2">
                  <Badge variant={getRoleVariant(member.role)}>
                    {t(member.role)}
                  </Badge>
                  <Badge variant={getStatusVariant(member.status)}>
                    {t(member.status)}
                  </Badge>
                  <span className="text-xs text-muted-foreground">
                    {t("joined")} {formatDate(member.joined_at)}
                  </span>
                </div>

                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => void handleEditRole(member)}
                  >
                    {t("edit")}
                  </Button>
                  <Button variant="ghost" size="sm" asChild>
                    <Link href={`/${localePrefix}/admin/members/${member.id}`}>
                      {t("viewPermissions")}
                    </Link>
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
