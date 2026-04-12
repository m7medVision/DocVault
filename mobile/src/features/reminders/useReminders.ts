import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import {
  getReminders,
  snoozeReminder,
  dismissReminder,
  completeReminder,
  Reminder,
  SnoozeReminderRequest,
} from './api';

export function useReminders() {
  const { accessToken } = useAuth();
  const [reminders, setReminders] = useState<Reminder[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const loadReminders = useCallback(async () => {
    if (!accessToken) return;

    try {
      const data = await getReminders(accessToken);
      setReminders(data);
    } catch (error) {
      console.error('Failed to load reminders:', error);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [accessToken]);

  useEffect(() => {
    loadReminders();
  }, [loadReminders]);

  const onRefresh = useCallback(() => {
    setRefreshing(true);
    loadReminders();
  }, [loadReminders]);

  const snooze = useCallback(
    async (reminderId: string, request: SnoozeReminderRequest) => {
      if (!accessToken) return null;
      const updated = await snoozeReminder(accessToken, reminderId, request);
      setReminders((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
      return updated;
    },
    [accessToken]
  );

  const dismiss = useCallback(
    async (reminderId: string) => {
      if (!accessToken) return null;
      const updated = await dismissReminder(accessToken, reminderId);
      setReminders((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
      return updated;
    },
    [accessToken]
  );

  const complete = useCallback(
    async (reminderId: string) => {
      if (!accessToken) return null;
      const updated = await completeReminder(accessToken, reminderId);
      setReminders((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
      return updated;
    },
    [accessToken]
  );

  return {
    reminders,
    loading,
    refreshing,
    onRefresh,
    snooze,
    dismiss,
    complete,
  };
}
