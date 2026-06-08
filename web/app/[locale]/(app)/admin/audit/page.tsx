"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Loader2 } from "lucide-react";
import { useAuditLog } from "@/features/admin/useAuditLog";

function formatTimestamp(timestamp: string) {
  const date = new Date(timestamp);
  return date.toLocaleString();
}

function getActionVariant(action: string) {
  switch (action) {
    case "upload":
      return "default" as const;
    case "delete":
      return "destructive" as const;
    case "create":
      return "default" as const;
    case "update":
      return "secondary" as const;
    default:
      return "outline" as const;
  }
}

export default function AuditLogPage() {
  const t = useTranslations("admin");
  const { events, loading, error, filters, setFilter } = useAuditLog();

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
      <h1 className="text-3xl font-bold">{t("auditLog")}</h1>

      <div className="flex gap-4">
        <Select
          value={filters.entity_type}
          onValueChange={(value) => setFilter("entity_type", value)}
        >
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder={t("entity")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t("all")}</SelectItem>
            <SelectItem value="document">{t("document")}</SelectItem>
            <SelectItem value="reminder">{t("reminder")}</SelectItem>
            <SelectItem value="user">{t("user")}</SelectItem>
          </SelectContent>
        </Select>

        <Select
          value={filters.action}
          onValueChange={(value) => setFilter("action", value)}
        >
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder={t("action")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t("all")}</SelectItem>
            <SelectItem value="upload">{t("upload")}</SelectItem>
            <SelectItem value="delete">{t("delete")}</SelectItem>
            <SelectItem value="create">{t("create")}</SelectItem>
            <SelectItem value="update">{t("update")}</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {events.length === 0 ? (
        <Card className="p-8 text-center">
          <p className="text-muted-foreground">{t("noAuditEvents")}</p>
        </Card>
      ) : (
        <Card>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("timestamp")}</TableHead>
                <TableHead>{t("entity")}</TableHead>
                <TableHead>{t("action")}</TableHead>
                <TableHead>{t("actor")}</TableHead>
                <TableHead>{t("details")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {events.map((event) => (
                <TableRow key={event.id}>
                  <TableCell className="whitespace-nowrap">
                    {formatTimestamp(event.created_at)}
                  </TableCell>
                  <TableCell>
                    <span className="font-medium capitalize">
                      {t(event.entity_type)}
                    </span>
                    <br />
                    <span className="text-xs text-muted-foreground">
                      {event.entity_id}
                    </span>
                  </TableCell>
                  <TableCell>
                    <Badge variant={getActionVariant(event.action)}>
                      {t(event.action)}
                    </Badge>
                  </TableCell>
                  <TableCell>{event.actor_id || t("system")}</TableCell>
                  <TableCell>
                    {event.metadata && (
                      <code className="rounded bg-muted px-1 py-0.5 text-xs">
                        {JSON.stringify(event.metadata)}
                      </code>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Card>
      )}
    </div>
  );
}
