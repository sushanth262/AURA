// TrueStat-style KPI row: 4 metric cards across the top of Screen 6
import React from 'react';
import { StyleSheet, View } from 'react-native';
import { MetricCard } from '@/components/ui/MetricCard';
import { formatDuration, formatPct } from '@/utils/formatting';
import { formatConfidence } from '@/utils/confidence';
import type { HistoryStats } from '@/types/api';

interface Props { stats: HistoryStats }

export function StatsSummaryBar({ stats }: Props) {
  return (
    <View style={styles.row}>
      <MetricCard
        label="Total (30d)"
        value={String(stats.total_count)}
        tint="none"
      />
      <MetricCard
        label="Avg Time-to-Diagnose"
        value={stats.avg_time_to_diagnose_seconds != null
          ? formatDuration(Math.round(stats.avg_time_to_diagnose_seconds))
          : '—'}
        tint="none"
      />
      <MetricCard
        label="Avg Confidence"
        value={formatConfidence(stats.avg_confidence_score ?? null)}
        tint={
          (stats.avg_confidence_score ?? 0) >= 0.75 ? 'success' :
          (stats.avg_confidence_score ?? 0) >= 0.5  ? 'warning' : 'danger'
        }
      />
      <MetricCard
        label="Resolved"
        value={formatPct(stats.resolved_pct)}
        tint={(stats.resolved_pct ?? 0) >= 80 ? 'success' : 'warning'}
        trend={(stats.resolved_pct ?? 0) >= 80 ? 'up' : 'down'}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', gap: 12, flexWrap: 'wrap' },
});
