import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  listPolicies,
  addPolicy,
  deletePolicy,
  type PolicyInput,
} from '@/lib/api/casbin';
import { useAuth } from '@/lib/useAuth';

export function usePolicies() {
  const { isAuthenticated } = useAuth();
  return useQuery({
    queryKey: ['casbinPolicies'],
    queryFn: listPolicies,
    enabled: isAuthenticated,
  });
}

export function usePolicyMutations() {
  const queryClient = useQueryClient();
  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ['casbinPolicies'] });

  const add = useMutation({
    mutationFn: (input: PolicyInput) => addPolicy(input),
    onSuccess: invalidate,
  });

  const remove = useMutation({
    mutationFn: (input: PolicyInput) => deletePolicy(input),
    onSuccess: invalidate,
  });

  return { add, remove };
}
