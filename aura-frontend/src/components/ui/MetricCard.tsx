// TrueStat-inspired metric card: colored tint, large bold value, delta indicator.
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Card } from './Card';
import { colors } from '@/theme/colors';
import { typography } from '@/theme/typography';

interface Props {
  label:  string;
  value:  string;
  delta?: string | null;         // e.g. "+12.86%"
  trend?: 'up' | 'down' | null;
  tint?:  'none' | 'success' | 'warning' | 'danger';
}

export function MetricCard({ label, value, delta, trend, tint = 'none' }: Props) {
  const trendColor =
    trend === 'up'   ? colors.success[500] :
    trend === 'down' ? colors.danger[500]  :
    colors.text.secondary;

  return (
    <Card tint={tint} style={styles.card}>
      <Text style={styles.label}>{label}</Text>
      <Text style={styles.value} numberOfLines={1} adjustsFontSizeToFit>
        {value}
      </Text>
      {delta != null && (
        <View style={styles.deltaRow}>
          <Text style={[styles.delta, { color: trendColor }]}>
            {trend === 'up' ? '↑ ' : trend === 'down' ? '↓ ' : ''}{delta}
          </Text>
        </View>
      )}
    </Card>
  );
}

const styles = StyleSheet.create({
  card:     { flex: 1, minWidth: 140 },
  label:    { ...typography.label, color: colors.text.secondary, marginBottom: 6 },
  value:    { ...typography.metricValue, color: colors.text.primary, marginBottom: 4 },
  deltaRow: { flexDirection: 'row', alignItems: 'center' },
  delta:    { ...typography.delta },
});
