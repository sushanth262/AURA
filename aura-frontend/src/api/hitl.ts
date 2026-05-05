import { apiClient } from './client';
import type { HITLDecision, HITLResponse } from '@/types/api';

export async function submitHITLDecision(
  taskId: string,
  body: HITLDecision,
): Promise<HITLResponse> {
  const { data } = await apiClient.post<HITLResponse>(
    `/api/investigations/${taskId}/hitl`,
    body,
  );
  return data;
}
