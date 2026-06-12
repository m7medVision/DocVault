import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';
import { useRouter } from 'expo-router';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useRole } from '@/hooks/use-role';
import { useMembers, useUpdateMemberRole } from '@/features/admin/hooks';
import { useTranslation } from '@/lib/i18n';

const ROLES: ('admin' | 'member' | 'viewer')[] = ['admin', 'member', 'viewer'];

export default function AdminMembersScreen() {
  const { t } = useTranslation();
  const theme = useTheme();
  const router = useRouter();
  const { isAdmin } = useRole();
  const { data, isLoading } = useMembers();
  const updateRole = useUpdateMemberRole();

  const members = data?.members ?? [];

  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <Pressable onPress={() => router.back()}>
          <ThemedText type="linkPrimary">{t('common.back')}</ThemedText>
        </Pressable>
        <ThemedText type="subtitle">{t('admin.members')}</ThemedText>
      </View>

      {!isAdmin ? (
        <ThemedText type="small" themeColor="error">
          {t('admin.notAuthorized')}
        </ThemedText>
      ) : isLoading ? (
        <ActivityIndicator style={styles.loader} />
      ) : members.length === 0 ? (
        <ThemedText type="small" themeColor="muted">
          {t('admin.membersEmpty')}
        </ThemedText>
      ) : (
        <View style={styles.list}>
          {members.map((m) => (
            <View
              key={m.membership_id}
              style={[styles.row, { backgroundColor: theme.surface, borderColor: theme.border }]}
            >
              <ThemedText type="smallBold" numberOfLines={1}>
                {m.display_name || m.email}
              </ThemedText>
              <ThemedText type="small" themeColor="textSecondary" numberOfLines={1}>
                {m.email}
              </ThemedText>
              <View style={styles.roles}>
                {ROLES.map((role) => {
                  const isCurrent = m.role === role;
                  const isOwner = m.role === 'owner';
                  return (
                    <Pressable
                      key={role}
                      disabled={isCurrent || isOwner || updateRole.isPending}
                      onPress={() => updateRole.mutate({ id: m.membership_id, role })}
                      style={[
                        styles.roleChip,
                        {
                          backgroundColor: isCurrent ? theme.accent : 'transparent',
                          borderColor: theme.border,
                          opacity: isOwner ? 0.4 : 1,
                        },
                      ]}
                    >
                      <ThemedText
                        type="code"
                        style={isCurrent ? { color: '#fff' } : undefined}
                      >
                        {role}
                      </ThemedText>
                    </Pressable>
                  );
                })}
                {m.role === 'owner' && (
                  <View style={[styles.roleChip, { backgroundColor: theme.accent }]}>
                    <ThemedText type="code" style={{ color: '#fff' }}>
                      owner
                    </ThemedText>
                  </View>
                )}
              </View>
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
    gap: Spacing.one,
  },
  roles: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: Spacing.one,
    marginTop: Spacing.one,
  },
  roleChip: {
    paddingHorizontal: Spacing.two,
    paddingVertical: Spacing.half,
    borderRadius: 999,
    borderWidth: StyleSheet.hairlineWidth,
  },
});
