import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { Button, Card, Input } from 'heroui-native';

import { ThemedText } from './themed-text';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';
import { useTranslation } from '@/lib/i18n';

interface FolderCreateInputProps {
  onSubmit: (name: string) => Promise<void> | void;
  onCancel: () => void;
  busy?: boolean;
  autoFocus?: boolean;
  placeholderKey?: string;
}

export function FolderCreateInput({
  onSubmit,
  onCancel,
  busy,
  autoFocus = true,
  placeholderKey = 'folders.folderNamePlaceholder',
}: FolderCreateInputProps) {
  const { t } = useTranslation();
  const theme = useTheme();
  const [name, setName] = useState('');

  async function handleSubmit() {
    const trimmed = name.trim();
    if (!trimmed) return;
    await onSubmit(trimmed);
    setName('');
  }

  function handleCancel() {
    setName('');
    onCancel();
  }

  return (
    <Card className="rounded-2xl border border-divider bg-content1 p-3">
      <View style={styles.body}>
        <Input
          placeholder={t(placeholderKey)}
          value={name}
          onChangeText={setName}
          autoFocus={autoFocus}
          returnKeyType="done"
          onSubmitEditing={() => void handleSubmit()}
          editable={!busy}
        />
        <View style={styles.actions}>
          <Pressable onPress={handleCancel} disabled={busy} hitSlop={8}>
            <ThemedText type="link" style={{ color: theme.muted }}>
              {t('common.cancel')}
            </ThemedText>
          </Pressable>
          <Button
            variant="primary"
            size="sm"
            onPress={() => void handleSubmit()}
            isDisabled={busy || !name.trim()}
          >
            <Button.Label>{t('common.save')}</Button.Label>
          </Button>
        </View>
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  body: {
    gap: Spacing.two,
  },
  actions: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: Spacing.three,
  },
});
