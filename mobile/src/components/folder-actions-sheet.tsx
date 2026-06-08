import { useState } from 'react';
import { Alert, Modal, Pressable, StyleSheet, View } from 'react-native';
import { Card, Input } from 'heroui-native';

import { ThemedText } from './themed-text';
import { PencilIcon, PlusIcon, TrashIcon } from './icons';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import { useFolders } from '@/features/folders/use-folders';
import type { Folder } from '@/features/folders/types';

interface FolderActionsSheetProps {
  folder: Folder;
  visible: boolean;
  onClose: () => void;
  onRequestCreateSubfolder?: (parent: Folder) => void;
}

export function FolderActionsSheet({
  folder,
  visible,
  onClose,
  onRequestCreateSubfolder,
}: FolderActionsSheetProps) {
  const { t } = useTranslation();
  const theme = useTheme();
  const { rename, remove, isMutating } = useFolders();
  const [renaming, setRenaming] = useState(false);
  const [draftName, setDraftName] = useState(folder.name);

  function openRename() {
    setDraftName(folder.name);
    setRenaming(true);
  }

  async function handleRename() {
    const next = draftName.trim();
    if (!next || next === folder.name) {
      setRenaming(false);
      return;
    }
    try {
      await rename(folder.id, next);
      setRenaming(false);
      onClose();
    } catch {
      setRenaming(false);
    }
  }

  function handleSubfolder() {
    onRequestCreateSubfolder?.(folder);
    onClose();
  }

  function confirmDelete() {
    Alert.alert(t('folders.confirmDelete'), t('folders.deleteBody'), [
      { text: t('common.cancel'), style: 'cancel' },
      {
        text: t('folders.delete'),
        style: 'destructive',
        onPress: async () => {
          try {
            await remove(folder.id);
            onClose();
          } catch {
            // hook exposes error in mutation lifecycle; nothing to do here
          }
        },
      },
    ]);
  }

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onClose}
    >
      <Pressable style={styles.backdrop} onPress={onClose}>
        <Pressable onPress={() => undefined}>
          <Card className="rounded-3xl border border-divider bg-content1 p-4" style={styles.sheet}>
            {renaming ? (
              <View style={styles.body}>
                <ThemedText type="smallBold">{t('folders.rename')}</ThemedText>
                <Input
                  value={draftName}
                  onChangeText={setDraftName}
                  autoFocus
                  returnKeyType="done"
                  onSubmitEditing={() => void handleRename()}
                  editable={!isMutating}
                />
                <View style={styles.actionsRow}>
                  <Pressable onPress={() => setRenaming(false)} hitSlop={8} disabled={isMutating}>
                    <ThemedText type="link" style={{ color: theme.muted }}>
                      {t('common.cancel')}
                    </ThemedText>
                  </Pressable>
                  <Pressable onPress={() => void handleRename()} hitSlop={8} disabled={isMutating}>
                    <ThemedText type="link" style={{ color: theme.accent }}>
                      {t('common.save')}
                    </ThemedText>
                  </Pressable>
                </View>
              </View>
            ) : (
              <View style={styles.body}>
                <ThemedText type="smallBold" numberOfLines={1}>
                  {folder.name}
                </ThemedText>
                <View style={styles.menu}>
                  <ActionRow
                    icon={<PencilIcon size={18} color={theme.foreground} strokeWidth={1.5} />}
                    label={t('folders.rename')}
                    onPress={openRename}
                  />
                  <ActionRow
                    icon={<PlusIcon size={18} color={theme.foreground} strokeWidth={2} />}
                    label={t('folders.newSubfolder')}
                    onPress={handleSubfolder}
                  />
                  <ActionRow
                    icon={<TrashIcon size={18} color={theme.error} strokeWidth={1.5} />}
                    label={t('folders.delete')}
                    onPress={confirmDelete}
                    danger
                    theme={theme}
                  />
                </View>
              </View>
            )}
          </Card>
        </Pressable>
      </Pressable>
    </Modal>
  );
}

function ActionRow({
  icon,
  label,
  onPress,
  danger,
  theme,
}: {
  icon: React.ReactNode;
  label: string;
  onPress: () => void;
  danger?: boolean;
  theme?: ReturnType<typeof useTheme>;
}) {
  return (
    <Pressable
      onPress={onPress}
      style={({ pressed }) => pressed && styles.pressed}
      accessibilityRole="button"
      accessibilityLabel={label}
    >
      <View style={styles.actionRow}>
        <View style={styles.actionIcon}>{icon}</View>
        <ThemedText
          type="default"
          style={danger && theme ? { color: theme.error } : undefined}
        >
          {label}
        </ThemedText>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  backdrop: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.45)',
    justifyContent: 'flex-end',
  },
  sheet: {
    width: '100%',
    borderBottomLeftRadius: 0,
    borderBottomRightRadius: 0,
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
  },
  body: {
    gap: Spacing.three,
  },
  menu: {
    gap: Spacing.one,
  },
  actionsRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  actionRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Spacing.three,
    paddingVertical: Spacing.two,
  },
  actionIcon: {
    width: 28,
    alignItems: 'center',
  },
  pressed: {
    opacity: 0.7,
  },
});
