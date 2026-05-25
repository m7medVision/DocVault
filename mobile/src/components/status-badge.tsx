import { StyleSheet, View } from 'react-native';

import { ThemedText } from './themed-text';
import { Spacing } from '@/constants/theme';
import { useTranslation } from '@/lib/i18n';

interface StatusBadgeProps {
  status: string;
}

const STATUS_COLORS: Record<string, string> = {
  processed: 'success',
  failed: 'danger',
  pending: 'warning',
  processing: 'warning',
};

export function StatusBadge({ status }: StatusBadgeProps) {
  const { t } = useTranslation();
  const label = t(`status.${status}`) || status;
  const color = STATUS_COLORS[status] || 'textSecondary';

  return (
    <View style={styles.badge}>
      <ThemedText type="code" themeColor={color as 'success' | 'danger' | 'warning' | 'textSecondary'}>
        {label}
      </ThemedText>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    alignSelf: 'flex-start',
    paddingHorizontal: Spacing.two,
    paddingVertical: Spacing.half,
    borderRadius: 999,
  },
});
