import { useMutation, useQueryClient } from '@tanstack/react-query';
import { submitHITLDecision } from '@/api/hitl';
import { evidenceKeys } from './useEvidenceBundle';
import type { HITLDecision } from '@/types/api';

export function useHITLDecision(taskId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: HITLDecision) => submitHITLDecision(taskId, body),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: evidenceKeys.bundle(taskId) }),
  });
}
