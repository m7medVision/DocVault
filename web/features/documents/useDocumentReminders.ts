'use client';

import { useState, useEffect, useCallback } from 'react';
import { listReminders, type Reminder } from '@/lib/api/reminders';

export interface DocumentReminder {
  id: string;
  document_id: string;
  rule_type: string;
  trigger_date: string;
  days_until: number;
  status: 'pending' | 'sent' | 'failed';
}

export interface UseDocumentRemindersResult {
  reminders: DocumentReminder[];
  loading: boolean;
  error: string | null;
}

export function useDocumentReminders(documentId: string): UseDocumentRemindersResult {
  const [reminders, setReminders] = useState<DocumentReminder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadReminders = useCallback(async () => {
    if (!documentId) return;

    try {
      setLoading(true);
      const response = await listReminders();
      const now = Date.now();

      const filteredReminders: DocumentReminder[] = response.reminders
        .filter((r: Reminder) => r.document_id === documentId)
        .map((reminder: Reminder) => ({
          id: reminder.id,
          document_id: reminder.document_id,
          rule_type: reminder.rule_type,
          trigger_date: reminder.trigger_date,
          days_until: Math.ceil(
            (new Date(reminder.trigger_date).getTime() - now) / (1000 * 60 * 60 * 24)
          ),
          status: reminder.active ? ('pending' as const) : ('sent' as const),
        }));

      setReminders(filteredReminders);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load reminders');
    } finally {
      setLoading(false);
    }
  }, [documentId]);

  useEffect(() => {
    void loadReminders();
  }, [loadReminders]);

  return {
    reminders,
    loading,
    error,
  };
}
