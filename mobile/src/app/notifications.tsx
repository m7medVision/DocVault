import { useMemo, useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';
import { useRouter } from 'expo-router';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import {
  useInfiniteNotifications,
  useMarkRead,
} from '@/features/notifications/use-infinite-notifications';
import type { Notification } from '@/features/notifications/types';
import { useTranslation } from '@/lib/i18n';

type Tab = 'all' | 'unread';

export default function NotificationsScreen() {
  const { t } = useTranslation();
  const theme = useTheme();
  const router = useRouter();
  const [tab, setTab] = useState<Tab>('all');

  const status = tab === 'unread' ? 'unread' : undefined;
  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useInfiniteNotifications(status);
  const markRead = useMarkRead();

  const notifications = useMemo(
    () => data?.pages.flatMap((p) => p.notifications) ?? [],
    [data],
  );
  const unreadCount = data?.pages[0]?.unread_count ?? 0;

  const handlePress = async (n: Notification) => {
    if (n.status === 'unread') {
      markRead.mutate(n.id);
    }
    if (n.link) {
      router.push(n.link as never);
    }
  };

  const handleMarkAll = () => {
    notifications.filter((n) => n.status === 'unread').forEach((n) => markRead.mutate(n.id));
  };

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <Pressable onPress={() => router.back()}>
          <ThemedText type="linkPrimary">{t('common.back')}</ThemedText>
        </Pressable>
        <ThemedText type="subtitle">{t('notifications.title')}</ThemedText>
        <ThemedText type="small" themeColor="muted">
          {t('notifications.subtitle')}
        </ThemedText>
      </View>

      <View style={styles.tabRow}>
        <View style={styles.tabs}>
          {(['all', 'unread'] as Tab[]).map((value) => (
            <Pressable
              key={value}
              onPress={() => setTab(value)}
              style={[
                styles.tab,
                { backgroundColor: tab === value ? theme.accent : theme.surface },
              ]}
            >
              <ThemedText type="smallBold" style={tab === value ? { color: '#fff' } : undefined}>
                {value === 'all' ? t('notifications.all') : t('notifications.unreadTab')}
                {value === 'unread' && unreadCount > 0 ? ` (${unreadCount})` : ''}
              </ThemedText>
            </Pressable>
          ))}
        </View>
        {unreadCount > 0 && (
          <Pressable onPress={handleMarkAll}>
            <ThemedText type="link" themeColor="accent">
              {t('notifications.markAllRead')}
            </ThemedText>
          </Pressable>
        )}
      </View>

      {isLoading ? (
        <ActivityIndicator style={styles.loader} />
      ) : notifications.length === 0 ? (
        <View style={[styles.card, { backgroundColor: theme.surface }]}>
          <ThemedText type="small" themeColor="muted">
            {t('notifications.empty')}
          </ThemedText>
        </View>
      ) : (
        <View style={styles.list}>
          {notifications.map((n) => (
            <Pressable
              key={n.id}
              onPress={() => handlePress(n)}
              style={[styles.row, { backgroundColor: theme.surface, borderColor: theme.border }]}
            >
              <View style={styles.rowText}>
                <View style={styles.titleRow}>
                  <ThemedText type="smallBold" numberOfLines={1} style={styles.flex}>
                    {n.title}
                  </ThemedText>
                  {n.status === 'unread' && (
                    <View style={[styles.dot, { backgroundColor: theme.accent }]} />
                  )}
                </View>
                {n.body ? (
                  <ThemedText type="small" themeColor="textSecondary" numberOfLines={2}>
                    {n.body}
                  </ThemedText>
                ) : null}
                <ThemedText type="code" themeColor="textSecondary">
                  {new Date(n.created_at).toLocaleDateString()}
                </ThemedText>
              </View>
            </Pressable>
          ))}

          {hasNextPage && (
            <Pressable
              onPress={() => fetchNextPage()}
              disabled={isFetchingNextPage}
              style={[styles.loadMore, { borderColor: theme.border }]}
            >
              {isFetchingNextPage ? (
                <ActivityIndicator size="small" />
              ) : (
                <ThemedText type="smallBold" themeColor="accent">
                  {t('notifications.loadMore')}
                </ThemedText>
              )}
            </Pressable>
          )}
        </View>
      )}
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  header: { gap: Spacing.one },
  tabRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  tabs: { flexDirection: 'row', gap: Spacing.two },
  tab: {
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.one + 2,
    borderRadius: 999,
  },
  loader: { marginTop: Spacing.four },
  card: {
    borderRadius: Spacing.three,
    padding: Spacing.four,
    alignItems: 'center',
  },
  list: { gap: Spacing.two },
  row: {
    flexDirection: 'row',
    borderRadius: Spacing.three,
    borderWidth: StyleSheet.hairlineWidth,
    padding: Spacing.three,
  },
  rowText: { flex: 1, gap: 2 },
  titleRow: { flexDirection: 'row', alignItems: 'center', gap: Spacing.two },
  flex: { flex: 1 },
  dot: { width: 8, height: 8, borderRadius: 999 },
  loadMore: {
    alignItems: 'center',
    paddingVertical: Spacing.two,
    borderRadius: 999,
    borderWidth: StyleSheet.hairlineWidth,
  },
});
