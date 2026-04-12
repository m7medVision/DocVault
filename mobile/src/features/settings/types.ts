export interface UserSettings {
  language: string;
  notificationsEnabled: boolean;
  emailNotifications: boolean;
  pushNotifications: boolean;
  reminderLeadTime: number;
  theme: 'light' | 'dark' | 'system';
}

export interface UseSettingsReturn {
  settings: UserSettings;
  updateSettings: (updates: Partial<UserSettings>) => void;
  isLoading: boolean;
}
