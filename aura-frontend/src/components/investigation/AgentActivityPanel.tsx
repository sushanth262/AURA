// Live agent activity feed — streams AGENT_STARTED / AGENT_COMPLETE events
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { AgentDomain, TaskProgressEvent, WSEventType } from '@/types/api';

interface Props {
  events: TaskProgressEvent[];
}

const DOMAIN_LABEL: Record<AgentDomain, string> = {
  telemetry: 'Telemetry / RCA',
  code:      'Code / Fix',
  context:   'Context / Docs',
  supervisor:'Supervisor',
};

const DOMAIN_COLOR: Record<AgentDomain, string> = {
  telemetry: '#3B82F6',  // blue
  code:      '#8B5CF6',  // violet
  context:   '#10B981',  // emerald
  supervisor:colors.brand[500],
};

const EVENT_LABEL: Partial<Record<WSEventType, string>> = {
  AGENT_STARTED:       'started',
  AGENT_COMPLETE:      'complete',
  SYNTHESIS_STARTED:   'synthesizing',
  SYNTHESIS_COMPLETE:  'synthesis done',
  HITL_REQUESTED:      'awaiting review',
  REMEDIATION_STARTED: 'remediating',
  TASK_FAILED:         'failed',
};

function dot(color: string) {
  return <View style={[styles.dot, { backgroundColor: color }]} />;
}

export function AgentActivityPanel({ events }: Props) {
  const visible = events.filter((e) =>
    ['AGENT_STARTED', 'AGENT_COMPLETE', 'SYNTHESIS_STARTED', 'SYNTHESIS_COMPLETE',
     'HITL_REQUESTED', 'REMEDIATION_STARTED', 'TASK_FAILED'].includes(e.event_type),
  );

  return (
    <Card>
      <Text style={styles.heading}>Agent Activity</Text>
      {visible.length === 0 ? (
        <Text style={styles.empty}>Waiting for agents…</Text>
      ) : (
        visible.slice().reverse().map((e, i) => {
          const domain = e.agent_domain ?? 'supervisor';
          const color  = DOMAIN_COLOR[domain] ?? colors.brand[500];
          const label  = DOMAIN_LABEL[domain] ?? domain;
          const verb   = EVENT_LABEL[e.event_type] ?? e.event_type.toLowerCase();
          const ts     = new Date(e.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
          return (
            <View key={i} style={styles.row}>
              {dot(color)}
              <View style={styles.rowBody}>
                <Text style={styles.rowTitle}>
                  <Text style={[styles.domain, { color }]}>{label}</Text>
                  {' '}{verb}
                </Text>
                <Text style={styles.ts}>{ts}</Text>
              </View>
            </View>
          );
        })
      )}
    </Card>
  );
}

const styles = StyleSheet.create({
  heading: { ...typography.h3, color: colors.text.primary, marginBottom: spacing[3] },
  empty:   { ...typography.bodySm, color: colors.text.tertiary, textAlign: 'center', paddingVertical: spacing[4] },
  row:     { flexDirection: 'row', alignItems: 'flex-start', gap: 10, marginBottom: spacing[3] },
  dot:     { width: 8, height: 8, borderRadius: radius.full, marginTop: 5 },
  rowBody: { flex: 1 },
  rowTitle:{ ...typography.body, color: colors.text.primary },
  domain:  { fontWeight: '600' },
  ts:      { ...typography.bodySm, color: colors.text.tertiary, marginTop: 2 },
});
