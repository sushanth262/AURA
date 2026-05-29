// Non-linear state graph visualization — swimlanes from GRAPH_PLANNED manifest or defaults
import React, { useMemo } from 'react';
import { Platform, ScrollView, StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { InvestigationStatus, TaskProgressEvent } from '@/types/api';
import { buildLanesFromEvents } from '@/utils/graphLanes';

interface Props {
  currentStatus: InvestigationStatus;
  events:        TaskProgressEvent[];
}

const STATUS_PHASE: Partial<Record<InvestigationStatus, string>> = {
  QUEUED:           'Queued',
  INTAKE:           'Intake',
  PLANNING:         'Planning',
  RETRIEVING:       'Retrieving',
  SYNTHESIS:        'Synthesizing',
  HITL_PENDING:     'Awaiting HITL',
  REPLANNING:       'Replanning',
  REMEDIATION:      'Remediating',
  MEMORY_WRITEBACK: 'Writing Memory',
  COMPLETE:         'Complete',
  PARTIAL_EVIDENCE: 'Partial',
  FAILED:           'Failed',
};

export function InvestigationGraph({ currentStatus, events }: Props) {
  const phase = STATUS_PHASE[currentStatus] ?? currentStatus;
  const lanes = useMemo(() => buildLanesFromEvents(events), [events]);

  return (
    <Card padding={0}>
      <View style={styles.header}>
        <Text style={styles.heading}>Investigation Graph</Text>
        <View style={[styles.phase, phaseStyle(currentStatus)]}>
          <Text style={[styles.phaseText, phaseTextStyle(currentStatus)]}>{phase}</Text>
        </View>
      </View>

      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        style={Platform.OS === 'web' ? styles.swimlaneScrollWeb : undefined}
      >
        <View style={styles.swimlanes}>
          {lanes.map((lane) => {
            const laneEvents = events.filter((e) => e.agent_domain === lane.domain);
            return (
              <View key={lane.domain} style={styles.lane}>
                <View style={[styles.laneHeader, { borderTopColor: lane.color }]}>
                  <View style={[styles.laneDot, { backgroundColor: lane.color }]} />
                  <Text style={[styles.laneLabel, { color: lane.color }]}>{lane.label}</Text>
                </View>
                {laneEvents.length === 0
                  ? <Text style={styles.laneEmpty}>—</Text>
                  : laneEvents.map((e, i) => {
                    const label = String(e.event_type ?? 'event').replace(/_/g, ' ');
                    const tsRaw = e.timestamp ? new Date(e.timestamp) : null;
                    const tsOk = tsRaw && !Number.isNaN(tsRaw.getTime())
                      ? tsRaw.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
                      : '—';
                    return (
                      <View key={`${e.sequence_num ?? i}-${label}`} style={styles.node}>
                        <Text style={styles.nodeType}>{label}</Text>
                        <Text style={styles.nodeTs}>{tsOk}</Text>
                      </View>
                    );
                  })
                }
              </View>
            );
          })}
        </View>
      </ScrollView>
    </Card>
  );
}

function phaseStyle(status: InvestigationStatus): object {
  if (status === 'COMPLETE')  return { backgroundColor: colors.status.COMPLETE.bg };
  if (status === 'FAILED')    return { backgroundColor: colors.status.FAILED.bg };
  if (status === 'HITL_PENDING') return { backgroundColor: colors.status.HITL_PENDING.bg };
  return { backgroundColor: colors.brand[50] };
}

function phaseTextStyle(status: InvestigationStatus): object {
  if (status === 'COMPLETE')  return { color: colors.status.COMPLETE.text };
  if (status === 'FAILED')    return { color: colors.status.FAILED.text };
  if (status === 'HITL_PENDING') return { color: colors.status.HITL_PENDING.text };
  return { color: colors.brand[500] };
}

const styles = StyleSheet.create({
  swimlaneScrollWeb: { flexGrow: 0, minHeight: 200 },
  header:      { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', padding: spacing[4], borderBottomWidth: 1, borderBottomColor: colors.border.light },
  heading:     { ...typography.h3, color: colors.text.primary },
  phase:       { paddingHorizontal: 10, paddingVertical: 4, borderRadius: radius.full },
  phaseText:   { ...typography.label },
  swimlanes:   { flexDirection: 'row', padding: spacing[3], gap: spacing[3] },
  lane:        { minWidth: 160, gap: spacing[2] },
  laneHeader:  { borderTopWidth: 3, paddingTop: spacing[2], flexDirection: 'row', alignItems: 'center', gap: 6, marginBottom: spacing[1] },
  laneDot:     { width: 8, height: 8, borderRadius: radius.full },
  laneLabel:   { ...typography.label },
  laneEmpty:   { ...typography.bodySm, color: colors.text.tertiary, paddingTop: 4 },
  node: {
    backgroundColor: colors.neutral[50],
    borderRadius: radius.sm,
    padding: spacing[2],
    borderWidth: 1,
    borderColor: colors.border.light,
  },
  nodeType: { ...typography.bodySm, color: colors.text.primary, fontWeight: '600', textTransform: 'capitalize' },
  nodeTs:   { ...typography.bodySm, color: colors.text.tertiary, marginTop: 2 },
});
