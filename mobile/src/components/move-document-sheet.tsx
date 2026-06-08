import { useMemo, useState } from 'react';
import { Modal, Pressable, ScrollView, StyleSheet, View } from 'react-native';
import { Button, Card } from 'heroui-native';

import { ThemedText } from './themed-text';
import { CheckIcon, FolderIcon } from './icons';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import { useFolders } from '@/features/folders/use-folders';
import type { Folder } from '@/features/folders/types';

interface MoveDocumentSheetProps {
  visible: boolean;
  documentTitle: string;
  currentFolderId?: string;
  onClose: () => void;
  onConfirm: (folderId?: string) => Promise<void> | void;
  busy?: boolean;
}

interface TreeRow {
  folder: Folder;
  depth: number;
  disabled: boolean;
}

export function MoveDocumentSheet({
  visible,
  documentTitle,
  currentFolderId,
  onClose,
  onConfirm,
  busy,
}: MoveDocumentSheetProps) {
  const { t } = useTranslation();
  const theme = useTheme();
  const { folders } = useFolders();
  const [selected, setSelected] = useState<string | undefined>(currentFolderId);

  const rows = useMemo<TreeRow[]>(() => {
    if (folders.length === 0) return [];
    const byParent = new Map<string | undefined, Folder[]>();
    for (const f of folders) {
      const list = byParent.get(f.parent_id) ?? [];
      list.push(f);
      byParent.set(f.parent_id, list);
    }
    for (const list of byParent.values()) {
      list.sort((a, b) => a.name.localeCompare(b.name));
    }

    const result: TreeRow[] = [];

    function walk(parentId: string | undefined, depth: number) {
      const children = byParent.get(parentId) ?? [];
      for (const child of children) {
        result.push({ folder: child, depth, disabled: false });
        walk(child.id, depth + 1);
      }
    }

    walk(undefined, 0);
    return result;
  }, [folders]);

  const isRootSelected = !selected;

  return (
    <Modal visible={visible} transparent animationType="slide" onRequestClose={onClose}>
      <Pressable style={styles.backdrop} onPress={onClose}>
        <View style={styles.root}>
          <Card className="rounded-3xl border border-divider bg-content1 p-0" style={styles.sheet}>
          <View style={styles.header}>
            <View style={styles.headerText}>
              <ThemedText type="smallBold">{t('folders.moveTitle')}</ThemedText>
              <ThemedText type="small" themeColor="textSecondary" numberOfLines={1}>
                {documentTitle}
              </ThemedText>
            </View>
            <Pressable onPress={onClose} hitSlop={12}>
              <ThemedText type="link" style={{ color: theme.muted }}>
                {t('common.cancel')}
              </ThemedText>
            </Pressable>
          </View>

          <ScrollView
            style={styles.scroll}
            contentContainerStyle={styles.scrollContent}
            keyboardShouldPersistTaps="handled"
          >
            <Pressable
              onPress={() => setSelected(undefined)}
              style={({ pressed }) => pressed && styles.pressed}
            >
              <Card
                className={`rounded-2xl border bg-content1 p-3 ${
                  isRootSelected ? 'border-accent' : 'border-divider'
                }`}
              >
                <View style={styles.row}>
                  <View style={[styles.iconWrap, { backgroundColor: `${theme.muted}1A` }]}>
                    <FolderIcon size={18} color={theme.foreground} strokeWidth={1.5} />
                  </View>
                  <View style={styles.textCol}>
                    <ThemedText type="smallBold">{t('folders.root')}</ThemedText>
                  </View>
                  {isRootSelected ? <CheckIcon size={18} color={theme.accent} strokeWidth={2} /> : null}
                </View>
              </Card>
            </Pressable>

            {rows.length === 0 ? (
              <Card className="rounded-2xl border border-divider bg-content1 p-4">
                <ThemedText type="small" themeColor="textSecondary">
                  {t('folders.empty')}
                </ThemedText>
              </Card>
            ) : (
              rows.map(({ folder, depth }) => {
                const isSelected = selected === folder.id;
                const isCurrent = currentFolderId === folder.id;
                return (
                  <Pressable
                    key={folder.id}
                    onPress={() => setSelected(folder.id)}
                    style={({ pressed }) => pressed && styles.pressed}
                  >
                    <Card
                      className={`rounded-2xl border bg-content1 p-3 ${
                        isSelected ? 'border-accent' : 'border-divider'
                      }`}
                    >
                      <View style={[styles.row, { paddingLeft: depth * Spacing.three }]}>
                        <View style={[styles.iconWrap, { backgroundColor: `${theme.accent}1A` }]}>
                          <FolderIcon size={18} color={theme.accent} strokeWidth={1.5} />
                        </View>
                        <View style={styles.textCol}>
                          <ThemedText type="smallBold" numberOfLines={1}>
                            {folder.name}
                          </ThemedText>
                          {isCurrent ? (
                            <ThemedText type="code" themeColor="textSecondary">
                              {t('folders.alreadyHere')}
                            </ThemedText>
                          ) : null}
                        </View>
                        {isSelected ? <CheckIcon size={18} color={theme.accent} strokeWidth={2} /> : null}
                      </View>
                    </Card>
                  </Pressable>
                );
              })
            )}
          </ScrollView>

          <View style={styles.footer}>
            <Button
              variant="primary"
              onPress={() => void onConfirm(selected)}
              isDisabled={busy || selected === currentFolderId}
            >
              <Button.Label>{t('folders.moveHere')}</Button.Label>
            </Button>
          </View>
          </Card>
        </View>
      </Pressable>
    </Modal>
  );
}

const styles = StyleSheet.create({
  backdrop: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.45)',
    justifyContent: 'flex-end',
  },
  root: {
    backgroundColor: 'transparent',
    maxHeight: '85%',
    width: '100%',
  },
  sheet: {
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    borderBottomLeftRadius: 0,
    borderBottomRightRadius: 0,
    overflow: 'hidden',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: Spacing.three,
    backgroundColor: 'transparent',
  },
  headerText: {
    flex: 1,
    gap: 2,
  },
  scroll: {
    backgroundColor: 'transparent',
  },
  scrollContent: {
    padding: Spacing.three,
    gap: Spacing.two,
    paddingBottom: Spacing.five,
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
    gap: 1,
  },
  pressed: {
    opacity: 0.75,
  },
  footer: {
    padding: Spacing.three,
    backgroundColor: 'transparent',
  },
});
