// Synthesis narrative card — renders structured JSON report with themed UI
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { NarrativeReport } from '@/types/api';

interface Props {
  narrative:  NarrativeReport;
  iteration?: number;
}

const SEVERITY_TINT: Record<string, { bg: string; text: string }> = {
  P1: colors.severity.P1,
  P2: colors.severity.P2,
  P3: colors.severity.P3,
  P4: colors.severity.P4,
};

const AGENT_ICON: Record<string, string> = {
  'Telemetry Agent': '📊',
  'Code Agent':      '🔧',
  'Context Agent':   '📋',
};

export function NarrativePanel({ narrative, iteration }: Props) {
  const meta = narrative.report_metadata;
  const sev = SEVERITY_TINT[meta.severity] ?? { bg: colors.neutral[100], text: colors.text.secondary };

  return (
    <Card>
      {/* ── Header: RCA Topic + Iteration ── */}
      <View style={styles.topRow}>
        <Text style={styles.heading} numberOfLines={2}>{meta.rca_topic}</Text>
        {iteration != null && iteration > 1 && (
          <View style={styles.iterBadge}>
            <Text style={styles.iterLabel}>Iteration {iteration}</Text>
          </View>
        )}
      </View>

      {/* ── Metadata pills ── */}
      <View style={styles.pillRow}>
        <View style={[styles.pill, { backgroundColor: sev.bg }]}>
          <Text style={[styles.pillText, { color: sev.text }]}>{meta.severity}</Text>
        </View>
        <View style={[styles.pill, { backgroundColor: colors.brand[50] }]}>
          <Text style={[styles.pillText, { color: colors.brand[500] }]}>{meta.service}</Text>
        </View>
        <View style={[styles.pill, { backgroundColor: colors.neutral[100] }]}>
          <Text style={[styles.pillText, { color: colors.neutral[600] }]}>{meta.status}</Text>
        </View>
      </View>

      {/* ── Symptoms ── */}
      {narrative.symptoms ? (
        <View style={styles.section}>
          <Text style={styles.sectionLabel}>REPORTED SYMPTOMS</Text>
          <View style={styles.symptomsBox}>
            <Text style={styles.symptomsText}>{narrative.symptoms}</Text>
          </View>
        </View>
      ) : null}

      {/* ── Agent Findings ── */}
      <View style={styles.section}>
        <Text style={styles.sectionLabel}>AGENT FINDINGS</Text>
        {narrative.agent_findings.map((agent, i) => (
          <View key={i} style={styles.agentCard}>
            <View style={styles.agentHeader}>
              <Text style={styles.agentIcon}>{AGENT_ICON[agent.agent_name] ?? '🔍'}</Text>
              <View style={{ flex: 1 }}>
                <Text style={styles.agentName}>{agent.agent_name}</Text>
                <Text style={styles.agentFocus}>{agent.focus}</Text>
              </View>
            </View>
            <Text style={styles.agentObservation}>{agent.observation}</Text>
          </View>
        ))}
      </View>

      {/* ── Conclusion ── */}
      <View style={styles.conclusionCard}>
        <View style={styles.conclusionHeader}>
          <Text style={styles.sectionLabel}>CONCLUSION</Text>
          <View style={[styles.pill, { backgroundColor: colors.info[50] }]}>
            <Text style={[styles.pillText, { color: colors.info[600] }]}>
              Confidence: {narrative.conclusion.confidence_level}
            </Text>
          </View>
        </View>
        <Text style={styles.conclusionSummary}>{narrative.conclusion.summary}</Text>
        <View style={styles.actionBox}>
          <Text style={styles.actionLabel}>RECOMMENDED ACTION</Text>
          <Text style={styles.actionText}>{narrative.conclusion.action_item}</Text>
        </View>
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  topRow:           { flexDirection: 'row', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: spacing[3], gap: spacing[2] },
  heading:          { ...typography.h2, color: colors.text.primary, flex: 1 },
  iterBadge:        { backgroundColor: colors.brand[50], paddingHorizontal: 8, paddingVertical: 3, borderRadius: 99 },
  iterLabel:        { ...typography.label, color: colors.brand[500] },

  pillRow:          { flexDirection: 'row', flexWrap: 'wrap', gap: 8, marginBottom: spacing[4] },
  pill:             { paddingHorizontal: 10, paddingVertical: 4, borderRadius: 99 },
  pillText:         { ...typography.label, fontSize: 11 },

  section:          { marginBottom: spacing[4] },
  sectionLabel:     { ...typography.label, color: colors.text.secondary, marginBottom: spacing[2] },

  symptomsBox:      { backgroundColor: colors.neutral[50], borderRadius: radius.md, padding: spacing[3], borderLeftWidth: 3, borderLeftColor: colors.warning[500] },
  symptomsText:     { ...typography.body, color: colors.text.primary, lineHeight: 20 },

  agentCard:        { backgroundColor: colors.neutral[50], borderRadius: radius.md, padding: spacing[3], marginBottom: spacing[2] },
  agentHeader:      { flexDirection: 'row', alignItems: 'center', gap: spacing[2], marginBottom: spacing[2] },
  agentIcon:        { fontSize: 20 },
  agentName:        { ...typography.h3, color: colors.text.primary },
  agentFocus:       { ...typography.bodySm, color: colors.text.secondary },
  agentObservation: { ...typography.body, color: colors.text.primary, lineHeight: 20 },

  conclusionCard:   { backgroundColor: colors.success[50], borderRadius: radius.md, padding: spacing[3], borderLeftWidth: 3, borderLeftColor: colors.success[500] },
  conclusionHeader: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginBottom: spacing[2] },
  conclusionSummary:{ ...typography.body, color: colors.text.primary, lineHeight: 20, marginBottom: spacing[3] },

  actionBox:        { backgroundColor: colors.surface, borderRadius: radius.sm, padding: spacing[3] },
  actionLabel:      { ...typography.label, color: colors.brand[500], marginBottom: spacing[1] },
  actionText:       { ...typography.body, color: colors.text.primary, lineHeight: 20 },
});
