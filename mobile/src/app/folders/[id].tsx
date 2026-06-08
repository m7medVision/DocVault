import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { Button, Card } from 'heroui-native';
import { useLocalSearchParams, useRouter } from 'expo-router';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { StatusBadge } from '@/components/status-badge';
import { FolderRow } from '@/components/folder-row';
import { FolderCreateInput } from '@/components/folder-create-input';
import { FolderActionsSheet } from '@/components/folder-actions-sheet';
import { PlusIcon, FolderIcon, ChevronRightIcon } from '@/components/icons';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import { useFolderDetail } from '@/features/folders/use-folder-detail';
import { useFolders } from '@/features/folders/use-folders';
import type { Folder } from '@/features/folders/types';

export default function FolderDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const theme = useTheme();
  const { t, isRTL } = useTranslation();
  const {
    folder,
    ancestors,
    children,
    documents,
    loading,
    documentsLoading,
    reload,
  } = useFolderDetail({ id });
  const { create, isMutating } = useFolders();
  const [creating, setCreating] = useState(false);
  const [actionsTarget, setActionsTarget] = useState<Folder | null>(null);

  async function handleCreate(name: string) {
    if (!folder) return;
    try {
      await create(name, folder.id);
      setCreating(false);
      await reload();
    } catch {
      setCreating(false);
    }
  }

  function handleRequestCreateSubfolder(parent: Folder) {
    setCreating(true);
    void parent;
  }

  if (loading) {
    return (
      <DocVaultScreen>
        <ThemedText type="small" themeColor="textSecondary">{t('common.loading')}</ThemedText>
      </DocVaultScreen>
    );
  }

  if (!folder) {
    return (
      <DocVaultScreen>
        <Pressable onPress={() => router.back()}>
          <ThemedText type="linkPrimary">{t('common.back')}</ThemedText>
        </Pressable>
        <Card className="rounded-2xl border border-divider bg-content1 p-4">
          <ThemedText type="smallBold" themeColor="error">
            {t('folders.couldNotLoad')}
          </ThemedText>
        </Card>
      </DocVaultScreen>
    );
  }

  return (
    <DocVaultScreen>
      <View style={styles.breadcrumbRow}>
        <Pressable onPress={() => router.push('/folders' as never)} hitSlop={8}>
          <ThemedText type="link" style={{ color: theme.accent }}>
            {t('folders.breadcrumbFolders')}
          </ThemedText>
        </Pressable>
        {ancestors.slice(0, -1).map((ancestor) => (
          <View key={ancestor.id} style={styles.crumb}>
            <ChevronRightIcon
              size={14}
              color={theme.muted}
              strokeWidth={1.5}
            />
            <Pressable
              onPress={() => router.push({ pathname: '/folders/[id]' as never, params: { id: ancestor.id } })}
              hitSlop={8}
            >
              <ThemedText type="link" style={{ color: theme.muted }}>
                {ancestor.name}
              </ThemedText>
            </Pressable>
          </View>
        ))}
        {ancestors.length > 1 ? (
          <View style={styles.crumb}>
            <ChevronRightIcon size={14} color={theme.muted} strokeWidth={1.5} />
            <ThemedText type="smallBold" numberOfLines={1} style={styles.currentCrumb}>
              {folder.name}
            </ThemedText>
          </View>
        ) : null}
      </View>

      <View style={[styles.header, isRTL && styles.rtlRow]}>
        <View style={[styles.iconWrap, { backgroundColor: `${theme.accent}1A` }]}>
          <FolderIcon size={22} color={theme.accent} strokeWidth={1.5} />
        </View>
        <View style={styles.headerText}>
          <ThemedText type="subtitle" numberOfLines={2}>{folder.name}</ThemedText>
          <ThemedText type="code" themeColor="textSecondary">
            {t('folders.documents')}: {documents.length}
            {children.length > 0 ? `  ·  ${t('folders.childFolders')}: ${children.length}` : ''}
          </ThemedText>
        </View>
        {!creating ? (
          <Button variant="secondary" size="sm" onPress={() => setCreating(true)}>
            <View style={styles.buttonInner}>
              <PlusIcon size={14} color={theme.foreground} strokeWidth={2} />
              <ThemedText type="smallBold">{t('folders.newSubfolder')}</ThemedText>
            </View>
          </Button>
        ) : null}
      </View>

      {creating ? (
        <FolderCreateInput
          onSubmit={handleCreate}
          onCancel={() => setCreating(false)}
          busy={isMutating}
        />
      ) : null}

      {children.length > 0 ? (
        <View style={styles.section}>
          <ThemedText type="smallBold" themeColor="textSecondary">
            {t('folders.childFolders')}
          </ThemedText>
          <View style={styles.list}>
            {children.map((child) => (
              <FolderRow
                key={child.id}
                folder={child}
                onPress={() => router.push({ pathname: '/folders/[id]' as never, params: { id: child.id } })}
                onLongPress={() => setActionsTarget(child)}
              />
            ))}
          </View>
        </View>
      ) : null}

      <View style={styles.section}>
        <ThemedText type="smallBold" themeColor="textSecondary">
          {t('folders.documents')}
        </ThemedText>
        {documentsLoading ? (
          <ThemedText type="small" themeColor="textSecondary">{t('common.loading')}</ThemedText>
        ) : documents.length === 0 ? (
          <Card className="rounded-2xl border border-divider bg-content1 p-6">
            <View style={styles.emptyBody}>
              <FolderIcon size={26} color={theme.muted} strokeWidth={1.25} />
              <ThemedText type="smallBold">{t('folders.noDocuments')}</ThemedText>
            </View>
          </Card>
        ) : (
          <View style={styles.list}>
            {documents.map((doc) => (
              <Pressable
                key={doc.id}
                onPress={() => router.push({ pathname: '/documents/[id]', params: { id: doc.id } })}
              >
                <Card className="rounded-2xl border border-divider bg-content1 p-3">
                  <View style={styles.docRow}>
                    <View style={styles.docInfo}>
                      <ThemedText type="smallBold" numberOfLines={1}>{doc.title}</ThemedText>
                      <ThemedText type="code" themeColor="textSecondary">
                        {doc.doc_type} · {new Date(doc.created_at).toLocaleDateString()}
                      </ThemedText>
                    </View>
                    <StatusBadge status={doc.status} />
                    <Pressable
                      onPress={() => router.push({ pathname: '/documents/move/[id]' as never, params: { id: doc.id } })}
                      hitSlop={12}
                      accessibilityRole="button"
                      accessibilityLabel={t('folders.moveTitle')}
                    >
                      <View style={[styles.moveBtn, { borderColor: theme.border }]}>
                        <ThemedText type="code" themeColor="textSecondary">{t('folders.moveTo')}</ThemedText>
                      </View>
                    </Pressable>
                  </View>
                </Card>
              </Pressable>
            ))}
          </View>
        )}
      </View>

      {actionsTarget ? (
        <FolderActionsSheet
          folder={actionsTarget}
          visible={Boolean(actionsTarget)}
          onClose={() => setActionsTarget(null)}
          onRequestCreateSubfolder={handleRequestCreateSubfolder}
        />
      ) : null}
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  breadcrumbRow: {
    flexDirection: 'row',
    alignItems: 'center',
    flexWrap: 'wrap',
    gap: Spacing.one,
  },
  crumb: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.one,
  },
  currentCrumb: {
    maxWidth: 160,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.three,
  },
  rtlRow: {
    flexDirection: 'row-reverse',
  },
  iconWrap: {
    width: 44,
    height: 44,
    borderRadius: 12,
    alignItems: 'center',
    justifyContent: 'center',
  },
  headerText: {
    flex: 1,
    gap: 1,
  },
  buttonInner: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.one,
  },
  section: {
    gap: Spacing.two,
  },
  list: {
    gap: Spacing.two,
  },
  docRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.three,
  },
  docInfo: {
    flex: 1,
    gap: 1,
  },
  emptyBody: {
    gap: Spacing.one,
    alignItems: 'center',
  },
  moveBtn: {
    borderWidth: 1,
    borderRadius: 999,
    paddingHorizontal: Spacing.two,
    paddingVertical: Spacing.half + 1,
  },
});
