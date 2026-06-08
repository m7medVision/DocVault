"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { useAuth } from "@/lib/useAuth";
import { updateEmail, updatePassword, updateProfile } from "@/lib/api/profile";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default function SettingsPage() {
  const t = useTranslations("settings");
  const tCommon = useTranslations("common");
  const { user, updateUser } = useAuth();

  return (
    <div className="mx-auto max-w-2xl space-y-8">
      <div>
        <h1 className="text-3xl font-bold">{t("title")}</h1>
        <p className="mt-1 text-muted-foreground">{t("description")}</p>
      </div>

      <ProfileSection user={user} updateUser={updateUser} t={t} tCommon={tCommon} />
      <EmailSection user={user} updateUser={updateUser} t={t} tCommon={tCommon} />
      <PasswordSection t={t} tCommon={tCommon} />
    </div>
  );
}

function ProfileSection({
  user,
  updateUser,
  t,
  tCommon,
}: {
  user: ReturnType<typeof useAuth>["user"];
  updateUser: ReturnType<typeof useAuth>["updateUser"];
  t: ReturnType<typeof useTranslations>;
  tCommon: ReturnType<typeof useTranslations>;
}) {
  const [displayName, setDisplayName] = useState(user?.displayName ?? "");
  const [locale, setLocale] = useState<string>(user?.locale ?? "en");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!displayName.trim()) return;

    setIsSubmitting(true);
    try {
      const updated = await updateProfile({
        display_name: displayName.trim(),
        locale,
      });
      updateUser({
        displayName: updated.display_name,
        locale: updated.locale,
      });
      toast.success(t("profileUpdated"));
    } catch (err) {
      toast.error(tCommon("error"));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("profile")}</CardTitle>
        <CardDescription>{t("profileDescription")}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="displayName">{t("displayName")}</Label>
            <Input
              id="displayName"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder={t("displayNamePlaceholder")}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="locale">{t("locale")}</Label>
            <Select value={locale} onValueChange={setLocale}>
              <SelectTrigger id="locale">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="en">{t("localeEnglish")}</SelectItem>
                <SelectItem value="ar">{t("localeArabic")}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? tCommon("loading") : t("updateProfile")}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function EmailSection({
  user,
  updateUser,
  t,
  tCommon,
}: {
  user: ReturnType<typeof useAuth>["user"];
  updateUser: ReturnType<typeof useAuth>["updateUser"];
  t: ReturnType<typeof useTranslations>;
  tCommon: ReturnType<typeof useTranslations>;
}) {
  const [email, setEmail] = useState(user?.email ?? "");
  const [currentPassword, setCurrentPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim() || !currentPassword) return;

    setIsSubmitting(true);
    try {
      const updated = await updateEmail({
        email: email.trim().toLowerCase(),
        current_password: currentPassword,
      });
      updateUser({
        email: updated.email,
        emailVerified: updated.email_verified,
      });
      setCurrentPassword("");
      toast.success(t("emailUpdated"));
    } catch (err) {
      const msg = err instanceof Error ? err.message : null;
      toast.error(msg || tCommon("error"));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("email")}</CardTitle>
        <CardDescription>{t("emailDescription")}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email">{t("newEmail")}</Label>
            <Input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder={t("newEmailPlaceholder")}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="emailCurrentPassword">{t("currentPassword")}</Label>
            <Input
              id="emailCurrentPassword"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              placeholder={t("currentPasswordPlaceholder")}
              required
            />
          </div>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? tCommon("loading") : t("updateEmail")}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function PasswordSection({
  t,
  tCommon,
}: {
  t: ReturnType<typeof useTranslations>;
  tCommon: ReturnType<typeof useTranslations>;
}) {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentPassword || !newPassword) return;

    setIsSubmitting(true);
    try {
      await updatePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });
      setCurrentPassword("");
      setNewPassword("");
      toast.success(t("passwordUpdated"));
    } catch (err) {
      const msg = err instanceof Error ? err.message : null;
      toast.error(msg || tCommon("error"));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("password")}</CardTitle>
        <CardDescription>{t("passwordDescription")}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="currentPassword">{t("currentPassword")}</Label>
            <Input
              id="currentPassword"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              placeholder={t("currentPasswordPlaceholder")}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="newPassword">{t("newPassword")}</Label>
            <Input
              id="newPassword"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder={t("newPasswordPlaceholder")}
              required
            />
          </div>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? tCommon("loading") : t("updatePassword")}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}