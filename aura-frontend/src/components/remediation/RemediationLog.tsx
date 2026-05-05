// Scrollable log of remediation steps with status indicators
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { RemediationLogEntry } from '@/types/api';

interface Props {
  entries: RemediationLogEntry[];
}

export function RemediationLog({ entries }: Props) {
  return (
    <Card>
      <Text style={styles.heading}>Remediation Log</Text>
      {entries.length === 0
        ? <Text style={styles.empty}>No log entries yet.</Text>
        : entries.map((entry, i) => <LogRow key={i} entry={entry} isLast={i === entries.length - 1} />)
      }
    </Card>
  );
}

function LogRow({ entry, isLast }: { entry: RemediationLogEntry; isLast: boolean }) {
  const ok = entry.status === 'complete';
  return (
    <View style={styles.row}>
      {/* Timeline connector */}
      <View style={styles.timeline}>
        <View style={[styles.dot, { backgroundColor: ok ? colors.tints.success.text : colors.tints.danger.text }]} />
        {!isLast && <View style={styles.connector} />}
      </View>
      <View style={[styles.body, !isLast && styles.bodySpaced]}>
        <Text style={styles.description}>{entry.description}</Text>
        <View style={styles.meta}>
          <View style={[styles.statusBadge, { backgroundColor: ok ? colors.tints.success.bg : colors.tints.danger.bg }]}>
            <Text style={[styles.statusText, { color: ok ? colors.tints.success.text : colors.tints.danger.text }]}>
              {ok ? '✓ Complete' : '✕ Failed'}
            </Text>
          </View>
          <Text style={styles.ts}>
            {new Date(entry.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
          </Text>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  heading:     { ...typography.h3, color: colors.text.primary, marginBottom: spacing[3] },
  empty:       { ...typography.bodySm, color: colors.text.tertiary, textAlign: 'center', paddingVertical: spacing[4] },
  row:         { flexDirection: 'row', gap: 12 },
  timeline:    { alignItems: 'center', width: 16 },
  dot:         { width: 12, height: 12, borderRadius: radius.full, marginTop: 3 },
  connector:   { flex: 1, width: 2, backgroundColor: colors.border.light, marginTop: 4 },
  body:        { flex: 1, paddingBottom: 0 },
  bodySpaced:  { paddingBottom: spacing[4] },
  description: { ...typography.body, color: colors.text.primary },
  meta:        { flexDirection: 'row', alignItems: 'center', gap: 10, marginTop: spacing[1] },
  statusBadge: { paddingHorizontal: 8, paddingVertical: 2, borderRadius: radius.full },
  statusText:  { ...typography.bodySm, fontWeight: '600' },
  ts:          { ...typography.bodySm, color: colors.text.tertiary },
});
