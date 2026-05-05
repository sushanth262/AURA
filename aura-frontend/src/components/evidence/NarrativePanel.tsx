// Synthesis narrative card — full text from the Supervisor's synthesis step
import React, { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

interface Props {
  narrative:  string;
  iteration?: number;
}

const PREVIEW_LINES = 5;

export function NarrativePanel({ narrative, iteration }: Props) {
  const lines = narrative.split('\n');
  const long  = lines.length > PREVIEW_LINES;
  const [expanded, setExpanded] = useState(false);

  const visible = long && !expanded
    ? lines.slice(0, PREVIEW_LINES).join('\n')
    : narrative;

  return (
    <Card>
      <View style={styles.header}>
        <Text style={styles.heading}>Synthesis Narrative</Text>
        {iteration != null && iteration > 1 && (
          <View style={styles.iterBadge}>
            <Text style={styles.iterText}>Iteration {iteration}</Text>
          </View>
        )}
      </View>

      <Text style={styles.body}>{visible}</Text>

      {long && (
        <Pressable onPress={() => setExpanded(!expanded)} style={styles.toggleBtn}>
          <Text style={styles.toggleText}>{expanded ? 'Show less' : 'Show full narrative'}</Text>
        </Pressable>
      )}
    </Card>
  );
}

const styles = StyleSheet.create({
  header:     { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: spacing[2] },
  heading:    { ...typography.h3, color: colors.text.primary },
  iterBadge:  { backgroundColor: colors.brand[50], paddingHorizontal: 8, paddingVertical: 3, borderRadius: 99 },
  iterText:   { ...typography.label, color: colors.brand[500] },
  body:       { ...typography.body, color: colors.text.primary, lineHeight: 22 },
  toggleBtn:  { marginTop: spacing[2], alignSelf: 'flex-start' },
  toggleText: { ...typography.bodySm, color: colors.brand[500], fontWeight: '600' },
});
