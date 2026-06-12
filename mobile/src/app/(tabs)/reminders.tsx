import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';
import { useRouter } from 'expo-router';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { ReminderRow } from '@/components/reminder-row';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useReminders, type ReminderFilter } from '@/features/reminders/use-reminders';
import { useTranslation } from '@/lib/i18n';

const FILTERS: ReminderFilter[] = ['active', 'all'];

export default function RemindersScreen() {
  const { t } = useTranslation();
  const theme = useTheme();
  const router = useRouter();
  const { reminders, filter, setFilter, loading, error, dismiss } = useReminders();

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <ThemedText type="subtitle">{t('reminders.title')}</ThemedText>
        <ThemedText type="small" themeColor="muted">
          {t('reminders.subtitle')}
        </ThemedText>
      </View>

      <View style={styles.tabs}>
        {FILTERS.map((f) => (
          <Pressable
            key={f}
            onPress={() => setFilter(f)}
            style={[
              styles.tab,
              { backgroundColor: filter === f ? theme.accent : theme.surface },
            ]}
          >
            <ThemedText type="smallBold" style={filter === f ? { color: '#fff' } : undefined}>
              {t(`reminders.${f}`)}
            </ThemedText>
          </Pressable>
        ))}
      </View>

      {loading ? (
        <ActivityIndicator style={styles.loader} />
      ) : error ? (
        <View style={[styles.card, { backgroundColor: theme.errorMuted }]}>
          <ThemedText type="small" themeColor="error">
            {error}
          </ThemedText>
        </View>
      ) : reminders.length === 0 ? (
        <View style={[styles.card, { backgroundColor: theme.surface }]}>
          <ThemedText type="small" themeColor="muted">
            {t('reminders.empty')}
          </ThemedText>
        </View>
      ) : (
        <View style={styles.list}>
          {reminders.map((reminder) => (
            <View key={reminder.id} style={styles.row}>
              <View style={styles.rowMain}>
                <ReminderRow
                  reminder={reminder}
                  onPress={() => router.push(`/documents/${reminder.document_id}`)}
                />
              </View>
              <Pressable
                onPress={() => dismiss(reminder.id)}
                style={[styles.dismissBtn, { borderColor: theme.border }]}
              >
                <ThemedText type="code" themeColor="textSecondary">
                  {t('reminders.dismiss')}
                </ThemedText>
              </Pressable>
            </View>
          ))}
        </View>
      )}
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  header: {
    gap: Spacing.one,
  },
  tabs: {
    flexDirection: 'row',
    gap: Spacing.two,
  },
  tab: {
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.one + 2,
    borderRadius: 999,
  },
  loader: {
    marginTop: Spacing.four,
  },
  card: {
    borderRadius: Spacing.three,
    padding: Spacing.four,
    alignItems: 'center',
  },
  list: {
    gap: Spacing.two,
  },
  row: {
    gap: Spacing.one,
  },
  rowMain: {
    flex: 1,
  },
  dismissBtn: {
    alignSelf: 'flex-end',
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.one,
    borderRadius: 999,
    borderWidth: StyleSheet.hairlineWidth,
  },
});
