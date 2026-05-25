import { StyleSheet, View } from 'react-native';
import { Card } from 'heroui-native';

import { ThemedText } from './themed-text';
import { Spacing } from '@/constants/theme';

interface StatCardProps {
  value: string | number;
  label: string;
  loading?: boolean;
  icon?: React.ReactNode;
}

export function StatCard({ value, label, loading, icon }: StatCardProps) {
  return (
    <Card className="rounded-2xl border border-divider bg-content1 px-3 py-3">
      <View style={styles.card}>
        {icon && <View style={styles.iconWrap}>{icon}</View>}
        <ThemedText type="title" style={styles.value}>
          {loading ? '—' : value}
        </ThemedText>
        <ThemedText type="small" themeColor="textSecondary">
          {label}
        </ThemedText>
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  card: {
    gap: Spacing.half,
    alignItems: 'flex-start',
  },
  iconWrap: {
    marginBottom: Spacing.half,
  },
  value: {
    fontSize: 28,
    lineHeight: 34,
    fontWeight: 700,
  },
});
