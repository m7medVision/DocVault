import { useMemo, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import { listReminders, setReminderActive } from './api';
import { listDocuments } from '@/features/documents/api';
import type { Reminder } from './types';

export type ReminderFilter = 'active' | 'all';

export function useReminders() {
  const queryClient = useQueryClient();
  const [filter, setFilter] = useState<ReminderFilter>('active');

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['reminders'],
    queryFn: async () => {
      const [reminderRes, documentRes] = await Promise.all([
        listReminders(),
        listDocuments({ limit: 100 }).catch(() => ({ documents: [], total: 0 })),
      ]);

      const titles = new Map(documentRes.documents.map((d) => [d.id, d.title]));
      const now = Date.now();

      const reminders: Reminder[] = reminderRes.reminders.map((r) => ({
        ...r,
        document_title: titles.get(r.document_id) ?? r.document_title,
        days_until: Math.ceil(
          (new Date(r.trigger_date).getTime() - now) / (1000 * 60 * 60 * 24),
        ),
      }));

      return { reminders };
    },
  });

  const setActiveMutation = useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) =>
      setReminderActive(id, active),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['reminders'] });
    },
  });

  const reminders = useMemo(() => {
    const all = data?.reminders ?? [];
    return filter === 'active' ? all.filter((r) => r.active) : all;
  }, [data, filter]);

  return {
    reminders,
    filter,
    setFilter,
    loading: isLoading,
    error: error ? (error instanceof Error ? error.message : 'Failed to load reminders') : null,
    reload: refetch,
    dismiss: (id: string) => setActiveMutation.mutateAsync({ id, active: false }),
    reactivate: (id: string) => setActiveMutation.mutateAsync({ id, active: true }),
  };
}
