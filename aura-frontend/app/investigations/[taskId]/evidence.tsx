// Screen 3 — Evidence Review
import React, { useState } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { ScreenContainer } from '@/components/layout/ScreenContainer';
import { ConfidencePanel } from '@/components/evidence/ConfidencePanel';
import { NarrativePanel } from '@/components/evidence/NarrativePanel';
import { AgentEvidenceTabs } from '@/components/evidence/AgentEvidenceTabs';
import { RootCauseCandidateList } from '@/components/evidence/RootCauseCandidateList';
import { EvidenceDeepLinkDrawer } from '@/components/evidence/EvidenceDeepLinkDrawer';
import { Button } from '@/components/ui/Button';
import { Spinner } from '@/components/ui/Spinner';
import { BackButton } from '@/components/ui/BackButton';
import { useEvidenceBundle } from '@/hooks/useEvidenceBundle';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { EvidenceRef } from '@/types/api';

export default function EvidenceScreen() {
  const params = useLocalSearchParams<{ taskId?: string | string[] }>();
  const raw    = params.taskId;
  const taskId = Array.isArray(raw) ? raw[0] : raw;
  const router = useRouter();
  const { bundle, isLoading } = useEvidenceBundle(taskId ?? '');
  const [selectedRef, setSelectedRef] = useState<EvidenceRef | null>(null);

  if (!taskId) {
    return (
      <ScreenContainer>
        <View style={styles.center}>
          <Text style={styles.emptyText}>Missing investigation task id.</Text>
        </View>
      </ScreenContainer>
    );
  }

  if (isLoading && !bundle) {
    return (
      <ScreenContainer>
        <View style={styles.center}>
          <Spinner size="large" />
          <Text style={styles.loadingText}>Fetching evidence bundle…</Text>
        </View>
      </ScreenContainer>
    );
  }

  if (!bundle) {
    return (
      <ScreenContainer>
        <View style={styles.center}>
          <Text style={styles.emptyText}>Evidence not available yet.</Text>
        </View>
      </ScreenContainer>
    );
  }

  return (
    <ScreenContainer>
      <BackButton label="Back to Investigation" />
      <View style={styles.header}>
        <Text style={styles.title}>Evidence Review</Text>
        <Text style={styles.incidentId}>
          INC-{bundle.incident_id.slice(-4).toUpperCase()} · Iteration {bundle.iteration}
        </Text>
      </View>

      <ConfidencePanel
        score={bundle.confidence_score}
        breakdown={bundle.confidence_breakdown}
      />

      <NarrativePanel
        narrative={bundle.narrative}
        iteration={bundle.iteration}
      />

      <RootCauseCandidateList candidates={bundle.root_cause_candidates} />

      <AgentEvidenceTabs
        summaries={bundle.per_agent_summaries}
        findings={bundle.agent_findings}
      />

      <View style={styles.actions}>
        <Button
          label="Approve & Remediate"
          variant="primary"
          onPress={() => router.push(`/investigations/${taskId}/approve` as never)}
          style={styles.flex1}
        />
        <Button
          label="Reject & Replan"
          variant="danger"
          onPress={() => router.push(`/investigations/${taskId}/reject` as never)}
          style={styles.flex1}
        />
      </View>

      <EvidenceDeepLinkDrawer
        item={selectedRef}
        onClose={() => setSelectedRef(null)}
      />
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  header:      { gap: spacing[1] },
  title:       { ...typography.h1, color: colors.text.primary },
  incidentId:  { ...typography.bodySm, color: colors.text.tertiary },
  center:      { flex: 1, alignItems: 'center', justifyContent: 'center', gap: spacing[3] },
  loadingText: { ...typography.body, color: colors.text.secondary },
  emptyText:   { ...typography.body, color: colors.text.tertiary },
  actions:     { flexDirection: 'row', gap: spacing[3] },
  flex1:       { flex: 1 },
});
