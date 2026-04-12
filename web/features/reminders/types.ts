export interface NormalizedReminder {
  id: string;
  document_id: string;
  document_title: string;
  rule_type: string;
  trigger_date: string;
  days_until: number;
  status: 'pending' | 'sent' | 'failed';
}

export { type Reminder } from '@/lib/api/reminders';
