import { server } from '../../mocks/server';
import { submitIncident, getIncident, listIncidentHistory } from '@/api/incidents';

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('submitIncident', () => {
  it('returns task_id and incident_id on success', async () => {
    const result = await submitIncident({
      title:       'Payment service latency spike',
      severity:    'P2',
      scope:       { service: 'payment-svc', cluster: null, region: null },
      time_window: { start: new Date().toISOString(), end: null },
      symptoms:    'p99 latency above 2s for 10 minutes',
    });

    expect(result.task_id).toBe('task-test-001');
    expect(result.incident_id).toBe('inc-test-001');
    expect(result.status).toBe('QUEUED');
  });
});

describe('getIncident', () => {
  it('returns incident state', async () => {
    const result = await getIncident('inc-test-001');
    expect(result.status).toBe('SYNTHESIS');
    expect(result.scope?.service).toBe('payment-svc');
  });
});

describe('listIncidentHistory', () => {
  it('returns paginated history with stats', async () => {
    const result = await listIncidentHistory({ page: 1, per_page: 20 });
    expect(result.items.length).toBe(1);
    expect(result.total).toBe(1);
    expect(result.stats.total_count).toBe(1);
    expect(result.stats.resolved_pct).toBe(100);
  });

  it('returns confidence score on items', async () => {
    const result = await listIncidentHistory({});
    const item = result.items[0];
    expect(item.confidence_score).toBe(0.87);
  });
});
