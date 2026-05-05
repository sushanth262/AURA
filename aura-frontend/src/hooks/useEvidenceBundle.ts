import { useQuery } from '@tanstack/react-query';
import { getEvidenceBundle, isEvidenceBundle } from '@/api/investigations';
import type { EvidenceBundle } from '@/types/api';

export const evidenceKeys = {
  bundle: (taskId: string) => ['investigations', taskId, 'evidence'] as const,
};

export function useEvidenceBundle(taskId: string) {
  const query = useQuery({
    queryKey: evidenceKeys.bundle(taskId),
    queryFn:  () => getEvidenceBundle(taskId),
    enabled:  Boolean(taskId),
    // Poll every 5 s while synthesis is still in progress (202)
    refetchInterval: (query) =>
      query.state.data && isEvidenceBundle(query.state.data) ? false : 5_000,
    staleTime: 0,
  });

  const bundle: EvidenceBundle | null =
    query.data && isEvidenceBundle(query.data) ? query.data : null;

  return { ...query, bundle, isReady: bundle !== null };
}
