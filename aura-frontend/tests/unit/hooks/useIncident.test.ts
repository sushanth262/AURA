import { renderHook, waitFor } from '@testing-library/react-native';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { server } from '../../mocks/server';
import { getIncident } from '@/api/incidents';

// Inline minimal hook for testing
function useIncident(incidentId: string) {
  const { useQuery } = require('@tanstack/react-query');
  return useQuery({
    queryKey: ['incidents', incidentId],
    queryFn:  () => getIncident(incidentId),
    enabled:  !!incidentId,
  });
}

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
};

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('getIncident', () => {
  it('fetches incident state from API', async () => {
    const result = await getIncident('inc-test-001');
    expect(result.incident_id).toBe('inc-test-001');
    expect(result.severity).toBe('P2');
    expect(result.title).toBe('Payment service latency spike');
  });
});
