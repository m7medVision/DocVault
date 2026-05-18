export interface Reminder {
  id: string;
  document_id: string;
  document_title?: string;
  tenant_id: string;
  rule_type: string;
  trigger_date: string;
  notify_days_before: number[];
  source: string;
  active: boolean;
  status?: string;
  days_until?: number;
  created_at: string;
}
