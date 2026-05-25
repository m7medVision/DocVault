import { router } from 'expo-router';
import { Pressable, ScrollView, StyleSheet, View } from 'react-native';
import { Button, Card } from 'heroui-native';

import { DocVaultScreen } from '@/components/docvault-screen';
import { StatCard } from '@/components/stat-card';
import { StatusBadge } from '@/components/status-badge';
import { ReminderRow } from '@/components/reminder-row';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useAuth } from '@/lib/auth/auth-context';
import { useTranslation } from '@/lib/i18n';
import { useDocumentStats } from '@/features/documents/use-document-stats';
import { useDocuments } from '@/features/documents/use-documents';
import { useReminders } from '@/features/reminders/use-reminders';
import { useNotifications } from '@/features/notifications/use-notifications';

function formatStorage(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function relativeTime(dateStr: string): string {
  const now = Date.now();
  const date = new Date(dateStr).getTime();
  const diffMs = now - date;
  const diffMins = Math.floor(diffMs / 60_000);
  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return `${diffHours}h ago`;
  const diffDays = Math.floor(diffHours / 24);
  if (diffDays < 30) return `${diffDays}d ago`;
  return new Date(dateStr).toLocaleDateString();
}

export default function HomeScreen() {
  const { user } = useAuth();
  const { t, locale } = useTranslation();
  const isRTL = locale === 'ar';

  const { data: stats, isLoading: statsLoading } = useDocumentStats();
  const { documents: recentDocs, loading: docsLoading } = useDocuments({ limit: 5 });
  const { reminders, loading: remindersLoading } = useReminders();
  const { data: notifData } = useNotifications({ status: 'unread', limit: 3 });

  const activeReminders = reminders
    .filter((r) => r.active)
    .sort((a, b) => (a.days_until ?? 999) - (b.days_until ?? 999))
    .slice(0, 3);

  const greeting = user?.displayName
    ? t('home.greeting', { name: user.displayName })
    : t('home.subtitle');

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <View style={styles.headerText}>
          <ThemedText type="subtitle">{greeting}</ThemedText>
          <ThemedText type="small" themeColor="textSecondary">
            {t('home.subtitle')}
          </ThemedText>
        </View>
        {notifData?.unreadCount ? (
          <Button variant="secondary" size="sm" onPress={() => router.push('/reminders')}>
            <Button.Label>
              {t('notifications.unread', { count: notifData.unreadCount })}
            </Button.Label>
          </Button>
        ) : null}
      </View>

      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.statsScroll}
      >
        <View style={[styles.statsRow, isRTL && styles.rtlRow]}>
          <View style={styles.statItem}>
            <StatCard
              value={statsLoading ? '—' : (stats?.total_documents ?? 0)}
              label={t('home.totalDocuments')}
              loading={statsLoading}
            />
          </View>
          <View style={styles.statItem}>
            <StatCard
              value={statsLoading ? '—' : (stats?.pending_documents ?? 0)}
              label={t('home.pendingDocuments')}
              loading={statsLoading}
            />
          </View>
          <View style={styles.statItem}>
            <StatCard
              value={statsLoading ? '—' : (stats?.completed_this_week ?? 0)}
              label={t('home.completedThisWeek')}
              loading={statsLoading}
            />
          </View>
          <View style={styles.statItem}>
            <StatCard
              value={statsLoading ? '—' : formatStorage(stats?.storage_used_bytes ?? 0)}
              label={t('home.storageUsed')}
              loading={statsLoading}
            />
          </View>
        </View>
      </ScrollView>

      <View style={[styles.quickActions, isRTL && styles.rtlRow]}>
        <Button variant="primary" onPress={() => router.push('/scan')} style={styles.actionBtn}>
          <Button.Label>{t('home.scan')}</Button.Label>
        </Button>
        <Button variant="secondary" onPress={() => router.push('/documents')} style={styles.actionBtn}>
          <Button.Label>{t('home.upload')}</Button.Label>
        </Button>
      </View>

      <SectionHeader
        title={t('home.recentDocuments')}
        link={t('home.viewAll')}
        onLinkPress={() => router.push('/documents')}
        isRTL={isRTL}
      />
      {docsLoading ? (
        <ThemedText type="small" themeColor="textSecondary">{t('common.loading')}</ThemedText>
      ) : recentDocs.length === 0 ? (
        <EmptyState message={t('home.noDocuments')} />
      ) : (
        <View style={styles.list}>
          {recentDocs.map((doc) => (
            <Pressable
              key={doc.id}
              onPress={() => router.push({ pathname: '/documents/[id]', params: { id: doc.id } })}
              style={({ pressed }) => pressed && styles.pressed}
            >
              <Card className="rounded-2xl border border-divider bg-content1 p-3">
                <View style={styles.docRow}>
                  <View style={styles.docInfo}>
                    <ThemedText type="smallBold" numberOfLines={1}>
                      {doc.title}
                    </ThemedText>
                    <ThemedText type="code" themeColor="textSecondary">
                      {doc.doc_type} · {relativeTime(doc.created_at)}
                    </ThemedText>
                  </View>
                  <StatusBadge status={doc.status} />
                </View>
              </Card>
            </Pressable>
          ))}
        </View>
      )}

      <SectionHeader
        title={t('home.upcomingReminders')}
        link={t('home.viewAll')}
        onLinkPress={() => router.push('/reminders')}
        isRTL={isRTL}
      />
      {remindersLoading ? (
        <ThemedText type="small" themeColor="textSecondary">{t('common.loading')}</ThemedText>
      ) : activeReminders.length === 0 ? (
        <EmptyState message={t('home.noReminders')} />
      ) : (
        <View style={styles.list}>
          {activeReminders.map((r) => (
            <ReminderRow
              key={r.id}
              reminder={r}
              onPress={() => r.document_id && router.push({ pathname: '/documents/[id]', params: { id: r.document_id } })}
            />
          ))}
        </View>
      )}

      {notifData?.notifications && notifData.notifications.length > 0 && (
        <>
          <SectionHeader
            title={t('notifications.title')}
            isRTL={isRTL}
          />
          <View style={styles.list}>
            {notifData.notifications.map((n) => (
              <Card key={n.id} className="rounded-2xl border border-divider bg-content1 p-3">
                <View style={styles.notifRow}>
                  <View style={styles.docInfo}>
                    <ThemedText type="smallBold" numberOfLines={1}>
                      {n.title}
                    </ThemedText>
                    {n.body ? (
                      <ThemedText type="small" themeColor="textSecondary" numberOfLines={2}>
                        {n.body}
                      </ThemedText>
                    ) : null}
                  </View>
                </View>
              </Card>
            ))}
          </View>
        </>
      )}
    </DocVaultScreen>
  );
}

function SectionHeader({
  title,
  link,
  onLinkPress,
  isRTL,
}: {
  title: string;
  link?: string;
  onLinkPress?: () => void;
  isRTL?: boolean;
}) {
  return (
    <View style={[styles.sectionHeader, isRTL && styles.rtlRow]}>
      <ThemedText type="smallBold">{title}</ThemedText>
      {link && onLinkPress ? (
        <ThemedText type="linkPrimary" onPress={onLinkPress}>
          {link}
        </ThemedText>
      ) : null}
    </View>
  );
}

function EmptyState({ message }: { message: string }) {
  return (
    <Card className="rounded-2xl border border-divider bg-content1 p-4">
      <View style={styles.emptyCard}>
        <ThemedText type="small" themeColor="textSecondary">
          {message}
        </ThemedText>
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  header: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    gap: Spacing.three,
  },
  headerText: {
    flex: 1,
    gap: Spacing.one,
  },
  statsScroll: {
    paddingBottom: Spacing.one,
  },
  statsRow: {
    flexDirection: 'row',
    gap: Spacing.two,
  },
  rtlRow: {
    flexDirection: 'row-reverse',
  },
  statItem: {
    width: 140,
  },
  quickActions: {
    flexDirection: 'row',
    gap: Spacing.two,
  },
  actionBtn: {
    flex: 1,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  list: {
    gap: Spacing.two,
  },
  docRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.two,
  },
  docInfo: {
    flex: 1,
    gap: 1,
  },
  notifRow: {
    flexDirection: 'row',
    alignItems: 'flex-start',
  },
  emptyCard: {
    alignItems: 'center',
    gap: Spacing.one,
  },
  pressed: {
    opacity: 0.75,
  },
});
