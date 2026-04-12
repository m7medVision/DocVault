'use client';

import { useState, useEffect, useCallback } from 'react';
import { listReminders, updateReminder, type Reminder } from './api';
import { listDocuments } from '@/lib/api';
import type { NormalizedReminder } from './types';
import { useAuth } from '@/lib/useAuth';

export interface UseRemindersReturn {
  reminders: NormalizedReminder[];
  loading: boolean;
  error: string | null;
  filter: 'all' | 'pending' | 'sent';
  setFilter: (filter: 'all' | 'pending' | 'sent') => void;
  handleDismiss: (id: string) => Promise<void>;
  handleSnooze: (id: string) => void;
}

export function useReminders(): UseRemindersReturn {
  const { isLoading: authLoading, isAuthenticated } = useAuth();
  const [reminders, setReminders] = useState<NormalizedReminder[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'pending' | 'sent'>('all');

  const loadReminders = useCallback(async () => {
    try {
      setIsLoading(true);
      const [reminderResponse, documentResponse] = await Promise.all([
        listReminders(),
        listDocuments({ limit: 100 }),
      ]);

      const documentTitles = new Map(
        documentResponse.documents.map((document) => [document.id, document.title])
      );

      const now = Date.now();
      const normalizedReminders: NormalizedReminder[] = reminderResponse.reminders.map(
        (reminder: Reminder) => ({
          id: reminder.id,
          document_id: reminder.document_id,
          document_title:
            documentTitles.get(reminder.document_id) || reminder.document_id,
          rule_type: reminder.rule_type,
          trigger_date: reminder.trigger_date,
          days_until: Math.ceil(
            (new Date(reminder.trigger_date).getTime() - now) / (1000 * 60 * 60 * 24)
          ),
          status: reminder.active ? ('pending' as const) : ('sent' as const),
        })
      );

      setReminders(normalizedReminders);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load reminders');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (authLoading) {
      return;
    }

    if (!isAuthenticated) {
      setReminders([]);
      setError(null);
      setIsLoading(false);
      return;
    }

    void loadReminders();
  }, [authLoading, isAuthenticated, loadReminders]);

  const handleSnooze = async (id: string) => {
    setError(`Reminder snoozing is not available yet for ${id}.`);
  };

  const handleDismiss = async (id: string) => {
    await updateReminder(id, false);
    setReminders((prev) => prev.filter((r) => r.id !== id));
  };

  return {
    reminders,
    loading: isLoading,
    error,
    filter,
    setFilter,
    handleDismiss,
    handleSnooze,
  };
}
