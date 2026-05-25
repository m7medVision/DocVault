import { Pressable, StyleSheet, View } from 'react-native';
import { Card } from 'heroui-native';

import { ThemedText } from './themed-text';
import { Spacing } from '@/constants/theme';
import { useTranslation } from '@/lib/i18n';
import type { Reminder } from '@/features/reminders/types';

interface ReminderRowProps {
  reminder: Reminder;
  onPress?: () => void;
}

export function ReminderRow({ reminder, onPress }: ReminderRowProps) {
  const { t } = useTranslation();
  const days = reminder.days_until;

  const countdown = (() => {
    if (days === undefined || days === null) return '';
    if (days < 0) return t('reminders.overdue', { days: Math.abs(days) });
    if (days === 0) return t('reminders.dueToday');
    return t('reminders.daysLeft', { days });
  })();

  const ruleLabel = reminder.rule_type
    ? reminder.rule_type.charAt(0).toUpperCase() + reminder.rule_type.slice(1).replace(/_/g, ' ')
    : '';

  return (
    <Pressable onPress={onPress} style={({ pressed }) => pressed && styles.pressed}>
      <Card className="rounded-2xl border border-divider bg-content1 p-3">
        <View style={styles.row}>
          <View style={styles.textCol}>
            <ThemedText type="smallBold" numberOfLines={1}>
              {reminder.document_title || ruleLabel}
            </ThemedText>
            <ThemedText type="small" themeColor="textSecondary" numberOfLines={1}>
              {ruleLabel}
            </ThemedText>
          </View>
          {countdown ? (
            <ThemedText type="code" themeColor={days !== undefined && days <= 3 ? 'warning' : 'textSecondary'}>
              {countdown}
            </ThemedText>
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
    gap: Spacing.two,
  },
  textCol: {
    flex: 1,
    gap: 1,
  },
});
