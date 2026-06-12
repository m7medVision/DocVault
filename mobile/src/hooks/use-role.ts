import { useAuth } from '@/lib/auth/auth-context';

export function useRole() {
  const { user } = useAuth();
  const role = user?.role;
  const isAdmin = role === 'owner' || role === 'admin';
  return { role, isAdmin };
}
