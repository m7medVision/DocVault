import { Pressable, StyleSheet, View } from 'react-native';
import { Card } from 'heroui-native';

import { ThemedText } from './themed-text';
import { ChevronRightIcon, FolderIcon } from './icons';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';
import type { Folder } from '@/features/folders/types';

interface FolderRowProps {
  folder: Folder;
  onPress?: () => void;
  onLongPress?: () => void;
  showChevron?: boolean;
}

export function FolderRow({ folder, onPress, onLongPress, showChevron = true }: FolderRowProps) {
  const theme = useTheme();
  const { t } = useTranslation();

  return (
    <Pressable
      onPress={onPress}
      onLongPress={onLongPress}
      delayLongPress={350}
      style={({ pressed }) => pressed && styles.pressed}
      accessibilityRole="button"
      accessibilityLabel={folder.name}
      accessibilityHint={onLongPress ? t('folders.rename') : undefined}
    >
      <Card className="rounded-2xl border border-divider bg-content1 p-3">
        <View style={styles.row}>
          <View style={[styles.iconWrap, { backgroundColor: `${theme.accent}1A` }]}>
            <FolderIcon size={18} color={theme.accent} strokeWidth={1.5} />
          </View>
          <View style={styles.textCol}>
            <ThemedText type="smallBold" numberOfLines={1}>
              {folder.name}
            </ThemedText>
            <ThemedText type="code" themeColor="textSecondary">
              {new Date(folder.created_at).toLocaleDateString()}
            </ThemedText>
          </View>
          {showChevron ? (
            <ChevronRightIcon size={18} color={theme.muted} strokeWidth={1.5} />
          ) : null}
        </View>
      </Card>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  pressed: {
    opacity: 0.75,
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
});
