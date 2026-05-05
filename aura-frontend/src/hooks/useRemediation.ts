import { useMutation } from '@tanstack/react-query';
import { triggerRemediation } from '@/api/remediation';
import type { RemediationTrigger } from '@/types/api';

export function useRemediation(taskId: string) {
  return useMutation({
    mutationFn: (body: RemediationTrigger) => triggerRemediation(taskId, body),
  });
}
