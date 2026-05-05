import { useQuery } from '@tanstack/react-query';
import { listIncidentHistory } from '@/api/incidents';
import type { HistoryFilters } from '@/types/api';

export const historyKeys = {
  all:      ['history'] as const,
  filtered: (f: HistoryFilters) => ['history', f] as const,
};

export function useIncidentHistory(filters: HistoryFilters = {}) {
  return useQuery({
    queryKey: historyKeys.filtered(filters),
    queryFn:  () => listIncidentHistory(filters),
    staleTime: 30_000,
    placeholderData: (prev) => prev,
  });
}
