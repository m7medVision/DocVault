import { Tabs } from 'expo-router';
import TabBar from '@/components/tab-bar';

export default function TabLayout() {
  return (
    <Tabs
      screenOptions={{
        headerShown: false,
      }}
      tabBar={(props) => <TabBar {...props} />}
    >
      <Tabs.Screen name="index" />
      <Tabs.Screen name="documents" />
      <Tabs.Screen name="scan" />
      <Tabs.Screen name="search" />
      <Tabs.Screen name="reminders" />
    </Tabs>
  );
}