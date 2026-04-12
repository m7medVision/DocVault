import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useEffect, useState } from 'react';
import { ActivityIndicator, I18nManager, View } from 'react-native';
import { AuthProvider } from '../src/contexts/AuthContext';
import { initI18n, isRTL } from '../src/i18n';
import '../src/theme/tokens';
import { tokens } from '../src/theme/tokens';

export default function RootLayout() {
  const [i18nReady, setI18nReady] = useState(false);

  useEffect(() => {
    const init = async () => {
      await initI18n();
      setI18nReady(true);
    };

    void init();
  }, []);

  useEffect(() => {
    if (!i18nReady) {
      return;
    }

    const rtl = isRTL();
    if (I18nManager.isRTL !== rtl) {
      I18nManager.allowRTL(rtl);
      I18nManager.forceRTL(rtl);
    }
  }, [i18nReady]);

  if (!i18nReady) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: tokens.backgroundColor.primary }}>
        <ActivityIndicator size="large" color={tokens.accentColor} />
      </View>
    );
  }

  return (
    <AuthProvider>
      <StatusBar style="light" />
      <Stack
        screenOptions={{
          headerShown: false,
          contentStyle: { backgroundColor: tokens.backgroundColor.primary },
        }}
      >
        <Stack.Screen name="index" />
        <Stack.Screen name="login" />
        <Stack.Screen name="register" />
        <Stack.Screen name="(app)" options={{ headerShown: false }} />
      </Stack>
    </AuthProvider>
  );
}
