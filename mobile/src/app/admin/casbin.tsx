import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';
import { useRouter } from 'expo-router';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useRole } from '@/hooks/use-role';
import { usePolicies } from '@/features/admin/hooks';
import { useTranslation } from '@/lib/i18n';

export default function AdminCasbinScreen() {
  const { t } = useTranslation();
  const theme = useTheme();
  const router = useRouter();
  const { isAdmin } = useRole();
  const { data, isLoading } = usePolicies();

  const policies = data?.policies ?? [];

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <Pressable onPress={() => router.back()}>
          <ThemedText type="linkPrimary">{t('common.back')}</ThemedText>
        </Pressable>
        <ThemedText type="subtitle">{t('admin.policies')}</ThemedText>
      </View>

      {!isAdmin ? (
        <ThemedText type="small" themeColor="error">
          {t('admin.notAuthorized')}
        </ThemedText>
      ) : isLoading ? (
        <ActivityIndicator style={styles.loader} />
      ) : policies.length === 0 ? (
        <ThemedText type="small" themeColor="muted">
          {t('admin.policiesEmpty')}
        </ThemedText>
      ) : (
        <View style={styles.list}>
          <View style={styles.headerRow}>
            <ThemedText type="code" themeColor="textSecondary" style={styles.col}>
              {t('admin.subject')}
            </ThemedText>
            <ThemedText type="code" themeColor="textSecondary" style={styles.col}>
              {t('admin.object')}
            </ThemedText>
            <ThemedText type="code" themeColor="textSecondary" style={styles.col}>
              {t('admin.action')}
            </ThemedText>
          </View>
          {policies.map((p, i) => (
            <View
              key={`${p.subject}-${p.object}-${p.action}-${i}`}
              style={[styles.row, { backgroundColor: theme.surface, borderColor: theme.border }]}
            >
              <ThemedText type="small" style={styles.col} numberOfLines={1}>
                {p.subject}
              </ThemedText>
              <ThemedText type="small" style={styles.col} numberOfLines={1}>
                {p.object}
              </ThemedText>
              <ThemedText type="small" style={styles.col} numberOfLines={1}>
                {p.action}
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
  list: { gap: Spacing.one },
  headerRow: {
    flexDirection: 'row',
    paddingHorizontal: Spacing.three,
  },
  row: {
    flexDirection: 'row',
    borderRadius: Spacing.two,
    borderWidth: StyleSheet.hairlineWidth,
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.two,
  },
  col: { flex: 1 },
});
