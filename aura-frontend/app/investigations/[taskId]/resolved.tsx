// Screen 5 — Resolved / Post-Remediation Summary
import React, { useMemo } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { ScreenContainer } from '@/components/layout/ScreenContainer';
import { RemediationLog } from '@/components/remediation/RemediationLog';
import { ErrorRateChart } from '@/components/remediation/ErrorRateChart';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { BackButton } from '@/components/ui/BackButton';
import { MetricCard } from '@/components/ui/MetricCard';
import { useEvidenceBundle } from '@/hooks/useEvidenceBundle';
import { useInvestigationStore } from '@/store/investigationStore';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import { formatDuration } from '@/utils/formatting';
import { formatConfidence } from '@/utils/confidence';
import type { DataPoint } from '@/components/remediation/ErrorRateChart';
import type { RemediationLogEntry } from '@/types/api';

export default function ResolvedScreen() {
  const { taskId } = useLocalSearchParams<{ taskId: string }>();
  const router     = useRouter();
  const { bundle } = useEvidenceBundle(taskId);
  const events     = useInvestigationStore((s) => s.getEvents(taskId));

  // Build remediation log from WS events
  const logEntries = useMemo<RemediationLogEntry[]>(() => {
    const remStart = events.find((e) => e.event_type === 'REMEDIATION_STARTED');
    const remEnd   = events.find((e) => e.event_type === 'REMEDIATION_COMPLETE');
    const entries: RemediationLogEntry[] = [];
    if (remStart) {
      entries.push({ timestamp: remStart.timestamp, description: 'Remediation started', status: 'complete' });
    }
    if (remEnd) {
      entries.push({ timestamp: remEnd.timestamp, description: 'Remediation complete', status: 'complete' });
    }
    return entries;
  }, [events]);

  // Synthetic chart data from event timestamps (real app would pull metric data)
  const remediationTs = useMemo(() => {
    const e = events.find((ev) => ev.event_type === 'REMEDIATION_STARTED');
    return e ? new Date(e.timestamp).getTime() : null;
  }, [events]);

  const chartData = useMemo<DataPoint[]>(() => {
    if (!remediationTs) return [];
    const points: DataPoint[] = [];
    const before = remediationTs - 20 * 60 * 1000;
    for (let i = 0; i <= 30; i++) {
      const x = before + i * (2 * 60 * 1000);
      const inIncident = x < remediationTs;
      points.push({ x, y: inIncident ? 0.1 + Math.random() * 0.3 : Math.random() * 0.03 });
    }
    return points;
  }, [remediationTs]);

  const confidence = bundle?.confidence_score ?? null;
  const diagnosisMs = useMemo(() => {
    const start = events[0]?.timestamp;
    const end   = events.find((e) => e.event_type === 'SYNTHESIS_COMPLETE')?.timestamp;
    if (!start || !end) return null;
    return new Date(end).getTime() - new Date(start).getTime();
  }, [events]);

  return (
    <ScreenContainer>
      <BackButton label="Back" fallbackHref="/" />
      <View style={styles.header}>
        <View style={styles.resolvedBadge}>
          <Text style={styles.resolvedBadgeText}>✓ Resolved</Text>
        </View>
        <Text style={styles.title}>Incident Resolved</Text>
        <Text style={styles.subtitle}>
          Remediation actions have been applied. Review the outcome below.
        </Text>
      </View>

      {/* KPI row */}
      <View style={styles.kpis}>
        <MetricCard
          label="Diagnosis Time"
          value={diagnosisMs != null ? formatDuration(Math.round(diagnosisMs / 1000)) : '—'}
          tint="none"
        />
        <MetricCard
          label="Confidence"
          value={formatConfidence(confidence)}
          tint={confidence != null ? (confidence >= 0.75 ? 'success' : confidence >= 0.5 ? 'warning' : 'danger') : 'none'}
        />
      </View>

      <ErrorRateChart
        data={chartData}
        remediationTs={remediationTs}
        label="Error Rate Recovery"
      />

      <RemediationLog entries={logEntries} />

      {bundle?.prior_incident_matches.length ? (
        <Card>
          <Text style={styles.sectionHeading}>Similar Past Incidents</Text>
          {bundle.prior_incident_matches.map((m) => (
            <View key={m.incident_id} style={styles.priorRow}>
              <Text style={styles.priorId}>INC-{m.incident_id.slice(-4).toUpperCase()}</Text>
              {m.title && <Text style={styles.priorTitle} numberOfLines={1}>{m.title}</Text>}
              <Text style={styles.priorSim}>{Math.round(m.similarity_score * 100)}% match</Text>
            </View>
          ))}
        </Card>
      ) : null}

      <Button
        label="Back to Home"
        variant="secondary"
        onPress={() => router.replace('/')}
      />
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  header:            { gap: spacing[2] },
  resolvedBadge:     { alignSelf: 'flex-start', backgroundColor: colors.tints.success.bg, borderRadius: 99, paddingHorizontal: 12, paddingVertical: 4 },
  resolvedBadgeText: { ...typography.label, color: colors.tints.success.text },
  title:             { ...typography.h1, color: colors.text.primary },
  subtitle:          { ...typography.body, color: colors.text.secondary },
  kpis:              { flexDirection: 'row', gap: 12 },
  sectionHeading:    { ...typography.h3, color: colors.text.primary, marginBottom: spacing[2] },
  priorRow:          { flexDirection: 'row', alignItems: 'center', gap: 10, paddingVertical: spacing[2], borderBottomWidth: 1, borderBottomColor: colors.border.light },
  priorId:           { ...typography.label, color: colors.brand[500], width: 72 },
  priorTitle:        { ...typography.bodySm, color: colors.text.primary, flex: 1 },
  priorSim:          { ...typography.bodySm, color: colors.text.secondary },
});
