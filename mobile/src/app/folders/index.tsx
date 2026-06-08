import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { Button, Card } from 'heroui-native';
import { useRouter } from 'expo-router';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { FolderRow } from '@/components/folder-row';
import { FolderCreateInput } from '@/components/folder-create-input';
import { FolderActionsSheet } from '@/components/folder-actions-sheet';
import { PlusIcon, FolderIcon } from '@/components/icons';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import { useFolders } from '@/features/folders/use-folders';
import type { Folder } from '@/features/folders/types';

function Skeleton({ theme }: { theme: ReturnType<typeof useTheme> }) {
  return (
    <Card className="rounded-2xl border border-divider bg-content1 p-3">
      <View style={styles.skeletonRow}>
        <View style={[styles.skeletonIcon, { backgroundColor: theme.muted, opacity: 0.15 }]} />
        <View style={styles.skeletonCol}>
          <View style={[styles.skeletonBar, { backgroundColor: theme.muted, opacity: 0.15, width: '60%' }]} />
          <View style={[styles.skeletonBar, { backgroundColor: theme.muted, opacity: 0.15, width: '30%' }]} />
        </View>
      </View>
    </Card>
  );
}

export default function FoldersScreen() {
  const router = useRouter();
  const theme = useTheme();
  const { t, isRTL } = useTranslation();
  const { folders, loading, error, reload, create, isMutating } = useFolders();
  const [creating, setCreating] = useState(false);
  const [actionsTarget, setActionsTarget] = useState<Folder | null>(null);

  const rootFolders = folders.filter((f) => !f.parent_id);

  async function handleCreate(name: string) {
    try {
      await create(name);
      setCreating(false);
    } catch {
      setCreating(false);
    }
  }

  function handleRequestCreateSubfolder(_parent: Folder) {
    setCreating(true);
  }

  return (
    <DocVaultScreen>
      <View style={[styles.header, isRTL && styles.rtlRow]}>
        <View style={styles.headerText}>
          <ThemedText type="subtitle">{t('folders.title')}</ThemedText>
          <ThemedText type="small" themeColor="textSecondary">
            {t('folders.subtitle')}
          </ThemedText>
        </View>
        {!creating ? (
          <Button
            variant="primary"
            size="sm"
            onPress={() => setCreating(true)}
          >
            <View style={styles.buttonInner}>
              <PlusIcon size={14} color="#fff" strokeWidth={2} />
              <ThemedText type="smallBold" style={styles.buttonLabel}>
                {t('folders.newFolder')}
              </ThemedText>
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

      {loading ? (
        <>
          <Skeleton theme={theme} />
          <Skeleton theme={theme} />
          <Skeleton theme={theme} />
        </>
      ) : error ? (
        <Pressable onPress={() => void reload()}>
          <Card className="rounded-2xl border border-divider bg-content1 p-4">
            <View style={styles.errorBody}>
              <ThemedText type="smallBold" themeColor="error">
                {t('folders.couldNotLoad')}
              </ThemedText>
              <ThemedText type="small" themeColor="textSecondary">
                {error}
              </ThemedText>
              <ThemedText type="link" style={{ color: theme.accent }}>
                {t('folders.retry')}
              </ThemedText>
            </View>
          </Card>
        </Pressable>
      ) : (
        <>
          <Pressable
            onPress={() => router.push('/documents')}
          >
            <Card className="rounded-2xl border border-divider bg-content1 p-3">
              <View style={styles.row}>
                <View style={[styles.iconWrap, { backgroundColor: `${theme.muted}1A` }]}>
                  <FolderIcon size={18} color={theme.foreground} strokeWidth={1.5} />
                </View>
                <View style={styles.textCol}>
                  <ThemedText type="smallBold">{t('folders.allDocuments')}</ThemedText>
                </View>
              </View>
            </Card>
          </Pressable>

          {rootFolders.length === 0 ? (
            <Card className="rounded-2xl border border-divider bg-content1 p-6">
              <View style={styles.emptyBody}>
                <FolderIcon size={28} color={theme.muted} strokeWidth={1.25} />
                <ThemedText type="smallBold">{t('folders.empty')}</ThemedText>
                <ThemedText type="small" themeColor="textSecondary" style={styles.emptyHint}>
                  {t('folders.emptyHint')}
                </ThemedText>
              </View>
            </Card>
          ) : (
            <View style={styles.list}>
              {rootFolders.map((folder) => (
                <FolderRow
                  key={folder.id}
                  folder={folder}
                  onPress={() => router.push({ pathname: '/folders/[id]' as never, params: { id: folder.id } })}
                  onLongPress={() => setActionsTarget(folder)}
                />
              ))}
            </View>
          )}
        </>
      )}

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
  header: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    gap: Spacing.three,
  },
  rtlRow: {
    flexDirection: 'row-reverse',
  },
  headerText: {
    flex: 1,
    gap: Spacing.one,
  },
  buttonInner: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.one,
  },
  buttonLabel: {
    color: '#fff',
  },
  list: {
    gap: Spacing.two,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.three,
  },
  iconWrap: {
    width: 36,
    height: 36,
    borderRadius: 10,
    alignItems: 'center',
    justifyContent: 'center',
  },
  textCol: {
    flex: 1,
  },
  emptyBody: {
    gap: Spacing.one,
    alignItems: 'center',
  },
  emptyHint: {
    textAlign: 'center',
  },
  errorBody: {
    gap: Spacing.one,
  },
  skeletonRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.three,
  },
  skeletonIcon: {
    width: 36,
    height: 36,
    borderRadius: 10,
  },
  skeletonCol: {
    flex: 1,
    gap: Spacing.one,
  },
  skeletonBar: {
    height: 10,
    borderRadius: 4,
  },
});
