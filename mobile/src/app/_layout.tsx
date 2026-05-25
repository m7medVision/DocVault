import { DarkTheme, DefaultTheme, ThemeProvider } from '@react-navigation/native';
import { HeroUINativeProvider, type HeroUINativeConfig } from 'heroui-native';
import { Slot, useRouter, usePathname } from 'expo-router';
import React, { useEffect } from 'react';
import { useColorScheme } from 'react-native';
import { GestureHandlerRootView } from 'react-native-gesture-handler';

import '@/global.css';
import { AnimatedSplashOverlay } from '@/components/animated-icon';
import { AuthProvider, useAuth } from '@/lib/auth/auth-context';

const heroUIConfig: HeroUINativeConfig = {
  textProps: {
    maxFontSizeMultiplier: 1.5,
    minimumFontScale: 0.5,
  },
};

const PUBLIC_ROUTES = ['/login', '/register'];

function AuthGuard({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (isLoading) return;

    const isPublic = PUBLIC_ROUTES.some((route) => pathname.startsWith(route));

    if (!isAuthenticated && !isPublic) {
      router.replace('/login');
    }

    if (isAuthenticated && isPublic) {
      router.replace('/');
    }
  }, [isAuthenticated, isLoading, pathname, router]);

  if (isLoading) return null;

  return <>{children}</>;
}

export default function RootLayout() {
  const colorScheme = useColorScheme();

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <HeroUINativeProvider config={heroUIConfig}>
        <ThemeProvider value={colorScheme === 'dark' ? DarkTheme : DefaultTheme}>
          <AuthProvider>
            <AuthGuard>
              <AnimatedSplashOverlay />
              <Slot />
            </AuthGuard>
          </AuthProvider>
        </ThemeProvider>
      </HeroUINativeProvider>
    </GestureHandlerRootView>
  );
}
