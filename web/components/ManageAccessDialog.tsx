"use client";

import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Loader2, Lock, Trash2, UserPlus } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  type AclResourceType,
  type Grant,
  type Group,
  createGrant,
  deleteGrant,
  listGrants,
  listGroups,
  restrictDocument,
  restrictFolder,
  unrestrictDocument,
  unrestrictFolder,
} from "@/lib/api/acl";
import { listMembers, type Member } from "@/lib/api/admin";

interface ManageAccessDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  resourceType: AclResourceType;
  resourceId: string;
  resourceTitle: string;
  initialRestricted?: boolean;
}

export function ManageAccessDialog({
  open,
  onOpenChange,
  resourceType,
  resourceId,
  resourceTitle,
  initialRestricted = false,
}: ManageAccessDialogProps) {
  const t = useTranslations("access");
  const [restricted, setRestricted] = useState(initialRestricted);
  const [grants, setGrants] = useState<Grant[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);
  const [principal, setPrincipal] = useState<string>("");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [grantRes, memberRes, groupRes] = await Promise.all([
        listGrants(resourceType, resourceId),
        listMembers(),
        listGroups(),
      ]);
      setGrants(grantRes.grants);
      setMembers(memberRes.members);
      setGroups(groupRes.groups);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("loadError"));
    } finally {
      setLoading(false);
    }
  }, [resourceType, resourceId, t]);

  useEffect(() => {
    if (open) {
      setRestricted(initialRestricted);
      setPrincipal("");
      void load();
    }
  }, [open, initialRestricted, load]);

  const handleToggleRestricted = async (next: boolean) => {
    setBusy(true);
    try {
      if (resourceType === "document") {
        await (next ? restrictDocument(resourceId) : unrestrictDocument(resourceId));
      } else {
        await (next ? restrictFolder(resourceId) : unrestrictFolder(resourceId));
      }
      setRestricted(next);
      toast.success(next ? t("restrictedOn") : t("restrictedOff"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("updateError"));
    } finally {
      setBusy(false);
    }
  };

  const handleAddGrant = async () => {
    if (!principal) return;
    const [principalType, principalId] = principal.split(":") as [
      "user" | "group",
      string,
    ];
    setBusy(true);
    try {
      await createGrant({
        resource_type: resourceType,
        resource_id: resourceId,
        principal_type: principalType,
        principal_id: principalId,
        permission: "read",
      });
      setPrincipal("");
      toast.success(t("grantAdded"));
      await load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("grantError"));
    } finally {
      setBusy(false);
    }
  };

  const handleRemoveGrant = async (grantId: string) => {
    setBusy(true);
    try {
      await deleteGrant(grantId);
      toast.success(t("grantRemoved"));
      await load();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t("grantError"));
    } finally {
      setBusy(false);
    }
  };

  const principalLabel = (grant: Grant): string => {
    if (grant.PrincipalType === "group") {
      const group = groups.find((g) => g.ID === grant.PrincipalID);
      return group ? group.Name : grant.PrincipalID;
    }
    const member = members.find((m) => m.user_id === grant.PrincipalID);
    return member ? member.display_name || member.email : grant.PrincipalID;
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Lock className="size-5" />
            {t("title")}
          </DialogTitle>
          <DialogDescription>{resourceTitle}</DialogDescription>
        </DialogHeader>

        <div className="space-y-5">
          <div className="flex items-center justify-between rounded-lg border p-3">
            <div className="space-y-0.5">
              <p className="text-sm font-medium">{t("restricted")}</p>
              <p className="text-xs text-muted-foreground">
                {t("restrictedHint")}
              </p>
            </div>
            <Button
              variant={restricted ? "default" : "outline"}
              size="sm"
              disabled={busy}
              onClick={() => void handleToggleRestricted(!restricted)}
            >
              {restricted ? t("on") : t("off")}
            </Button>
          </div>

          {restricted && (
            <div className="space-y-3">
              <div className="flex items-end gap-2">
                <div className="flex-1 space-y-1">
                  <label className="text-xs text-muted-foreground">
                    {t("addGrant")}
                  </label>
                  <Select value={principal} onValueChange={setPrincipal}>
                    <SelectTrigger>
                      <SelectValue placeholder={t("selectPrincipal")} />
                    </SelectTrigger>
                    <SelectContent>
                      {members.map((m) => (
                        <SelectItem key={`user:${m.user_id}`} value={`user:${m.user_id}`}>
                          {(m.display_name || m.email) + ` · ${t("user")}`}
                        </SelectItem>
                      ))}
                      {groups.map((g) => (
                        <SelectItem key={`group:${g.ID}`} value={`group:${g.ID}`}>
                          {g.Name + ` · ${t("group")}`}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <Button
                  size="sm"
                  className="gap-1.5"
                  disabled={busy || !principal}
                  onClick={() => void handleAddGrant()}
                >
                  <UserPlus className="size-4" />
                  {t("add")}
                </Button>
              </div>

              {loading ? (
                <div className="flex h-20 items-center justify-center">
                  <Loader2 className="size-5 animate-spin text-muted-foreground" />
                </div>
              ) : grants.length === 0 ? (
                <p className="py-4 text-center text-sm text-muted-foreground">
                  {t("noGrants")}
                </p>
              ) : (
                <ul className="space-y-2">
                  {grants.map((grant) => (
                    <li
                      key={grant.ID}
                      className="flex items-center justify-between rounded-lg border p-2.5"
                    >
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium">
                          {principalLabel(grant)}
                        </span>
                        <Badge variant="secondary">
                          {t(grant.PrincipalType)}
                        </Badge>
                        <Badge variant="outline">{t(grant.Permission)}</Badge>
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        disabled={busy}
                        onClick={() => void handleRemoveGrant(grant.ID)}
                      >
                        <Trash2 className="size-4 text-destructive" />
                      </Button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
