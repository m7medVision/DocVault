"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuth } from "@/lib/useAuth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

type Mode = "login" | "register";

interface AuthScreenProps {
  mode: Mode;
}

function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function validatePassword(password: string): string | null {
  if (password.length < 8) return "passwordLength";
  if (!/[A-Z]/.test(password)) return "passwordUppercase";
  if (!/[a-z]/.test(password)) return "passwordLowercase";
  if (!/[0-9]/.test(password)) return "passwordNumber";
  if (!/[^A-Za-z0-9]/.test(password)) return "passwordSpecial";
  return null;
}

function getSafeRedirect(redirect: string | null, locale: string): string {
  if (!redirect || !redirect.startsWith("/")) {
    return `/${locale}`;
  }
  return redirect;
}

export default function AuthScreen({ mode }: AuthScreenProps) {
  const locale = useLocale();
  const router = useRouter();
  const searchParams = useSearchParams();
  const t = useTranslations("auth");
  const tCommon = useTranslations("common");
  const { isAuthenticated, isLoading, login, register } = useAuth();
  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [rememberMe, setRememberMe] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const redirectTo = useMemo(
    () => getSafeRedirect(searchParams.get("redirect"), locale),
    [locale, searchParams]
  );
  const queryString = searchParams.toString();

  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.replace(redirectTo);
    }
  }, [isAuthenticated, isLoading, redirectTo, router]);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError(null);

    if (!isValidEmail(email)) {
      setError(t("invalidEmail"));
      return;
    }

    const passwordErrorKey = validatePassword(password);
    if (passwordErrorKey) {
      setError(t(passwordErrorKey));
      return;
    }

    if (mode === "register") {
      if (!displayName.trim()) {
        setError(t("displayNameRequired"));
        return;
      }

      if (password !== confirmPassword) {
        setError(t("passwordMismatch"));
        return;
      }
    }

    try {
      setIsSubmitting(true);

      if (mode === "login") {
        await login({
          email: email.trim(),
          password,
          rememberMe,
        });
      } else {
        await register({
          displayName: displayName.trim(),
          email: email.trim(),
          password,
          locale: locale === "ar" ? "ar" : "en",
          rememberMe,
        });
      }

      router.replace(redirectTo);
      router.refresh();
    } catch (submitError) {
      setError(
        submitError instanceof Error ? submitError.message : t("genericError")
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const switchHref =
    mode === "login"
      ? `/${locale}/auth/register${queryString ? `?${queryString}` : ""}`
      : `/${locale}/auth/login${queryString ? `?${queryString}` : ""}`;

  const spotlightItems = [
    t("spotlightSync"),
    t("spotlightSearch"),
    t("spotlightShare"),
  ];

  return (
    <div className="grid min-h-[calc(100vh-4rem)] lg:grid-cols-2 lg:gap-8">
      <div className="hidden flex-col justify-end bg-primary p-8 lg:flex lg:rounded-2xl">
        <span className="mb-4 inline-flex w-fit rounded-full bg-white/14 px-3 py-1 text-sm font-medium text-white">
          {t("platformLabel")}
        </span>
        <h1 className="mb-4 font-serif text-4xl text-white">
          {mode === "login" ? t("welcomeTitle") : t("registerTitle")}
        </h1>
        <p className="mb-6 max-w-md text-white/85">
          {mode === "login" ? t("welcomeBody") : t("registerBody")}
        </p>
        <ul className="grid gap-3">
          {spotlightItems.map((item) => (
            <li
              key={item}
              className="rounded-xl bg-white/12 p-4 text-white backdrop-blur"
            >
              {item}
            </li>
          ))}
        </ul>
      </div>

      <Card className="m-auto w-full max-w-md p-6">
        <CardHeader className="space-y-1">
          <div className="mb-2 inline-flex rounded-full bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
            {mode === "login" ? t("signIn") : t("createAccount")}
          </div>
          <CardTitle className="text-2xl">
            {mode === "login" ? t("signInTitle") : t("createAccountTitle")}
          </CardTitle>
          <CardDescription>
            {mode === "login" ? t("signInHint") : t("registerHint")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {mode === "register" && (
              <div className="space-y-2">
                <Label htmlFor="displayName">{t("displayName")}</Label>
                <Input
                  id="displayName"
                  autoComplete="name"
                  disabled={isSubmitting}
                  name="displayName"
                  onChange={(event) => setDisplayName(event.target.value)}
                  placeholder={t("displayNamePlaceholder")}
                  type="text"
                  value={displayName}
                />
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="email">{t("email")}</Label>
              <Input
                id="email"
                autoComplete="email"
                disabled={isSubmitting}
                inputMode="email"
                name="email"
                onChange={(event) => setEmail(event.target.value)}
                placeholder="name@company.com"
                type="email"
                value={email}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">{t("password")}</Label>
              <Input
                id="password"
                autoComplete={mode === "login" ? "current-password" : "new-password"}
                disabled={isSubmitting}
                name="password"
                onChange={(event) => setPassword(event.target.value)}
                placeholder="••••••••"
                type="password"
                value={password}
              />
            </div>

            {mode === "register" && (
              <div className="space-y-2">
                <Label htmlFor="confirmPassword">{t("confirmPassword")}</Label>
                <Input
                  id="confirmPassword"
                  autoComplete="new-password"
                  disabled={isSubmitting}
                  name="confirmPassword"
                  onChange={(event) => setConfirmPassword(event.target.value)}
                  placeholder="••••••••"
                  type="password"
                  value={confirmPassword}
                />
              </div>
            )}

            <div className="flex items-center space-x-2">
              <Checkbox
                id="rememberMe"
                checked={rememberMe}
                disabled={isSubmitting}
                onCheckedChange={(checked) =>
                  setRememberMe(checked as boolean)
                }
              />
              <Label htmlFor="rememberMe" className="text-sm font-normal">
                {t("rememberMe")}
              </Label>
            </div>

            <p className="text-sm text-muted-foreground">{t("passwordHint")}</p>

            {error && (
              <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
                {error}
              </div>
            )}

            <Button className="w-full" disabled={isSubmitting} type="submit">
              {isSubmitting
                ? mode === "login"
                  ? t("signingIn")
                  : t("creatingAccount")
                : mode === "login"
                  ? t("signIn")
                  : t("createAccount")}
            </Button>
          </form>

          <p className="mt-4 text-center text-sm text-muted-foreground">
            {mode === "login" ? t("needAccount") : t("haveAccount")}{" "}
            <Link
              href={switchHref}
              className="font-medium text-primary hover:underline"
            >
              {mode === "login" ? t("createAccount") : t("signIn")}
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
