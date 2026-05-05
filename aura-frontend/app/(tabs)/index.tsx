// Screen 1 — Incident Intake
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { useMutation } from '@tanstack/react-query';
import { ScreenContainer } from '@/components/layout/ScreenContainer';
import { IncidentForm } from '@/components/incidents/IncidentForm';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import { submitIncident } from '@/api/incidents';
import type { IncidentSubmission } from '@/types/api';

export default function IncidentIntakeScreen() {
  const router = useRouter();

  const mutation = useMutation({
    mutationFn: (payload: IncidentSubmission) => submitIncident(payload),
    onSuccess: (data) => {
      router.push(`/investigations/${data.task_id}/progress` as never);
    },
  });

  return (
    <ScreenContainer>
      <View style={styles.header}>
        <Text style={styles.title}>New Incident</Text>
        <Text style={styles.subtitle}>
          Submit an alert or anomaly to start an AI-driven root cause investigation.
        </Text>
      </View>

      {mutation.error && (
        <View style={styles.errorBanner}>
          <Text style={styles.errorText}>
            {(mutation.error as any)?.message ?? 'Submission failed. Please try again.'}
          </Text>
        </View>
      )}

      <IncidentForm
        onSubmit={(payload) => mutation.mutate(payload)}
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
    backgroundColor: colors.tints.danger.bg,
    padding: spacing[3],
    borderRadius: 8,
    borderLeftWidth: 3,
    borderLeftColor: colors.tints.danger.text,
  },
  errorText: { ...typography.body, color: colors.tints.danger.text },
});
