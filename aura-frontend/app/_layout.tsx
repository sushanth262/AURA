import React, { useEffect } from 'react';
import { Platform, View } from 'react-native';
import { Stack } from 'expo-router';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { AppShell } from '@/components/layout/AppShell';
import { useAuthStore } from '@/store/authStore';
import { colors } from '@/theme/colors';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 2, staleTime: 30_000 },
  },
});

export default function RootLayout() {
  const hydrate = useAuthStore((s) => s.hydrate);

  useEffect(() => {
    hydrate();
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <SafeAreaProvider style={Platform.OS === 'web' ? { flex: 1, minHeight: '100vh' as const } : undefined}>
        <AppShell>
          <View style={{ flex: 1, alignSelf: 'stretch', minWidth: 0, minHeight: 0, width: '100%' }}>
            <Stack
              screenOptions={{
                headerShown: false,
                contentStyle: Platform.OS === 'web'
                  ? { backgroundColor: colors.canvas, flex: 1, minHeight: '100%' as const, width: '100%' }
                  : { backgroundColor: colors.canvas, flex: 1 },
                animation: 'slide_from_right',
              }}
            />
          </View>
        </AppShell>
      </SafeAreaProvider>
    </QueryClientProvider>
  );
}
