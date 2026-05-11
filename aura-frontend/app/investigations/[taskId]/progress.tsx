// Screen 2 — Live Investigation Progress
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { useLocalSearchParams, usePathname, useRouter, useSegments } from 'expo-router';
import { useQuery } from '@tanstack/react-query';
import { ScreenContainer } from '@/components/layout/ScreenContainer';
import { InvestigationGraph } from '@/components/investigation/InvestigationGraph';
import { AgentActivityPanel } from '@/components/investigation/AgentActivityPanel';
import { TimelinePanel } from '@/components/investigation/TimelinePanel';
import { IncidentHeader } from '@/components/incidents/IncidentHeader';
import { Spinner } from '@/components/ui/Spinner';
import { useInvestigationWS } from '@/hooks/useInvestigationWS';
import { useInvestigationStore } from '@/store/investigationStore';
import { getIncident, getIncidentByTaskId } from '@/api/incidents';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { ApiError, Finding, TaskProgressEvent } from '@/types/api';
import { useAuthStore } from '@/store/authStore';

/** Web static export sometimes hydrates search params late; derive task id from path / segments. */
function taskIdFromPathname(pathname: string | undefined): string {
  if (!pathname) return '';
  const m = pathname.match(/\/investigations\/([^/?#]+)/);
  return (m?.[1] ?? '').trim();
}

function taskIdFromSegments(segments: readonly string[]): string {
  const hit = segments.find((s) => /^TSK-[A-Za-z0-9]+$/.test(s));
  return hit ?? '';
}

export default function ProgressScreen() {
  const router = useRouter();
  const pathname = usePathname();
  const segments = useSegments();
  const [mounted, setMounted] = useState(false);
  const token = useAuthStore((s) => s.token);
  const isReady = useAuthStore((s) => s.isReady);
  const params = useLocalSearchParams<{ taskId?: string | string[] }>();
  const rawParam = Array.isArray(params.taskId) ? params.taskId[0] : params.taskId;
  const safeTaskId = (
    (rawParam ?? '').trim()
    || taskIdFromPathname(pathname)
    || taskIdFromSegments(segments as string[])
  ).trim();
  const events = useInvestigationStore(
    useCallback((s: { getEvents: (id: string) => TaskProgressEvent[] }) => s.getEvents(safeTaskId), [safeTaskId]),
  );

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (isReady && !token) {
      router.replace('/login' as never);
    }
  }, [isReady, token, router]);

  useInvestigationWS(safeTaskId);

  const incidentIdFromWs = String(events[0]?.incident_id ?? '').trim();

  const {
    data: incidentByTask,
    isLoading: isTaskLoading,
    isError: isTaskError,
    error: taskError,
    refetch: refetchTask,
  } = useQuery({
    queryKey: ['incidents', 'byTask', safeTaskId],
    queryFn:  () => getIncidentByTaskId(safeTaskId),
    enabled:  Boolean(safeTaskId && token && isReady),
  });

  const {
    data: incidentById,
    isLoading: isIdLoading,
    isError: isIdError,
    error: idError,
    refetch: refetchById,
  } = useQuery({
    queryKey: ['incidents', 'byId', incidentIdFromWs],
    queryFn:  () => getIncident(incidentIdFromWs),
    enabled:  Boolean(incidentIdFromWs && token && isReady && !incidentByTask),
  });

  const incident = incidentByTask ?? incidentById;
  const isLoading = isTaskLoading || (!incidentByTask && isIdLoading);
  const loadErr = !incident && (isTaskError || isIdError)
    ? String((taskError as ApiError)?.message ?? (idError as ApiError)?.message ?? 'Could not load incident')
    : null;

  const synthEvent = useMemo(
    () => events.find((e) => e.event_type === 'SYNTHESIS_COMPLETE'),
    [events],
  );
  const findings: Finding[] = (synthEvent?.payload?.findings as Finding[]) ?? [];
  const currentStatus = incident?.status ?? 'QUEUED';

  if (!mounted || !safeTaskId) {
    return (
      <ScreenContainer scrollable={false}>
        <View style={styles.fillCenter}>
          <Spinner size="large" />
        </View>
      </ScreenContainer>
    );
  }

  return (
    <ScreenContainer>
      <View style={styles.header}>
        <Text style={styles.title}>Live Investigation</Text>
        <Text style={styles.taskId}>Task {safeTaskId ? safeTaskId.slice(-8).toUpperCase() : 'PENDING'}</Text>
      </View>

      {loadErr && (
        <View style={styles.errBanner}>
          <Text style={styles.errText}>{loadErr}</Text>
          <Text style={styles.errHint}>
            If this is 404, redeploy aura-bff-api and aura-supervisor with the /v1/api/incidents/by-task endpoint.
          </Text>
          <Pressable
            onPress={() => {
              refetchTask();
              if (incidentIdFromWs) refetchById();
            }}
            accessibilityRole="button"
            accessibilityLabel="Retry loading incident"
          >
            <Text style={styles.retry}>Tap to retry</Text>
          </Pressable>
        </View>
      )}

      {isLoading && !incident ? (
        <View style={styles.fillCenter}><Spinner size="large" /></View>
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
  header:    { gap: spacing[1] },
  title:     { ...typography.h1, color: colors.text.primary },
  taskId:    { ...typography.bodySm, color: colors.text.tertiary },
  center:      { paddingVertical: spacing[8], alignItems: 'center' },
  fillCenter:  {
    flexGrow:        1,
    minHeight:       240,
    alignItems:      'center',
    justifyContent:  'center',
    paddingVertical: spacing[8],
    width:           '100%',
  },
  errBanner: {
    backgroundColor: colors.tints.danger.bg,
    padding:         spacing[3],
    borderRadius:    8,
    gap:             spacing[2],
    borderLeftWidth: 3,
    borderLeftColor: colors.tints.danger.text,
  },
  errText:  { ...typography.body, color: colors.tints.danger.text },
  errHint:  { ...typography.bodySm, color: colors.text.secondary },
  retry:    { ...typography.bodySm, color: colors.brand[500], fontWeight: '600' },
});
