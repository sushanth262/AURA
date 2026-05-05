import React from 'react';
import { StyleSheet, View } from 'react-native';
import { colors } from '@/theme/colors';
import { radius } from '@/theme/spacing';

interface Props {
  value:  number;   // 0–100
  color?: string;
  height?: number;
}

export function ProgressBar({ value, color = colors.brand[500], height = 6 }: Props) {
  const pct = Math.min(100, Math.max(0, value));
  return (
    <View style={[styles.track, { height }]}>
      <View style={[styles.fill, { width: `${pct}%`, backgroundColor: color, height }]} />
    </View>
  );
}

const styles = StyleSheet.create({
  track: { backgroundColor: colors.neutral[200], borderRadius: radius.full, overflow: 'hidden' },
  fill:  { borderRadius: radius.full },
});
