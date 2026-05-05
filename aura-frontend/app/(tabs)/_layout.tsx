import React from 'react';
import { Tabs } from 'expo-router';
import { Platform } from 'react-native';
import { colors } from '@/theme/colors';
import { typography } from '@/theme/typography';

export default function TabsLayout() {
  // On wide web the sidebar handles navigation; tabs are only rendered on mobile
  const showTabBar = Platform.OS !== 'web';

  return (
    <Tabs
      screenOptions={{
        headerShown:      false,
        tabBarActiveTintColor:   colors.brand[500],
        tabBarInactiveTintColor: colors.text.tertiary,
        tabBarStyle: showTabBar
          ? { backgroundColor: colors.surface, borderTopColor: colors.border.light }
          : { display: 'none' },
        tabBarLabelStyle: { ...typography.bodySm, marginBottom: 2 },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{ title: 'New Incident', tabBarLabel: 'New' }}
      />
      <Tabs.Screen
        name="history"
        options={{ title: 'History', tabBarLabel: 'History' }}
      />
    </Tabs>
  );
}
