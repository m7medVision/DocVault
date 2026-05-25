import { useCallback, useEffect, useState } from 'react';

import { listReminders, dismissReminder, snoozeReminder } from './api';
import type { Reminder } from './types';

export function useReminders() {
  const [reminders, setReminders] = useState<Reminder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await listReminders();
      setReminders(response.reminders);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load reminders');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const dismiss = useCallback(async (id: string) => {
    try {
      await dismissReminder(id);
      setReminders((prev) => prev.filter((r) => r.id !== id));
    } catch {}
  }, []);

  const snooze = useCallback(async (id: string, minutes: number) => {
    try {
      await snoozeReminder(id, minutes);
      void load();
    } catch {}
  }, [load]);

  return { reminders, loading, error, reload: load, dismiss, snooze };
}