import { StyleSheet, View } from 'react-native';
import { Card } from 'heroui-native';

import { DocVaultScreen } from '@/components/docvault-screen';
import { ThemedText } from '@/components/themed-text';
import { Spacing } from '@/constants/theme';

export default function RemindersScreen() {
  return (
    <DocVaultScreen>
      <View style={styles.header}>
        <ThemedText type="subtitle">Reminders</ThemedText>
        <ThemedText type="small" themeColor="textSecondary">
          Track renewals, due dates, expirations, and dismissed alerts.
        </ThemedText>
      </View>

      <Card className="rounded-3xl border border-divider bg-content1 p-6">
        <View style={styles.emptyCard}>
          <ThemedText type="smallBold">No reminders loaded</ThemedText>
          <ThemedText type="small" themeColor="textSecondary">
            The next step is wiring this screen to the reminders API and actions.
          </ThemedText>
        </View>
      </Card>
    </DocVaultScreen>
  );
}

const styles = StyleSheet.create({
  header: {
    gap: Spacing.one,
  },
  emptyCard: {
    gap: Spacing.one,
    alignItems: 'center',
  },
});
