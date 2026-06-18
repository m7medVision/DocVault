import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  addGroupMember,
  createGroup,
  deleteGroup,
  listGroups,
  removeGroupMember,
} from '@/lib/api/acl';
import { listMembers } from '@/lib/api/admin';
import { useAuth } from '@/lib/useAuth';

export function useGroups() {
  const { isAuthenticated } = useAuth();
  return useQuery({
    queryKey: ['groups'],
    queryFn: listGroups,
    enabled: isAuthenticated,
  });
}

export function useOrgMembers() {
  const { isAuthenticated } = useAuth();
  return useQuery({
    queryKey: ['orgMembers'],
    queryFn: listMembers,
    enabled: isAuthenticated,
  });
}

export function useGroupMutations() {
  const queryClient = useQueryClient();
  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ['groups'] });

  const create = useMutation({
    mutationFn: (name: string) => createGroup(name),
    onSuccess: invalidate,
  });

  const remove = useMutation({
    mutationFn: (id: string) => deleteGroup(id),
    onSuccess: invalidate,
  });

  // The backend exposes no list-members endpoint, so add/remove are
  // fire-and-forget against the org member list (both are idempotent server-side).
  const addMember = useMutation({
    mutationFn: ({ groupId, userId }: { groupId: string; userId: string }) =>
      addGroupMember(groupId, userId),
  });

  const removeMember = useMutation({
    mutationFn: ({ groupId, userId }: { groupId: string; userId: string }) =>
      removeGroupMember(groupId, userId),
  });

  return { create, remove, addMember, removeMember };
}
