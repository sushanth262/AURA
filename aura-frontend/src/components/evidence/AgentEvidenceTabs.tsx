// Tabbed per-agent summaries and findings list (Telemetry / Code / Context tabs)
import React, { useState } from 'react';
import { Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { AgentDomain, AgentSummary, Finding } from '@/types/api';

interface Props {
  summaries: AgentSummary[];
  findings:  Finding[];
}

const DOMAIN_ORDER: AgentDomain[] = ['telemetry', 'code', 'context'];
const DOMAIN_LABEL: Record<AgentDomain, string> = {
  telemetry: 'Telemetry',
  code:      'Code',
  context:   'Context',
  supervisor:'Supervisor',
};
const DOMAIN_COLOR: Record<AgentDomain, string> = {
  telemetry: '#3B82F6',
  code:      '#8B5CF6',
  context:   '#10B981',
  supervisor:colors.brand[500],
};
const STATUS_TINT: Record<string, string> = {
  SUCCESS: colors.tints.success.bg,
  PARTIAL: colors.tints.warning.bg,
  FAILED:  colors.tints.danger.bg,
  SKIPPED: colors.neutral[100],
};

export function AgentEvidenceTabs({ summaries, findings }: Props) {
  const [activeTab, setActiveTab] = useState<AgentDomain>('telemetry');

  const summary  = summaries.find((s) => s.domain === activeTab);
  const tabFinds = findings.filter((f) => f.domain === activeTab);

  return (
    <Card padding={0}>
      {/* Tab bar */}
      <View style={styles.tabBar}>
        {DOMAIN_ORDER.map((d) => {
          const active = d === activeTab;
          const sum    = summaries.find((s) => s.domain === d);
          return (
            <Pressable
              key={d}
              onPress={() => setActiveTab(d)}
              style={[styles.tab, active && { borderBottomColor: DOMAIN_COLOR[d] }]}
            >
              <Text style={[styles.tabText, active && { color: DOMAIN_COLOR[d] }]}>
                {DOMAIN_LABEL[d]}
              </Text>
              {sum && (
                <View style={[styles.statusDot, { backgroundColor: STATUS_TINT[sum.status] }]} />
              )}
            </Pressable>
          );
        })}
      </View>

      {/* Tab body */}
      <ScrollView style={styles.body}>
        {summary ? (
          <View style={styles.summaryBox}>
            <Text style={styles.summaryText}>{summary.summary}</Text>
            <View style={styles.summaryMeta}>
              <Text style={styles.metaItem}>{summary.finding_count} findings</Text>
              {summary.execution_duration_ms != null && (
                <Text style={styles.metaItem}>{(summary.execution_duration_ms / 1000).toFixed(1)}s</Text>
              )}
              <View style={[styles.statusChip, { backgroundColor: STATUS_TINT[summary.status] }]}>
                <Text style={styles.statusText}>{summary.status}</Text>
              </View>
            </View>
          </View>
        ) : (
          <Text style={styles.empty}>No summary available.</Text>
        )}

        {tabFinds.length > 0 && (
          <View style={styles.findingsList}>
            <Text style={styles.findingsHeading}>Findings ({tabFinds.length})</Text>
            {tabFinds.map((f) => (
              <View key={f.finding_id} style={styles.findingRow}>
                <Text style={styles.findingType}>{f.type.replace(/_/g, ' ')}</Text>
                <Text style={styles.findingDesc}>{f.description}</Text>
                <Text style={styles.findingConf}>Confidence: {Math.round(f.confidence * 100)}%</Text>
              </View>
            ))}
          </View>
        )}
      </ScrollView>
    </Card>
  );
}

const styles = StyleSheet.create({
  tabBar: {
    flexDirection: 'row',
    borderBottomWidth: 1,
    borderBottomColor: colors.border.light,
  },
  tab: {
    flex: 1, alignItems: 'center', paddingVertical: spacing[3],
    borderBottomWidth: 2, borderBottomColor: 'transparent',
    flexDirection: 'row', justifyContent: 'center', gap: 6,
  },
  tabText:         { ...typography.label, color: colors.text.secondary },
  statusDot:       { width: 7, height: 7, borderRadius: radius.full },
  body:            { padding: spacing[4] },
  summaryBox:      { gap: spacing[2], marginBottom: spacing[3] },
  summaryText:     { ...typography.body, color: colors.text.primary, lineHeight: 22 },
  summaryMeta:     { flexDirection: 'row', alignItems: 'center', gap: 10, flexWrap: 'wrap' },
  metaItem:        { ...typography.bodySm, color: colors.text.secondary },
  statusChip:      { paddingHorizontal: 8, paddingVertical: 3, borderRadius: radius.full },
  statusText:      { ...typography.label, color: colors.text.primary },
  empty:           { ...typography.bodySm, color: colors.text.tertiary, textAlign: 'center', paddingVertical: spacing[4] },
  findingsList:    { gap: spacing[2] },
  findingsHeading: { ...typography.label, color: colors.text.secondary, marginBottom: spacing[1] },
  findingRow: {
    padding: spacing[3],
    backgroundColor: colors.neutral[50],
    borderRadius: radius.md,
    gap: 4,
  },
  findingType: { ...typography.label, color: colors.text.primary, textTransform: 'capitalize' },
  findingDesc: { ...typography.bodySm, color: colors.text.primary },
  findingConf: { ...typography.bodySm, color: colors.text.secondary },
});
