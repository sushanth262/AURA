// Screen 6 — Incident History
import React, { useEffect, useMemo, useState } from 'react';
import { RefreshControl, StyleSheet, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { useQuery } from '@tanstack/react-query';
import { ScreenContainer } from '@/components/layout/ScreenContainer';
import { StatsSummaryBar } from '@/components/history/StatsSummaryBar';
import { FilterBar } from '@/components/history/FilterBar';
import { IncidentTable } from '@/components/history/IncidentTable';
import { Spinner } from '@/components/ui/Spinner';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import { listIncidentHistory } from '@/api/incidents';
import type { HistoryFilters, IncidentHistoryPage } from '@/types/api';
import { useAuthStore } from '@/store/authStore';

const EMPTY_STATS: IncidentHistoryPage['stats'] = {
  total_count: 0,
  avg_time_to_diagnose_seconds: null,
  avg_confidence_score: null,
  resolved_pct: null,
};

function emptyHistoryPage(perPage: number): IncidentHistoryPage {
  return {
    items:    [],
    page:     1,
    per_page: perPage,
    total:    0,
    stats:    EMPTY_STATS,
  };
}

export default function HistoryScreen() {
  const router = useRouter();
  const token = useAuthStore((s) => s.token);
  const isReady = useAuthStore((s) => s.isReady);
  const [filters, setFilters] = useState<HistoryFilters>({ page: 1, per_page: 20 });

  useEffect(() => {
    if (isReady && !token) {
      router.replace('/login' as never);
    }
  }, [isReady, token, router]);

  const { data, isLoading, isError, isRefetching, refetch } = useQuery({
    queryKey: ['incidents', 'history', filters],
    queryFn:  () => listIncidentHistory(filters),
    placeholderData: (prev) => prev,
    enabled: isReady && !!token,
  });

  const showTableShell = Boolean(data) || isError;
  const tablePage = useMemo(
    () => (data ?? (isError ? emptyHistoryPage(filters.per_page) : null)),
    [data, isError, filters.per_page],
  );
  const stats = data?.stats ?? (isError ? EMPTY_STATS : undefined);

  return (
    <ScreenContainer
      refreshControl={(
        <RefreshControl
          refreshing={isRefetching}
          onRefresh={() => { void refetch(); }}
          tintColor={colors.brand[500]}
        />
      )}
    >
      <View style={styles.header}>
        <Text style={styles.title}>Incident History</Text>
      </View>

      {(stats != null) && <StatsSummaryBar stats={stats} />}

      <FilterBar
        filters={filters}
        onChange={setFilters}
      />

      {isError && (
        <View style={styles.errorBanner}>
          <Text style={styles.errorText}>
            Could not reach the API. Showing empty data — fix EXPO_PUBLIC_API_BASE_URL or sign in, then pull to
            refresh.
          </Text>
        </View>
      )}

      {isLoading && !data && !isError && (
        <View style={styles.center}>
          <Spinner size="large" />
        </View>
      )}

      {showTableShell && tablePage && (
        <IncidentTable
          page={tablePage}
          onPage={(p) => setFilters((f) => ({ ...f, page: p }))}
        />
      )}
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  header:      { gap: spacing[1] },
  title:       { ...typography.h1, color: colors.text.primary },
  center:      { flex: 1, alignItems: 'center', justifyContent: 'center', paddingVertical: spacing[8] },
  errorBanner: {
    backgroundColor: colors.tints.danger.bg, padding: spacing[3],
    borderRadius: 8, borderLeftWidth: 3, borderLeftColor: colors.tints.danger.text,
  },
  errorText: { ...typography.body, color: colors.tints.danger.text },
});
