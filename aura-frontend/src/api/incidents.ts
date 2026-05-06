import { apiClient } from './client';
import type {
  IncidentSubmission,
  IncidentQueuedResponse,
  IncidentStateResponse,
  IncidentHistoryPage,
  HistoryFilters,
} from '@/types/api';

export async function submitIncident(
  body: IncidentSubmission,
): Promise<IncidentQueuedResponse> {
  const { data } = await apiClient.post<IncidentQueuedResponse>('/api/incidents', body);
  return data;
}

export async function getIncident(incidentId: string): Promise<IncidentStateResponse> {
  const { data } = await apiClient.get<IncidentStateResponse>(
    `/api/incidents/${encodeURIComponent(incidentId)}`,
  );
  return data;
}

/** Resolved via supervisor task id — survives SPA navigations where ?incident_id is dropped on web. */
export async function getIncidentByTaskId(taskId: string): Promise<IncidentStateResponse> {
  const { data } = await apiClient.get<IncidentStateResponse>(
    `/api/incidents/by-task/${encodeURIComponent(taskId)}`,
  );
  return data;
}

export async function listIncidentHistory(
  filters: HistoryFilters = {},
): Promise<IncidentHistoryPage> {
  const { data } = await apiClient.get<IncidentHistoryPage>('/api/incidents/history', {
    params: filters,
  });
  return data;
}
