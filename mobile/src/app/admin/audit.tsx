import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';
import { useRouter } from 'expo-router';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useRole } from '@/hooks/use-role';
import { useAuditLog } from '@/features/admin/hooks';
import { useTranslation } from '@/lib/i18n';

export default function AdminAuditScreen() {
  const { t } = useTranslation();
  const theme = useTheme();
  const router = useRouter();
  const { isAdmin } = useRole();
  const { data, isLoading } = useAuditLog();

  const events = data?.events ?? [];

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <Pressable onPress={() => router.back()}>
          <ThemedText type="linkPrimary">{t('common.back')}</ThemedText>
        </Pressable>
        <ThemedText type="subtitle">{t('admin.audit')}</ThemedText>
      </View>

      {!isAdmin ? (
        <ThemedText type="small" themeColor="error">
          {t('admin.notAuthorized')}
        </ThemedText>
      ) : isLoading ? (
        <ActivityIndicator style={styles.loader} />
      ) : events.length === 0 ? (
        <ThemedText type="small" themeColor="muted">
          {t('admin.auditEmpty')}
        </ThemedText>
      ) : (
        <View style={styles.list}>
          {events.map((e) => (
            <View
              key={e.id}
              style={[styles.row, { backgroundColor: theme.surface, borderColor: theme.border }]}
            >
              <View style={styles.rowTop}>
                <ThemedText type="smallBold">
                  {e.action} · {e.entity_type}
                </ThemedText>
                <ThemedText type="code" themeColor="textSecondary">
                  {new Date(e.created_at).toLocaleDateString()}
                </ThemedText>
              </View>
              <ThemedText type="code" themeColor="textSecondary" numberOfLines={1}>
                {e.entity_id}
              </ThemedText>
            </View>
          ))}
        </View>
      )}
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  header: { gap: Spacing.one },
  loader: { marginTop: Spacing.four },
  list: { gap: Spacing.two },
  row: {
    borderRadius: Spacing.three,
    borderWidth: StyleSheet.hairlineWidth,
    padding: Spacing.three,
    gap: 2,
  },
  rowTop: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
});
