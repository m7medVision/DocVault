import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import { listReminders, dismissReminder, snoozeReminder } from './api';

export function useReminders() {
  const queryClient = useQueryClient();

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['reminders'],
    queryFn: listReminders,
  });

  const dismissMutation = useMutation({
    mutationFn: dismissReminder,
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: ['reminders'] });
      const previous = queryClient.getQueryData<{ reminders: typeof data extends { reminders: infer R } ? R : never }>(['reminders']);
      queryClient.setQueryData(['reminders'], (old: { reminders: { id: string }[] } | undefined) => {
        if (!old) return old;
        return { ...old, reminders: old.reminders.filter((r) => r.id !== id) };
      });
      return { previous };
    },
    onError: (_err, _id, context) => {
      if (context?.previous) {
        queryClient.setQueryData(['reminders'], context.previous);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['reminders'] });
    },
  });

  const snoozeMutation = useMutation({
    mutationFn: ({ id, minutes }: { id: string; minutes: number }) => snoozeReminder(id, minutes),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['reminders'] });
    },
  });

  return {
    reminders: data?.reminders ?? [],
    loading: isLoading,
    error: error ? (error instanceof Error ? error.message : 'Failed to load reminders') : null,
    reload: refetch,
    dismiss: (id: string) => dismissMutation.mutateAsync(id),
    snooze: (id: string, minutes: number) => snoozeMutation.mutateAsync({ id, minutes }),
  };
}
