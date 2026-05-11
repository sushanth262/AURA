// Screen 4b — HITL Approval + action selection
import React, { useState } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useMutation } from '@tanstack/react-query';
import { ScreenContainer } from '@/components/layout/ScreenContainer';
import { ActionChecklist } from '@/components/hitl/ActionChecklist';
import { ReAuthInput } from '@/components/hitl/ReAuthInput';
import { Card } from '@/components/ui/Card';
import { BackButton } from '@/components/ui/BackButton';
import { useEvidenceBundle } from '@/hooks/useEvidenceBundle';
import { useAuthStore } from '@/store/authStore';
import { submitHITLDecision } from '@/api/hitl';
import { triggerRemediation } from '@/api/remediation';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

export default function ApproveScreen() {
  const { taskId }  = useLocalSearchParams<{ taskId: string }>();
  const router      = useRouter();
  const userId      = useAuthStore((s) => s.userId) ?? 'unknown';
  const { bundle }  = useEvidenceBundle(taskId);

  const [selected, setSelected] = useState<string[]>([]);
  const [authError, setAuthError] = useState<string | null>(null);

  const hitlMutation = useMutation({
    mutationFn: () =>
      submitHITLDecision(taskId, {
        decision:    'APPROVED',
        reviewer_id: userId,
      }),
  });

  const remediationMutation = useMutation({
    mutationFn: () =>
      triggerRemediation(taskId, { approved_action_ids: selected }),
    onSuccess: () => {
      router.replace(`/investigations/${taskId}/resolved` as never);
    },
  });

  async function handleConfirm(pin: string) {
    setAuthError(null);
    // In a real implementation, validate the PIN via API before submitting
    if (!pin || pin.length < 4) {
      setAuthError('PIN must be at least 4 characters.');
      return;
    }
    try {
      await hitlMutation.mutateAsync();
      await remediationMutation.mutateAsync();
    } catch (err: any) {
      setAuthError(err?.message ?? 'Approval failed. Please try again.');
    }
  }

  function toggleAction(id: string) {
    setSelected((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  }

  const loading = hitlMutation.isPending || remediationMutation.isPending;

  return (
    <ScreenContainer>
      <BackButton label="Back to Evidence" />
      <View style={styles.header}>
        <Text style={styles.title}>Approve & Remediate</Text>
        <Text style={styles.subtitle}>
          Review and select the actions to execute. Confirm your identity to proceed.
        </Text>
      </View>

      {bundle ? (
        <ActionChecklist
          actions={bundle.recommended_actions}
          selected={selected}
          onToggle={toggleAction}
        />
      ) : (
        <Card>
          <Text style={styles.loadingText}>Loading recommendations…</Text>
        </Card>
      )}

      <Card>
        <ReAuthInput
          onConfirm={handleConfirm}
          loading={loading}
          error={authError}
        />
      </Card>
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  header:      { gap: spacing[1] },
  title:       { ...typography.h1, color: colors.text.primary },
  subtitle:    { ...typography.body, color: colors.text.secondary },
  loadingText: { ...typography.body, color: colors.text.tertiary, textAlign: 'center' },
});
