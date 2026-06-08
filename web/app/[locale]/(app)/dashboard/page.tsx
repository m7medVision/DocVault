"use client";

import { useTranslations } from "next-intl";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Upload, FileText, Search, Bell, Clock, Database, FileStack, Check, Info } from "lucide-react";
import { useDocumentStats } from "@/features/documents/useDocumentStats";
import { useRecentDocuments } from "@/features/documents/useRecentDocuments";
import { useNotifications } from "@/features/notifications/useNotifications";
import { useMarkRead } from "@/features/notifications/useMarkRead";
import { formatDistanceToNow } from "date-fns";
import { Skeleton } from "@/components/ui/skeleton";

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export default function Dashboard() {
  const tDashboard = useTranslations("dashboard");
  const tDocuments = useTranslations("documents");
  const tNotifications = useTranslations("notifications");
  const { data: stats, isLoading: statsLoading } = useDocumentStats();
  const { data: recentDocs, isLoading: recentLoading } = useRecentDocuments(10);
  const { data: notifData, isLoading: notifLoading } = useNotifications({ status: "unread", limit: 5 });
  const { mutate: markRead, isPending: isMarkingRead } = useMarkRead();

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">{tDashboard("title")}</h1>
        <Button asChild>
          <Link href="/documents/upload">
            <Upload className="me-2 h-4 w-4" />
            {tDocuments("upload")}
          </Link>
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{tDashboard("totalDocuments")}</CardTitle>
            <FileStack className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {statsLoading ? <Skeleton className="h-8 w-16" /> : stats?.total_documents ?? 0}
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{tDashboard("pendingProcessing")}</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {statsLoading ? <Skeleton className="h-8 w-16" /> : stats?.pending_documents ?? 0}
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{tDashboard("completedThisWeek")}</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {statsLoading ? <Skeleton className="h-8 w-16" /> : stats?.completed_this_week ?? 0}
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">{tDashboard("storageUsed")}</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {statsLoading ? <Skeleton className="h-8 w-24" /> : formatBytes(stats?.storage_used_bytes ?? 0)}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-[1fr_250px] lg:grid-cols-[1fr_300px]">
          <Card className="col-span-1">
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>{tDashboard("recentDocuments")}</CardTitle>
              <Button variant="outline" size="sm" asChild>
              <Link href="/documents">{tDashboard("viewAll")}</Link>
              </Button>
            </CardHeader>
          <CardContent>
            {recentLoading ? (
              <div className="space-y-2">
                {[1,2,3,4].map(i => <Skeleton key={i} className="h-12 w-full" />)}
              </div>
            ) : recentDocs?.documents?.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <FileText className="mx-auto h-8 w-8 mb-2 opacity-50" />
                <p>{tDashboard("noDocumentsUploadedYet")}</p>
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{tDashboard("titleColumn")}</TableHead>
                    <TableHead>{tDashboard("typeColumn")}</TableHead>
                    <TableHead>{tDashboard("statusColumn")}</TableHead>
                    <TableHead className="text-end">{tDashboard("uploadedColumn")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {recentDocs?.documents?.map((doc) => (
                    <TableRow key={doc.id}>
                      <TableCell className="font-medium">
                        <Link href={`/documents/${doc.id}`} className="hover:underline">
                          {doc.title}
                        </Link>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{doc.doc_type}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant={doc.status === 'processed' ? 'default' : doc.status === 'failed' ? 'destructive' : 'secondary'}>
                          {doc.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-end text-muted-foreground text-sm">
                        {formatDistanceToNow(new Date(doc.created_at), { addSuffix: true })}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>

        <div className="flex flex-col gap-4">
          <Card className="col-span-1 h-fit">
            <CardHeader>
              <CardTitle>{tDashboard("quickActions")}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              <Button className="w-full justify-start" asChild>
                <Link href="/documents/upload">
                  <Upload className="me-2 h-4 w-4" />
                  {tDashboard("uploadNewDocument")}
                </Link>
              </Button>
              <Button variant="outline" className="w-full justify-start" asChild>
                <Link href="/search">
                  <Search className="me-2 h-4 w-4" />
                  {tDashboard("searchDatabase")}
                </Link>
              </Button>
              <Button variant="outline" className="w-full justify-start" asChild>
                <Link href="/reminders">
                  <Bell className="me-2 h-4 w-4" />
                  {tDashboard("manageReminders")}
                </Link>
              </Button>
            </CardContent>
          </Card>

          <Card className="col-span-1 flex-1">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle>{tDashboard("unreadNotifications")}</CardTitle>
              {notifData?.unread_count ? (
                <Badge variant="secondary">{notifData.unread_count} {tDashboard("new")}</Badge>
              ) : null}
            </CardHeader>
            <CardContent>
              {notifLoading ? (
                <div className="space-y-4 mt-2">
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
              ) : notifData?.notifications?.length === 0 || !notifData?.notifications ? (
                <div className="text-center py-8 text-muted-foreground flex flex-col items-center">
                  <Bell className="mx-auto h-8 w-8 mb-2 opacity-20" />
                  <p className="text-sm">{tNotifications("allCaughtUp")}</p>
                </div>
              ) : (
                <div className="flex flex-col gap-4 mt-2">
                  {notifData.notifications.map((notif) => (
                    <div key={notif.id} className="flex items-start gap-3">
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
                        className="h-6 w-6 shrink-0 text-muted-foreground hover:text-primary -mt-1"
                        onClick={() => markRead(notif.id)}
                        disabled={isMarkingRead}
                        title={tNotifications("markAsRead")}
                      >
                        <Check className="h-4 w-4" />
                        <span className="sr-only">{tNotifications("markAsRead")}</span>
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
