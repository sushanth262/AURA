// Screen 1 — Incident Intake
import React, { useEffect } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { useMutation } from '@tanstack/react-query';
import { ScreenContainer } from '@/components/layout/ScreenContainer';
import { IncidentForm } from '@/components/incidents/IncidentForm';
import { Spinner } from '@/components/ui/Spinner';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import { submitIncident } from '@/api/incidents';
import type { IncidentSubmission } from '@/types/api';
import { useAuthStore } from '@/store/authStore';
import { useIncidentDraftStore } from '@/store/incidentDraftStore';

export default function IncidentIntakeScreen() {
  const router = useRouter();
  const token = useAuthStore((s) => s.token);
  const isReady = useAuthStore((s) => s.isReady);
  const resetDraft = useIncidentDraftStore((s) => s.reset);

  const mutation = useMutation({
    mutationFn: (payload: IncidentSubmission) => submitIncident(payload),
    onSuccess: (data) => {
      resetDraft();
      router.push({
        pathname: '/investigations/[taskId]/progress',
        params:   { taskId: data.task_id },
      } as never);
    },
  });

  useEffect(() => {
    if (isReady && !token) {
      router.replace('/login' as never);
    }
  }, [isReady, token, router]);

  if (!isReady) {
    return (
      <ScreenContainer scrollable={false}>
        <Spinner fullscreen />
      </ScreenContainer>
    );
  }

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
