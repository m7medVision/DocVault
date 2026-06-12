import { Pressable, StyleSheet, View } from 'react-native';
import { useRouter } from 'expo-router';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useRole } from '@/hooks/use-role';
import { useTranslation } from '@/lib/i18n';

export default function AdminHomeScreen() {
  const { t } = useTranslation();
  const theme = useTheme();
  const router = useRouter();
  const { isAdmin } = useRole();

  if (!isAdmin) {
    return (
      <DocVaultScreen>
        <Pressable onPress={() => router.back()}>
          <ThemedText type="linkPrimary">{t('common.back')}</ThemedText>
        </Pressable>
        <ThemedText type="small" themeColor="error">
          {t('admin.notAuthorized')}
        </ThemedText>
      </DocVaultScreen>
    );
  }

  const links: { label: string; route: '/admin/members' | '/admin/audit' | '/admin/casbin' }[] = [
    { label: t('admin.members'), route: '/admin/members' },
    { label: t('admin.audit'), route: '/admin/audit' },
    { label: t('admin.policies'), route: '/admin/casbin' },
  ];

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <Pressable onPress={() => router.back()}>
          <ThemedText type="linkPrimary">{t('common.back')}</ThemedText>
        </Pressable>
        <ThemedText type="subtitle">{t('admin.title')}</ThemedText>
        <ThemedText type="small" themeColor="muted">
          {t('admin.subtitle')}
        </ThemedText>
      </View>

      <View style={styles.list}>
        {links.map((link) => (
          <Pressable
            key={link.route}
            onPress={() => router.push(link.route)}
            style={[styles.row, { backgroundColor: theme.surface, borderColor: theme.border }]}
          >
            <ThemedText type="smallBold">{link.label}</ThemedText>
            <ThemedText type="code" themeColor="textSecondary">
              ›
            </ThemedText>
          </Pressable>
        ))}
      </View>
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  header: { gap: Spacing.one },
  list: { gap: Spacing.two },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    borderRadius: Spacing.three,
    borderWidth: StyleSheet.hairlineWidth,
    paddingHorizontal: Spacing.three,
    paddingVertical: Spacing.three,
  },
});
