import React from 'react';
import { Pressable, StyleSheet, Text, View, Platform } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { BottomTabBarProps } from '@react-navigation/bottom-tabs';

import { HomeIcon, DocsIcon, ScanIcon, SearchIcon, SettingsIcon } from './icons';
import { Spacing } from '@/constants/theme';
import { useTheme } from '@/hooks/use-theme';

interface TabConfig {
  label: string;
  name: string;
  href: string;
  Icon: React.FC<{ size?: number; color: string; strokeWidth?: number }>;
  center?: boolean;
}

const TABS: TabConfig[] = [
  { label: 'Home', name: 'index', href: '/', Icon: HomeIcon },
  { label: 'Docs', name: 'documents', href: '/documents', Icon: DocsIcon },
  { label: 'Scan', name: 'scan', href: '/scan', Icon: ScanIcon, center: true },
  { label: 'Search', name: 'search', href: '/search', Icon: SearchIcon },
  { label: 'Settings', name: 'settings', href: '/settings', Icon: SettingsIcon },
];

type TabBarProps = BottomTabBarProps;

export default function TabBar({ state, navigation }: TabBarProps) {
  const theme = useTheme();

  const inactiveColor = theme.muted;
  const activeColor = theme.accent;

  function navigate(tab: TabConfig, index: number) {
    const route = state.routes.find((item) => item.name === tab.name) ?? state.routes[index];
    const event = navigation.emit({ type: 'tabPress', target: route.key, canPreventDefault: true });

    if (event.defaultPrevented) return;

    if (tab.center) {
      navigation.navigate(tab.name, { autoOpen: Date.now().toString() });
    } else {
      navigation.navigate(tab.name);
    }
  }

  return (
    <SafeAreaView edges={['bottom']} style={[styles.safeArea, { backgroundColor: theme.background, borderTopColor: theme.border }]}>
      <View style={styles.bar}>
        {TABS.map((tab, index) => {
          const active = state.routes[state.index]?.name === tab.name;

          if (tab.center) {
            return (
              <Pressable
                key={tab.href}
                hitSlop={16}
                onPress={() => navigate(tab, index)}
                style={styles.centerButton}
                accessibilityRole="button"
                accessibilityLabel={tab.label}
              >
                <View style={[styles.centerCircle, { backgroundColor: theme.accent, shadowColor: theme.accent }]}> 
                  <ScanIcon size={24} color="#fff" strokeWidth={2} />
                </View>
              </Pressable>
            );
          }

          return (
            <Pressable
              key={tab.href}
              onPress={() => navigate(tab, index)}
              style={styles.tab}
              accessibilityRole="tab"
              accessibilityLabel={tab.label}
              accessibilityState={{ selected: active }}
            >
              <tab.Icon size={22} color={active ? activeColor : inactiveColor} />
              <Text style={[styles.tabLabel, { color: active ? activeColor : inactiveColor }, active && styles.tabLabelActive]}>
                {tab.label}
              </Text>
            </Pressable>
          );
        })}
      </View>
    </SafeAreaView>
  );
}

const TAB_BAR_HEIGHT = 72;
const CENTER_BUTTON_SIZE = 56;

const styles = StyleSheet.create({
  safeArea: {
    borderTopWidth: StyleSheet.hairlineWidth,
  },
  bar: {
    height: TAB_BAR_HEIGHT,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-around',
    paddingHorizontal: Spacing.one,
  },
  tab: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: Spacing.one,
    gap: 1,
  },
  tabLabel: {
    fontSize: 10,
    lineHeight: 13,
    fontWeight: '500',
  },
  tabLabelActive: {
    fontWeight: '600',
  },
  centerButton: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    height: TAB_BAR_HEIGHT,
  },
  centerCircle: {
    width: CENTER_BUTTON_SIZE,
    height: CENTER_BUTTON_SIZE,
    borderRadius: CENTER_BUTTON_SIZE / 2,
    alignItems: 'center',
    justifyContent: 'center',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    transform: [{ translateY: -10 }],
    ...Platform.select({ android: { elevation: 8 } }),
  },
});
