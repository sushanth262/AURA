// Screen 4a — HITL Rejection
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useMutation } from '@tanstack/react-query';
import { ScreenContainer } from '@/components/layout/ScreenContainer';
import { RejectionForm } from '@/components/hitl/RejectionForm';
import { useAuthStore } from '@/store/authStore';
import { submitHITLDecision } from '@/api/hitl';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { RejectionCategory } from '@/types/api';

export default function RejectScreen() {
  const { taskId }  = useLocalSearchParams<{ taskId: string }>();
  const router      = useRouter();
  const userId      = useAuthStore((s) => s.userId) ?? 'unknown';

  const mutation = useMutation({
    mutationFn: ({ reason, categories }: { reason: string; categories: RejectionCategory[] }) =>
      submitHITLDecision(taskId, {
        decision:    'REJECTED',
        reason,
        categories,
        reviewer_id: userId,
      }),
    onSuccess: () => {
      // Supervisor will replan — navigate back to progress to watch re-planning
      router.replace(`/investigations/${taskId}/progress` as never);
    },
  });

  return (
    <ScreenContainer>
      <View style={styles.header}>
        <Text style={styles.title}>Reject Analysis</Text>
        <Text style={styles.subtitle}>
          Provide feedback so the Supervisor can refine the investigation.
        </Text>
      </View>

      {mutation.error && (
        <View style={styles.errorBanner}>
          <Text style={styles.errorText}>
            {(mutation.error as any)?.message ?? 'Submission failed. Please try again.'}
          </Text>
        </View>
      )}

      <RejectionForm
        onSubmit={(reason, categories) => mutation.mutate({ reason, categories })}
        loading={mutation.isPending}
      />
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  header:      { gap: spacing[1] },
  title:       { ...typography.h1, color: colors.text.primary },
  subtitle:    { ...typography.body, color: colors.text.secondary },
  errorBanner: {
    backgroundColor: colors.tints.danger.bg, padding: spacing[3],
    borderRadius: 8, borderLeftWidth: 3, borderLeftColor: colors.tints.danger.text,
  },
  errorText: { ...typography.body, color: colors.tints.danger.text },
});
