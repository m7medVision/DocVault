"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Loader2, Plus, Trash2, Users, UserPlus, UserMinus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  useGroups,
  useGroupMutations,
  useOrgMembers,
} from "@/features/admin/useGroups";

export default function GroupsPage() {
  const t = useTranslations("groups");
  const { data, isLoading } = useGroups();
  const { data: memberData } = useOrgMembers();
  const { create, remove, addMember, removeMember } = useGroupMutations();

  const [name, setName] = useState("");
  const [selected, setSelected] = useState<Record<string, string>>({});

  const groups = data?.groups ?? [];
  const members = memberData?.members ?? [];

  const handleCreate = async () => {
    const trimmed = name.trim();
    if (!trimmed) return;
    try {
      await create.mutateAsync(trimmed);
      toast.success(t("groupCreated"));
      setName("");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("error"));
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await remove.mutateAsync(id);
      toast.success(t("groupDeleted"));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("error"));
    }
  };

  const handleAddMember = async (groupId: string) => {
    const userId = selected[groupId];
    if (!userId) return;
    try {
      await addMember.mutateAsync({ groupId, userId });
      toast.success(t("memberAdded"));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("error"));
    }
  };

  const handleRemoveMember = async (groupId: string) => {
    const userId = selected[groupId];
    if (!userId) return;
    try {
      await removeMember.mutateAsync({ groupId, userId });
      toast.success(t("memberRemoved"));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("error"));
    }
  };

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-2xl font-bold">
          <Users className="size-6" />
          {t("title")}
        </h1>
        <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
      </div>

      <Card>
        <CardContent className="flex flex-wrap items-end gap-3 p-4">
          <div className="flex-1 space-y-1">
            <label className="text-xs text-muted-foreground">
              {t("groupName")}
            </label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") void handleCreate();
              }}
              placeholder={t("groupNamePlaceholder")}
              className="max-w-xs"
            />
          </div>
          <Button
            onClick={() => void handleCreate()}
            disabled={create.isPending || !name.trim()}
            className="gap-1.5"
          >
            {create.isPending ? (
              <Loader2 className="size-4 animate-spin" />
            ) : (
              <Plus className="size-4" />
            )}
            {t("createGroup")}
          </Button>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex h-32 items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : groups.length === 0 ? (
        <Card className="p-8 text-center">
          <p className="text-sm text-muted-foreground">{t("noGroups")}</p>
        </Card>
      ) : (
        <div className="space-y-3">
          {groups.map((group) => (
            <Card key={group.ID} className="p-4">
              <div className="flex items-center justify-between gap-4">
                <div className="space-y-0.5">
                  <p className="font-medium">{group.Name}</p>
                  <p className="text-xs text-muted-foreground">
                    {t("created")}{" "}
                    {new Date(group.CreatedAt).toLocaleDateString()}
                  </p>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => void handleDelete(group.ID)}
                  disabled={remove.isPending}
                >
                  <Trash2 className="size-4 text-destructive" />
                </Button>
              </div>

              <div className="mt-3 flex flex-wrap items-end gap-2 border-t pt-3">
                <div className="flex-1 space-y-1">
                  <label className="text-xs text-muted-foreground">
                    {t("manageMembers")}
                  </label>
                  <Select
                    value={selected[group.ID] ?? ""}
                    onValueChange={(value) =>
                      setSelected((prev) => ({ ...prev, [group.ID]: value }))
                    }
                  >
                    <SelectTrigger className="max-w-xs">
                      <SelectValue placeholder={t("selectMember")} />
                    </SelectTrigger>
                    <SelectContent>
                      {members.map((m) => (
                        <SelectItem key={m.user_id} value={m.user_id}>
                          {m.display_name || m.email}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  className="gap-1.5"
                  disabled={addMember.isPending || !selected[group.ID]}
                  onClick={() => void handleAddMember(group.ID)}
                >
                  <UserPlus className="size-4" />
                  {t("addMember")}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  className="gap-1.5"
                  disabled={removeMember.isPending || !selected[group.ID]}
                  onClick={() => void handleRemoveMember(group.ID)}
                >
                  <UserMinus className="size-4" />
                  {t("removeMember")}
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
