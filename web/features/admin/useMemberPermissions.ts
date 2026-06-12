import { useQuery } from '@tanstack/react-query';
import { getMemberPermissions } from '@/lib/api/casbin';
import { useAuth } from '@/lib/useAuth';

export function useMemberPermissions(membershipId: string) {
  const { isAuthenticated } = useAuth();
  return useQuery({
    queryKey: ['memberPermissions', membershipId],
    queryFn: () => getMemberPermissions(membershipId),
    enabled: isAuthenticated && !!membershipId,
  });
}
