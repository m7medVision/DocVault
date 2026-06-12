import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

import { getAuditLog, listMembers, listPolicies, updateMemberRole } from './api';

export function useAuditLog() {
  return useQuery({
    queryKey: ['admin', 'audit'],
    queryFn: () => getAuditLog(),
  });
}

export function useMembers() {
  return useQuery({
    queryKey: ['admin', 'members'],
    queryFn: listMembers,
  });
}

export function usePolicies() {
  return useQuery({
    queryKey: ['admin', 'policies'],
    queryFn: listPolicies,
  });
}

export function useUpdateMemberRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, role }: { id: string; role: 'admin' | 'member' | 'viewer' }) =>
      updateMemberRole(id, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'members'] });
    },
  });
}
