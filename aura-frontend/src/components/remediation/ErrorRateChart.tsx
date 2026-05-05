// Error rate chart — bar sparkline (RN primitives only; works for web export + native)
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

export interface DataPoint {
  x: number; // epoch ms
  y: number; // error rate 0-1
}

interface Props {
  data: DataPoint[];
  remediationTs?: number | null; // epoch ms — colors bars before/after
  label?: string;
}

export function ErrorRateChart({ data, remediationTs, label = 'Error Rate' }: Props) {
  return (
    <Card>
      <Text style={styles.heading}>{label}</Text>

      <SparklineFallback data={data} remediationTs={remediationTs} />

      {remediationTs && (
        <View style={styles.legend}>
          <View style={styles.legendItem}>
            <View style={[styles.legendDot, { backgroundColor: colors.tints.danger.text }]} />
            <Text style={styles.legendLabel}>Error rate</Text>
          </View>
          <View style={styles.legendItem}>
            <View style={[styles.legendDash, { borderColor: colors.brand[500] }]} />
            <Text style={styles.legendLabel}>Remediation applied</Text>
          </View>
        </View>
      )}
    </Card>
  );
}

function SparklineFallback({ data, remediationTs }: Pick<Props, 'data' | 'remediationTs'>) {
  if (data.length === 0) {
    return <Text style={styles.empty}>No data</Text>;
  }
  const max = Math.max(...data.map((d) => d.y), 0.001);
  return (
    <View style={styles.fallback}>
      {data.map((pt, i) => {
        const heightPct = (pt.y / max) * 100;
        const isPast = remediationTs ? pt.x < remediationTs : true;
        return (
          <View key={i} style={styles.bar}>
            <View
              style={[
                styles.barFill,
                {
                  height: `${heightPct}%` as any,
                  backgroundColor: isPast ? colors.tints.danger.text : colors.tints.success.text,
                },
              ]}
            />
          </View>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  heading: { ...typography.h3, color: colors.text.primary, marginBottom: spacing[2] },
  empty: { ...typography.bodySm, color: colors.text.tertiary, textAlign: 'center', paddingVertical: spacing[4] },
  legend: { flexDirection: 'row', gap: 16, marginTop: spacing[2] },
  legendItem: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  legendDot: { width: 8, height: 8, borderRadius: 4 },
  legendDash: { width: 14, height: 0, borderTopWidth: 2, borderStyle: 'dashed' },
  legendLabel: { ...typography.bodySm, color: colors.text.secondary },
  fallback: { height: 80, flexDirection: 'row', alignItems: 'flex-end', gap: 2, paddingVertical: 4 },
  bar: { flex: 1, height: '100%', justifyContent: 'flex-end' },
  barFill: { width: '100%', borderRadius: 2 },
});
