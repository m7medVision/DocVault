export interface AuditFilters {
  entity_type: string;
  action: string;
}

export interface NormalizedAuditEvent {
  id: string;
  entity_type: string;
  entity_id: string;
  action: string;
  actor_id?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface NormalizedMember {
  id: string;
  email: string;
  name: string;
  role: 'admin' | 'member' | 'viewer';
  status: 'active' | 'invited' | 'disabled';
  joined_at: string;
}

export { type AuditEvent, type AuditLogOptions, type Member } from '@/lib/api/admin';
