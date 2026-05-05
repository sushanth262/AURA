// Overall confidence score + sub-score breakdown panel (Screen 3)
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { ConfidenceBar } from '@/components/ui/ConfidenceBar';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import { formatConfidence } from '@/utils/confidence';
import type { ConfidenceBreakdown } from '@/types/api';

interface Props {
  score:     number | null;
  breakdown: ConfidenceBreakdown;
}

const SUB_SCORES: Array<{ key: keyof ConfidenceBreakdown; label: string }> = [
  { key: 'citation_strength',  label: 'Citation Strength' },
  { key: 'agent_agreement',    label: 'Agent Agreement' },
  { key: 'memory_match_boost', label: 'Memory Match Boost' },
  { key: 'rejection_penalty',  label: 'Rejection Penalty' },
];

export function ConfidencePanel({ score, breakdown }: Props) {
  return (
    <Card tint={tintFor(score)}>
      <View style={styles.header}>
        <Text style={styles.heading}>Overall Confidence</Text>
        <Text style={styles.score}>{formatConfidence(score)}</Text>
      </View>

      <ConfidenceBar score={score} />

      <View style={styles.breakdown}>
        {SUB_SCORES.map(({ key, label }) => (
          <View key={key} style={styles.subRow}>
            <Text style={styles.subLabel}>{label}</Text>
            <View style={styles.subBar}>
              <ConfidenceBar score={breakdown[key]} showLabel={false} />
            </View>
            <Text style={styles.subValue}>
              {key === 'rejection_penalty'
                ? `−${Math.round(Math.abs(breakdown[key]) * 100)}%`
                : `${Math.round(breakdown[key] * 100)}%`}
            </Text>
          </View>
        ))}
      </View>
    </Card>
  );
}

function tintFor(score: number | null): 'success' | 'warning' | 'danger' | 'none' {
  if (score === null)  return 'none';
  if (score >= 0.75)   return 'success';
  if (score >= 0.5)    return 'warning';
  return 'danger';
}

const styles = StyleSheet.create({
  header:   { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: spacing[2] },
  heading:  { ...typography.h3, color: colors.text.primary },
  score:    { ...typography.metricValue, color: colors.text.primary },
  breakdown:{ marginTop: spacing[3], gap: spacing[2] },
  subRow:   { flexDirection: 'row', alignItems: 'center', gap: 10 },
  subLabel: { ...typography.bodySm, color: colors.text.secondary, width: 140 },
  subBar:   { flex: 1 },
  subValue: { ...typography.label, color: colors.text.primary, width: 40, textAlign: 'right' },
});
