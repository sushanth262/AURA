import { buildLanesFromEvents, DEFAULT_LANES } from '@/utils/graphLanes';
import type { TaskProgressEvent } from '@/types/api';

describe('buildLanesFromEvents', () => {
  it('returns default four lanes without GRAPH_PLANNED', () => {
    const lanes = buildLanesFromEvents([]);
    expect(lanes).toHaveLength(4);
    expect(lanes.map((l) => l.domain)).toEqual(DEFAULT_LANES.map((l) => l.domain));
  });

  it('builds five lanes from GRAPH_PLANNED manifest', () => {
    const events: TaskProgressEvent[] = [
      {
        task_id:      'TSK-1',
        incident_id:  'INC-1',
        event_type:   'GRAPH_PLANNED',
        agent_domain: 'supervisor',
        sequence_num: 4,
        timestamp:    '2026-05-28T20:00:00Z',
        payload: {
          graph_manifest: {
            graph_version: 1,
            lanes: [
              { domain: 'supervisor',     label: 'Supervisor',       color: '#1B2B65' },
              { domain: 'telemetry',      label: 'Telemetry / RCA',  color: '#3B82F6' },
              { domain: 'code',           label: 'Code / Fix',       color: '#8B5CF6' },
              { domain: 'context',        label: 'Context / Docs',   color: '#10B981' },
              { domain: 'communications', label: 'Communications', color: '#F59E0B' },
            ],
          },
        },
      },
    ];
    const lanes = buildLanesFromEvents(events);
    expect(lanes).toHaveLength(5);
    expect(lanes[4].label).toBe('Communications');
    expect(lanes[4].domain).toBe('communications');
  });
});
