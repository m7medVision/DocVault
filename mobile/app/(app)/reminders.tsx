import React, { useState } from 'react';
import {
  View,
  Text,
  FlatList,
  StyleSheet,
  TouchableOpacity,
  RefreshControl,
  Alert,
  Modal,
  Pressable,
} from 'react-native';
import { useTranslation } from 'react-i18next';
import { useReminders } from '../../src/features/reminders/useReminders';
import {
  Reminder,
  formatDueDate,
  getReminderStatusInfo,
  SNOOZE_OPTIONS,
} from '../../src/features/reminders/api';
import { tokens } from '../../src/theme/tokens';

type FilterType = 'all' | 'pending' | 'overdue';

export default function RemindersScreen() {
  const { t } = useTranslation();
  const { reminders, loading, refreshing, onRefresh, snooze, dismiss, complete } = useReminders();
  const [filter, setFilter] = useState<FilterType>('all');
  const [snoozeModalVisible, setSnoozeModalVisible] = useState(false);
  const [selectedReminder, setSelectedReminder] = useState<Reminder | null>(null);
  const [actionLoading, setActionLoading] = useState(false);

  const filteredReminders = reminders.filter((reminder) => {
    switch (filter) {
      case 'pending':
        return reminder.status === 'pending';
      case 'overdue':
        return new Date(reminder.due_date) < new Date() && reminder.status === 'pending';
      default:
        return true;
    }
  });

  const sortedReminders = [...filteredReminders].sort((a, b) => {
    const aDate = new Date(a.due_date);
    const bDate = new Date(b.due_date);
    const now = new Date();

    const aOverdue = aDate < now && a.status === 'pending';
    const bOverdue = bDate < now && b.status === 'pending';

    if (aOverdue && !bOverdue) return -1;
    if (!aOverdue && bOverdue) return 1;

    return aDate.getTime() - bDate.getTime();
  });

  const handleSnooze = async (minutes: number) => {
    if (!selectedReminder) return;

    setActionLoading(true);
    try {
      await snooze(selectedReminder.id, { snooze_minutes: minutes });
      setSnoozeModalVisible(false);
      setSelectedReminder(null);
    } catch {
      Alert.alert('Error', 'Failed to snooze reminder');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDismiss = (reminder: Reminder) => {
    Alert.alert(
      'Dismiss Reminder',
      `Are you sure you want to dismiss the reminder for "${reminder.document_title}"?`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Dismiss',
          style: 'destructive',
          onPress: async () => {
            setActionLoading(true);
            try {
              await dismiss(reminder.id);
            } catch {
              Alert.alert('Error', 'Failed to dismiss reminder');
            } finally {
              setActionLoading(false);
            }
          },
        },
      ]
    );
  };

  const handleComplete = async (reminder: Reminder) => {
    setActionLoading(true);
    try {
      await complete(reminder.id);
    } catch {
      Alert.alert('Error', 'Failed to complete reminder');
    } finally {
      setActionLoading(false);
    }
  };

  const openSnoozeModal = (reminder: Reminder) => {
    setSelectedReminder(reminder);
    setSnoozeModalVisible(true);
  };

  const renderReminderItem = ({ item }: { item: Reminder }) => {
    const statusInfo = getReminderStatusInfo(item.status);
    const isOverdue = new Date(item.due_date) < new Date() && item.status === 'pending';
    const isActive = item.status === 'pending';

    return (
      <View style={[styles.card, isOverdue && styles.cardOverdue]}>
        <View style={styles.cardHeader}>
          <Text style={styles.documentTitle} numberOfLines={2}>
            {item.document_title}
          </Text>
          <View style={[styles.statusBadge, { backgroundColor: statusInfo.bgColor }]}>
            <Text style={[styles.statusText, { color: statusInfo.color }]}>
              {isOverdue ? t('overdueReminder') : statusInfo.label}
            </Text>
          </View>
        </View>

        <View style={styles.dueDateContainer}>
          <Text style={styles.dueDateLabel}>Due</Text>
          <Text style={[styles.dueDateValue, isOverdue && styles.dueDateOverdue]}>
            {formatDueDate(item.due_date)}
          </Text>
        </View>

        {item.snoozed_count > 0 && (
          <Text style={styles.snoozeInfo}>
            Snoozed {item.snoozed_count} time{item.snoozed_count > 1 ? 's' : ''}
          </Text>
        )}

        {isActive && (
          <View style={styles.actions}>
            <TouchableOpacity
              style={styles.snoozeButton}
              onPress={() => openSnoozeModal(item)}
              disabled={actionLoading}
            >
              <Text style={styles.snoozeButtonText}>{t('snooze')}</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.dismissButton}
              onPress={() => handleDismiss(item)}
              disabled={actionLoading}
            >
              <Text style={styles.dismissButtonText}>{t('dismiss')}</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.completeButton}
              onPress={() => handleComplete(item)}
              disabled={actionLoading}
            >
              <Text style={styles.completeButtonText}>Done</Text>
            </TouchableOpacity>
          </View>
        )}
      </View>
    );
  };

  const FilterTabs = () => (
    <View style={styles.filterContainer}>
      {(['all', 'pending', 'overdue'] as FilterType[]).map((f) => (
        <TouchableOpacity
          key={f}
          style={[styles.filterTab, filter === f && styles.filterTabActive]}
          onPress={() => setFilter(f)}
        >
          <Text
            style={[
              styles.filterTabText,
              filter === f && styles.filterTabTextActive,
            ]}
          >
            {f.charAt(0).toUpperCase() + f.slice(1)}
          </Text>
        </TouchableOpacity>
      ))}
    </View>
  );

  const EmptyState = () => (
    <View style={styles.emptyContainer}>
      <Text style={styles.emptyText}>
        {filter === 'all'
          ? t('noReminders')
          : filter === 'overdue'
          ? t('noOverdueReminders')
          : t('noPendingReminders')}
      </Text>
      <Text style={styles.emptySubtext}>
        {filter === 'all'
          ? t('setRemindersForDocuments')
          : t('allCaughtUp')}
      </Text>
    </View>
  );

  return (
    <View style={styles.container}>
      <FilterTabs />

      <FlatList
        data={sortedReminders}
        renderItem={renderReminderItem}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.listContent}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor={tokens.accentColor}
          />
        }
        ListEmptyComponent={!loading ? <EmptyState /> : null}
      />

      <Modal
        visible={snoozeModalVisible}
        transparent
        animationType="fade"
        onRequestClose={() => setSnoozeModalVisible(false)}
      >
        <Pressable
          style={styles.modalOverlay}
          onPress={() => setSnoozeModalVisible(false)}
        >
          <View style={styles.modalContent}>
            <Text style={styles.modalTitle}>{t('snoozeReminder')}</Text>
            <Text style={styles.modalSubtitle} numberOfLines={1}>
              {selectedReminder?.document_title}
            </Text>

            {SNOOZE_OPTIONS.map((option) => (
              <TouchableOpacity
                key={option.value}
                style={styles.snoozeOption}
                onPress={() => handleSnooze(option.value)}
                disabled={actionLoading}
              >
                <Text style={styles.snoozeOptionText}>{option.label}</Text>
              </TouchableOpacity>
            ))}

            <TouchableOpacity
              style={styles.modalCancelButton}
              onPress={() => setSnoozeModalVisible(false)}
            >
              <Text style={styles.modalCancelText}>Cancel</Text>
            </TouchableOpacity>
          </View>
        </Pressable>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: tokens.backgroundColor.secondary,
  },
  filterContainer: {
    flexDirection: 'row',
    padding: tokens.screenPadding,
    paddingBottom: 0,
    gap: 8,
  },
  filterTab: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: tokens.backgroundColor.tertiary,
  },
  filterTabActive: {
    backgroundColor: tokens.accentColor,
  },
  filterTabText: {
    color: tokens.textColor.secondary,
    fontSize: 14,
    fontWeight: '500',
  },
  filterTabTextActive: {
    color: '#ffffff',
  },
  listContent: {
    flexGrow: 1,
    padding: tokens.screenPadding,
    paddingTop: 16,
  },
  card: {
    backgroundColor: tokens.backgroundColor.tertiary,
    borderRadius: tokens.cardRadius,
    padding: tokens.cardPadding,
    marginBottom: 12,
    ...tokens.cardShadow,
  },
  cardOverdue: {
    borderLeftWidth: 3,
    borderLeftColor: tokens.errorColor,
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  documentTitle: {
    flex: 1,
    fontSize: 16,
    fontWeight: '600',
    color: tokens.textColor.primary,
    marginRight: 12,
  },
  statusBadge: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  statusText: {
    fontSize: 12,
    fontWeight: '600',
  },
  dueDateContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  dueDateLabel: {
    fontSize: 14,
    color: tokens.textColor.muted,
    marginRight: 8,
  },
  dueDateValue: {
    fontSize: 14,
    color: tokens.textColor.secondary,
    fontWeight: '500',
  },
  dueDateOverdue: {
    color: tokens.errorColor,
  },
  snoozeInfo: {
    fontSize: 12,
    color: tokens.textColor.muted,
    fontStyle: 'italic',
  },
  actions: {
    flexDirection: 'row',
    marginTop: 16,
    gap: 8,
  },
  snoozeButton: {
    flex: 1,
    paddingVertical: 10,
    borderRadius: 8,
    backgroundColor: tokens.backgroundColor.secondary,
    alignItems: 'center',
  },
  snoozeButtonText: {
    color: tokens.textColor.secondary,
    fontSize: 14,
    fontWeight: '500',
  },
  dismissButton: {
    flex: 1,
    paddingVertical: 10,
    borderRadius: 8,
    backgroundColor: 'rgba(107, 114, 128, 0.2)',
    alignItems: 'center',
  },
  dismissButtonText: {
    color: tokens.textColor.secondary,
    fontSize: 14,
    fontWeight: '500',
  },
  completeButton: {
    flex: 1,
    paddingVertical: 10,
    borderRadius: 8,
    backgroundColor: tokens.accentColor,
    alignItems: 'center',
  },
  completeButtonText: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '600',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingTop: 100,
  },
  emptyText: {
    color: tokens.textColor.secondary,
    fontSize: 18,
    fontWeight: '600',
  },
  emptySubtext: {
    color: tokens.textColor.muted,
    fontSize: 14,
    marginTop: 8,
    textAlign: 'center',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  modalContent: {
    backgroundColor: tokens.backgroundColor.tertiary,
    borderRadius: 16,
    padding: 24,
    width: '100%',
    maxWidth: 320,
  },
  modalTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: tokens.textColor.primary,
    marginBottom: 4,
    textAlign: 'center',
  },
  modalSubtitle: {
    fontSize: 14,
    color: tokens.textColor.muted,
    marginBottom: 20,
    textAlign: 'center',
  },
  snoozeOption: {
    paddingVertical: 14,
    borderBottomWidth: 1,
    borderBottomColor: tokens.backgroundColor.secondary,
  },
  snoozeOptionText: {
    fontSize: 16,
    color: tokens.textColor.primary,
    textAlign: 'center',
  },
  modalCancelButton: {
    marginTop: 16,
    paddingVertical: 14,
    alignItems: 'center',
  },
  modalCancelText: {
    fontSize: 16,
    color: tokens.accentColor,
    fontWeight: '600',
  },
});
