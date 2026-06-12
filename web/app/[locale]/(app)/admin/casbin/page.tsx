"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Loader2, Plus, Trash2, ShieldCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Card, CardContent } from "@/components/ui/card";
import { usePolicies, usePolicyMutations } from "@/features/admin/useCasbin";

export default function CasbinPage() {
  const t = useTranslations("admin");
  const { data, isLoading } = usePolicies();
  const { add, remove } = usePolicyMutations();

  const [form, setForm] = useState({ subject: "", object: "", action: "" });

  const handleAdd = async () => {
    if (!form.subject || !form.object || !form.action) return;
    try {
      await add.mutateAsync(form);
      toast.success(t("policyAdded"));
      setForm({ subject: "", object: "", action: "" });
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Error");
    }
  };

  const handleRemove = async (subject: string, object: string, action: string) => {
    try {
      await remove.mutateAsync({ subject, object, action });
      toast.success(t("policyRemoved"));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Error");
    }
  };

  const policies = data?.policies ?? [];

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-2xl font-bold">
          <ShieldCheck className="size-6" />
          {t("policies")}
        </h1>
        <p className="text-sm text-muted-foreground">{t("policiesSubtitle")}</p>
      </div>

      <Card>
        <CardContent className="flex flex-wrap items-end gap-3 p-4">
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">{t("subject")}</label>
            <Input
              value={form.subject}
              onChange={(e) => setForm((f) => ({ ...f, subject: e.target.value }))}
              placeholder="role:admin"
              className="w-40"
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">{t("object")}</label>
            <Input
              value={form.object}
              onChange={(e) => setForm((f) => ({ ...f, object: e.target.value }))}
              placeholder="documents"
              className="w-40"
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">{t("action")}</label>
            <Input
              value={form.action}
              onChange={(e) => setForm((f) => ({ ...f, action: e.target.value }))}
              placeholder="read"
              className="w-32"
            />
          </div>
          <Button onClick={handleAdd} disabled={add.isPending} className="gap-1.5">
            {add.isPending ? <Loader2 className="size-4 animate-spin" /> : <Plus className="size-4" />}
            {t("addPolicy")}
          </Button>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex h-32 items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : policies.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("noPolicies")}</p>
      ) : (
        <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("subject")}</TableHead>
                <TableHead>{t("object")}</TableHead>
                <TableHead>{t("action")}</TableHead>
                <TableHead className="w-16" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {policies.map((p, i) => (
                <TableRow key={`${p.subject}-${p.object}-${p.action}-${i}`}>
                  <TableCell>
                    <Badge variant="secondary">{p.subject}</Badge>
                  </TableCell>
                  <TableCell className="font-mono text-sm">{p.object}</TableCell>
                  <TableCell className="font-mono text-sm">{p.action}</TableCell>
                  <TableCell>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleRemove(p.subject, p.object, p.action)}
                      disabled={remove.isPending}
                    >
                      <Trash2 className="size-4 text-destructive" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
