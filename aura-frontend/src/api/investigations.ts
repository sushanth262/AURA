import { apiClient } from './client';
import type { EvidenceBundle, InvestigationInProgressResponse } from '@/types/api';

export async function getEvidenceBundle(
  taskId: string,
): Promise<EvidenceBundle | InvestigationInProgressResponse> {
  const { data, status } = await apiClient.get<EvidenceBundle | InvestigationInProgressResponse>(
    `/api/investigations/${taskId}/evidence`,
    { validateStatus: (s) => s === 200 || s === 202 },
  );
  return data;
}

export function isEvidenceBundle(
  r: EvidenceBundle | InvestigationInProgressResponse,
): r is EvidenceBundle {
  return 'narrative' in r;
}
