"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTranslations } from "next-intl";
import { Loader2, ArrowLeft, UserCog } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useMemberPermissions } from "@/features/admin/useMemberPermissions";

export default function MemberPermissionsPage() {
  const params = useParams();
  const memberId = params.id as string;
  const t = useTranslations("admin");
  const pathname = usePathname();
  const localePrefix = pathname.split("/")[1] || "en";

  const { data, isLoading, error } = useMemberPermissions(memberId);

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="rounded-lg bg-destructive/10 p-4 text-center text-destructive">
        {error instanceof Error ? error.message : "Error"}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Button variant="ghost" size="sm" asChild className="gap-1.5">
        <Link href={`/${localePrefix}/admin/members`}>
          <ArrowLeft className="size-4" />
          {t("members")}
        </Link>
      </Button>

      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-2xl font-bold">
          <UserCog className="size-6" />
          {data.member.display_name || data.member.email}
        </h1>
        <p className="text-sm text-muted-foreground">{t("permissionsSubtitle")}</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("currentRole")}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-wrap gap-2">
            <Badge>{data.current_role}</Badge>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("roles")}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-wrap gap-2">
            {data.roles.length > 0 ? (
              data.roles.map((r) => (
                <Badge key={r} variant="secondary">
                  {r}
                </Badge>
              ))
            ) : (
              <span className="text-sm text-muted-foreground">—</span>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("permissions")}</CardTitle>
        </CardHeader>
        <CardContent>
          {data.permissions.length === 0 ? (
            <p className="text-sm text-muted-foreground">—</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("member")}</TableHead>
                  <TableHead>{t("object")}</TableHead>
                  <TableHead>{t("action")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.permissions.map((p, i) => (
                  <TableRow key={`${p.Role}-${p.Object}-${p.Action}-${i}`}>
                    <TableCell>
                      <Badge variant="secondary">{p.Role}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm">{p.Object}</TableCell>
                    <TableCell className="font-mono text-sm">{p.Action}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
