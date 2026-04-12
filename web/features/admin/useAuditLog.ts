'use client';

import { useState, useEffect, useCallback } from 'react';
import { getAuditLog, type AuditLogOptions } from './api';
import type { AuditFilters, NormalizedAuditEvent } from './types';

export interface UseAuditLogReturn {
  events: NormalizedAuditEvent[];
  loading: boolean;
  error: string | null;
  filters: AuditFilters;
  setFilter: (key: keyof AuditFilters, value: string) => void;
}

export function useAuditLog(): UseAuditLogReturn {
  const [events, setEvents] = useState<NormalizedAuditEvent[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filters, setFilters] = useState<AuditFilters>({
    entity_type: 'all',
    action: 'all',
  });

  const loadAuditLog = useCallback(async () => {
    try {
      setIsLoading(true);
      const options: AuditLogOptions = {
        entity_type: filters.entity_type !== 'all' ? filters.entity_type : undefined,
        action: filters.action !== 'all' ? filters.action : undefined,
      };
      const response = await getAuditLog(options);
      setEvents(
        response.events.map((event) => ({
          id: event.id,
          entity_type: event.entity_type,
          entity_id: event.entity_id,
          action: event.action,
          actor_id: event.actor_id,
          metadata: event.metadata,
          created_at: event.created_at,
        }))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load audit log');
    } finally {
      setIsLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    loadAuditLog();
  }, [loadAuditLog]);

  const handleFilterChange = (key: keyof AuditFilters, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
  };

  return {
    events,
    loading: isLoading,
    error,
    filters,
    setFilter: handleFilterChange,
  };
}
