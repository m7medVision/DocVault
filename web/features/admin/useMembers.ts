'use client';

import { useState, useEffect, useCallback } from 'react';
import { listMembers, inviteMember, updateMemberRole, type Member } from './api';
import type { NormalizedMember } from './types';

export interface UseMembersReturn {
  members: NormalizedMember[];
  loading: boolean;
  error: string | null;
  filter: string;
  setFilter: (filter: string) => void;
  handleInvite: () => Promise<void>;
  handleEditRole: (member: NormalizedMember) => Promise<void>;
}

export function useMembers(): UseMembersReturn {
  const [members, setMembers] = useState<NormalizedMember[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<string>('all');

  const loadMembers = useCallback(async () => {
    try {
      setIsLoading(true);
      const response = await listMembers();
      setMembers(
        response.members.map((member: Member) => ({
          id: member.membership_id,
          email: member.email,
          name: member.display_name,
          role: member.role === 'owner' ? 'admin' : (member.role as NormalizedMember['role']),
          status: 'active' as const,
          joined_at: member.created_at,
        }))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load members');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadMembers();
  }, [loadMembers]);

  const handleInvite = async () => {
    const email = window.prompt('Invite member email');
    if (!email) {
      return;
    }

    const role = window.prompt('Role: admin, member, or viewer', 'member');
    if (!role || !['admin', 'member', 'viewer'].includes(role)) {
      setError('Invalid role. Use admin, member, or viewer.');
      return;
    }

    await inviteMember(email, role as NormalizedMember['role']);
    await loadMembers();
  };

  const handleEditRole = async (member: NormalizedMember) => {
    const role = window.prompt(
      `Update role for ${member.email}: admin, member, or viewer`,
      member.role
    );
    if (!role || !['admin', 'member', 'viewer'].includes(role)) {
      return;
    }

    await updateMemberRole(member.id, role as NormalizedMember['role']);
    setMembers((prev) =>
      prev.map((item) =>
        item.id === member.id ? { ...item, role: role as NormalizedMember['role'] } : item
      )
    );
  };

  return {
    members,
    loading: isLoading,
    error,
    filter,
    setFilter,
    handleInvite,
    handleEditRole,
  };
}
