import { apiClient } from './client';
import type { RemediationTrigger, RemediationResponse } from '@/types/api';

export async function triggerRemediation(
  taskId: string,
  body: RemediationTrigger,
): Promise<RemediationResponse> {
  const { data } = await apiClient.post<RemediationResponse>(
    `/api/investigations/${taskId}/remediation`,
    body,
  );
  return data;
}
