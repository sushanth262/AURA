import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { submitIncident, getIncident } from '@/api/incidents';
import type { IncidentSubmission } from '@/types/api';

export const incidentKeys = {
  all:    ['incidents'] as const,
  detail: (id: string) => ['incidents', id] as const,
};

export function useGetIncident(incidentId: string) {
  return useQuery({
    queryKey: incidentKeys.detail(incidentId),
    queryFn:  () => getIncident(incidentId),
    enabled:  Boolean(incidentId),
    staleTime: 10_000,
  });
}

export function useSubmitIncident() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: IncidentSubmission) => submitIncident(body),
    onSuccess: () => qc.invalidateQueries({ queryKey: incidentKeys.all }),
  });
}
