// Protected app layout - requires authentication
// All screens under (app)/ require valid authentication
// Uses shared design tokens from /shared/theme
// i18next integration for RTL support and translations

import { useEffect } from 'react';
import { View, ActivityIndicator, StyleSheet, I18nManager } from 'react-native';
import { Stack, useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../../src/contexts/AuthContext';
import { isRTL, changeLanguage } from '../../src/i18n';
import { tokens } from '../../src/theme/tokens';

export default function AppLayout() {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();
  const { t } = useTranslation();

  // Handle RTL when language changes
  useEffect(() => {
    const rtl = isRTL();
    if (I18nManager.isRTL !== rtl) {
      I18nManager.allowRTL(rtl);
      I18nManager.forceRTL(rtl);
    }
  }, []);

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.replace('/login');
    }
  }, [isLoading, isAuthenticated]);

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={tokens.accentColor} />
      </View>
    );
  }

  if (!isAuthenticated) {
    return null; // Will redirect
  }

  return (
    <Stack
      screenOptions={{
        headerStyle: {
          backgroundColor: tokens.backgroundColor.primary,
        },
        headerTintColor: tokens.textColor.primary,
        headerTitleStyle: {
          fontWeight: '600',
        },
        contentStyle: {
          backgroundColor: tokens.backgroundColor.secondary,
        },
      }}
    >
      <Stack.Screen
        name="documents"
        options={{
          title: t('documents'),
          headerLargeTitle: true,
        }}
      />
      <Stack.Screen
        name="search"
        options={{
          title: t('search'),
          headerLargeTitle: true,
        }}
      />
      <Stack.Screen
        name="reminders"
        options={{
          title: t('reminders'),
          headerLargeTitle: true,
        }}
      />
      <Stack.Screen
        name="settings"
        options={{
          title: t('settings'),
        }}
      />
    </Stack>
  );
}

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: tokens.backgroundColor.primary,
  },
});
