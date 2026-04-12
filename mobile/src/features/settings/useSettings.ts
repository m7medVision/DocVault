import { useState, useCallback } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { changeLanguage, isRTL } from '@/i18n';
import type { UserSettings } from './types';

const DEFAULT_SETTINGS: UserSettings = {
  language: 'en',
  notificationsEnabled: true,
  emailNotifications: true,
  pushNotifications: true,
  reminderLeadTime: 60,
  theme: 'system',
};

export function useSettings() {
  const { user, logout } = useAuth();
  const [settings, setSettings] = useState<UserSettings>({
    ...DEFAULT_SETTINGS,
    language: isRTL() ? 'ar' : 'en',
  });
  const [languageModalVisible, setLanguageModalVisible] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const updateSettings = useCallback((updates: Partial<UserSettings>) => {
    setSettings((prev) => ({ ...prev, ...updates }));
  }, []);

  const changeLanguageSetting = useCallback(async (langCode: string) => {
    setIsLoading(true);
    try {
      await changeLanguage(langCode);
      setSettings((prev) => ({ ...prev, language: langCode }));
    } catch (error) {
      console.error('Failed to change language:', error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const handleLogout = useCallback(async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Logout error:', error);
    }
  }, [logout]);

  return {
    user,
    settings,
    languageModalVisible,
    isLoading,
    updateSettings,
    setLanguageModalVisible,
    changeLanguageSetting,
    handleLogout,
  };
}
