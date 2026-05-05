import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { colors } from '@/theme/colors';
import type { InvestigationStatus } from '@/types/api';

const DISPLAY: Record<InvestigationStatus, string> = {
  QUEUED:           'Queued',
  INTAKE:           'Normalizing',
  PLANNING:         'Planning',
  RETRIEVING:       'Investigating',
  SYNTHESIS:        'Synthesizing',
  HITL_PENDING:     'Awaiting Review',
  REPLANNING:       'Replanning',
  REMEDIATION:      'Remediating',
  MEMORY_WRITEBACK: 'Saving',
  COMPLETE:         'Resolved',
  PARTIAL_EVIDENCE: 'Partial Evidence',
  FAILED:           'Failed',
};

interface Props { status: InvestigationStatus; size?: 'sm' | 'md' }

export function StatusChip({ status, size = 'md' }: Props) {
  const palette = colors.status[status] ?? { bg: '#F1F5F9', text: '#475569' };
  return (
    <View style={[styles.chip, { backgroundColor: palette.bg }, size === 'sm' && styles.sm]}>
      <Text style={[styles.text, { color: palette.text }, size === 'sm' && styles.textSm]}>
        {DISPLAY[status] ?? status}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  chip:   { paddingHorizontal: 10, paddingVertical: 4, borderRadius: 20 },
  sm:     { paddingHorizontal: 7, paddingVertical: 2 },
  text:   { fontSize: 12, fontWeight: '600' },
  textSm: { fontSize: 10 },
});
