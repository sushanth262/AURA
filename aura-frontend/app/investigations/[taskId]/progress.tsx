// Screen 2 — Live Investigation Progress
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useLocalSearchParams } from 'expo-router';
import { useQuery } from '@tanstack/react-query';
import { ScreenContainer } from '@/components/layout/ScreenContainer';
import { InvestigationGraph } from '@/components/investigation/InvestigationGraph';
import { AgentActivityPanel } from '@/components/investigation/AgentActivityPanel';
import { TimelinePanel } from '@/components/investigation/TimelinePanel';
import { IncidentHeader } from '@/components/incidents/IncidentHeader';
import { Spinner } from '@/components/ui/Spinner';
import { useInvestigationWS } from '@/hooks/useInvestigationWS';
import { useInvestigationStore } from '@/store/investigationStore';
import { getIncident } from '@/api/incidents';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { Finding } from '@/types/api';

export default function ProgressScreen() {
  const { taskId } = useLocalSearchParams<{ taskId: string }>();
  const events     = useInvestigationStore((s) => s.getEvents(taskId));

  // Derive incident_id from WS events (first event carries it)
  const incidentId = events[0]?.incident_id;

  const { data: incident, isLoading } = useQuery({
    queryKey: ['incidents', incidentId],
    queryFn:  () => getIncident(incidentId!),
    enabled:  !!incidentId,
  });

  // Open WebSocket and drive store
  useInvestigationWS(taskId);

  // Extract findings from SYNTHESIS_COMPLETE payload if present
  const synthEvent = events.find((e) => e.event_type === 'SYNTHESIS_COMPLETE');
  const findings: Finding[] = (synthEvent?.payload?.findings as Finding[]) ?? [];

  const currentStatus = incident?.status ?? 'QUEUED';

  return (
    <ScreenContainer>
      <View style={styles.header}>
        <Text style={styles.title}>Live Investigation</Text>
        <Text style={styles.taskId}>Task {taskId.slice(-8).toUpperCase()}</Text>
      </View>

      {isLoading && !incident ? (
        <View style={styles.center}><Spinner size="large" /></View>
      ) : incident ? (
        <IncidentHeader incident={incident} />
      ) : null}

      <InvestigationGraph
        currentStatus={currentStatus}
        events={events}
      />

      <AgentActivityPanel events={events} />

      {findings.length > 0 && <TimelinePanel findings={findings} />}
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  header:  { gap: spacing[1] },
  title:   { ...typography.h1, color: colors.text.primary },
  taskId:  { ...typography.bodySm, color: colors.text.tertiary },
  center:  { paddingVertical: spacing[8], alignItems: 'center' },
});
