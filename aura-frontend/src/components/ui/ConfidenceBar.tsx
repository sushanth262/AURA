import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { getConfidenceColor, getConfidenceTier } from '@/utils/confidence';
import { colors } from '@/theme/colors';
import { radius } from '@/theme/spacing';

interface Props {
  score:    number | null;
  showTier?: boolean;
  height?:  number;
}

export function ConfidenceBar({ score, showTier = true, height = 8 }: Props) {
  const pct   = score === null ? 0 : Math.min(1, Math.max(0, score)) * 100;
  const color = getConfidenceColor(score);
  const tier  = getConfidenceTier(score);

  return (
    <View style={styles.container}>
      <View style={[styles.track, { height }]}>
        <View style={[styles.fill, { width: `${pct}%`, backgroundColor: color, height }]} />
      </View>
      <View style={styles.labels}>
        <Text style={[styles.score, { color }]}>
          {score !== null ? score.toFixed(2) : '—'}
        </Text>
        {showTier && (
          <Text style={[styles.tier, { color }]}>{tier}</Text>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: 6 },
  track:     { backgroundColor: colors.neutral[200], borderRadius: radius.full, overflow: 'hidden' },
  fill:      { borderRadius: radius.full },
  labels:    { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  score:     { fontSize: 13, fontWeight: '700' },
  tier:      { fontSize: 11, fontWeight: '600', letterSpacing: 0.5 },
});
