import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { colors } from '@/theme/colors';
import type { Severity } from '@/types/api';

interface Props { severity: Severity; size?: 'sm' | 'md' }

export function SeverityBadge({ severity, size = 'md' }: Props) {
  const palette = colors.severity[severity];
  return (
    <View style={[styles.base, { backgroundColor: palette.bg }, size === 'sm' && styles.sm]}>
      <View style={[styles.dot, { backgroundColor: palette.dot }]} />
      <Text style={[styles.label, { color: palette.text }, size === 'sm' && styles.labelSm]}>
        {severity}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  base:    { flexDirection: 'row', alignItems: 'center', gap: 5, paddingHorizontal: 8, paddingVertical: 4, borderRadius: 6 },
  sm:      { paddingHorizontal: 6, paddingVertical: 2 },
  dot:     { width: 6, height: 6, borderRadius: 3 },
  label:   { fontSize: 12, fontWeight: '600' },
  labelSm: { fontSize: 10 },
});
