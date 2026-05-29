// Findings timeline — chronological list of agent findings with confidence bars
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { ConfidenceBar } from '@/components/ui/ConfidenceBar';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { Finding } from '@/types/api';
import { domainColor, domainLabel } from '@/utils/graphLanes';

interface Props {
  findings: Finding[];
}

export function TimelinePanel({ findings }: Props) {
  const sorted = [...findings].sort((a, b) => {
    if (a.timeline_ts && b.timeline_ts) return a.timeline_ts.localeCompare(b.timeline_ts);
    return 0;
  });

  return (
    <Card>
      <Text style={styles.heading}>Findings Timeline</Text>
      {sorted.length === 0
        ? <Text style={styles.empty}>No findings yet.</Text>
        : sorted.map((f) => <FindingItem key={f.finding_id} finding={f} />)
      }
    </Card>
  );
}

function FindingItem({ finding }: { finding: Finding }) {
  const color = domainColor(finding.domain);
  const label = domainLabel(finding.domain);
  const ts = finding.timeline_ts
    ? new Date(finding.timeline_ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    : null;

  return (
    <View style={styles.item}>
      <View style={[styles.bar, { backgroundColor: color }]} />
      <View style={styles.body}>
        <View style={styles.meta}>
          <Text style={[styles.domain, { color }]}>{label}</Text>
          <Text style={styles.findingType}>{finding.type.replace(/_/g, ' ')}</Text>
          {ts && <Text style={styles.ts}>{ts}</Text>}
        </View>
        <Text style={styles.description}>{finding.description}</Text>
        <View style={styles.confRow}>
          <Text style={styles.confLabel}>Confidence</Text>
          <View style={{ flex: 1 }}>
            <ConfidenceBar score={finding.confidence} showLabel={false} />
          </View>
        </View>
        {finding.supporting_evidence.length > 0 && (
          <Text style={styles.evidence}>
            {finding.supporting_evidence.length} evidence ref{finding.supporting_evidence.length !== 1 ? 's' : ''}
          </Text>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  heading:     { ...typography.h3, color: colors.text.primary, marginBottom: spacing[3] },
  empty:       { ...typography.bodySm, color: colors.text.tertiary, textAlign: 'center', paddingVertical: spacing[4] },
  item:        { flexDirection: 'row', marginBottom: spacing[3], gap: 10 },
  bar:         { width: 3, borderRadius: radius.full, minHeight: 40 },
  body:        { flex: 1, gap: spacing[1] },
  meta:        { flexDirection: 'row', alignItems: 'center', gap: 8, flexWrap: 'wrap' },
  domain:      { ...typography.label, fontWeight: '700' },
  findingType: { ...typography.label, color: colors.text.secondary, textTransform: 'capitalize' },
  ts:          { ...typography.bodySm, color: colors.text.tertiary },
  description: { ...typography.body, color: colors.text.primary },
  confRow:     { flexDirection: 'row', alignItems: 'center', gap: 8 },
  confLabel:   { ...typography.bodySm, color: colors.text.secondary, width: 72 },
  evidence:    { ...typography.bodySm, color: colors.brand[500] },
});
