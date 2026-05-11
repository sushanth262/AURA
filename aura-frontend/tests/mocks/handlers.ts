import { http, HttpResponse } from 'msw';
import type {
  IncidentQueuedResponse, IncidentStateResponse, EvidenceBundle,
  IncidentHistoryPage, HITLResponse, RemediationResponse,
} from '@/types/api';

const BASE = 'http://localhost:8000';

export const handlers = [
  // POST /api/incidents
  http.post(`${BASE}/api/incidents`, () => {
    const body: IncidentQueuedResponse = {
      task_id:     'task-test-001',
      incident_id: 'inc-test-001',
      status:      'QUEUED',
    };
    return HttpResponse.json(body, { status: 202 });
  }),

  // GET /api/incidents/by-task/:taskId
  http.get(`${BASE}/api/incidents/by-task/:taskId`, ({ params }) => {
    const body: IncidentStateResponse = {
      incident_id: 'inc-test-001',
      task_id:     params.taskId as string,
      status:      'SYNTHESIS',
      severity:    'P2',
      title:       'Payment service latency spike',
      scope:       { service: 'payment-svc', cluster: 'prod-eu', region: 'eu-west-1' },
      created_at:  new Date().toISOString(),
      updated_at:  new Date().toISOString(),
    };
    return HttpResponse.json(body);
  }),

  // GET /api/incidents/:incidentId
  http.get(`${BASE}/api/incidents/:incidentId`, ({ params }) => {
    const body: IncidentStateResponse = {
      incident_id: params.incidentId as string,
      task_id:     'task-test-001',
      status:      'SYNTHESIS',
      severity:    'P2',
      title:       'Payment service latency spike',
      scope:       { service: 'payment-svc', cluster: 'prod-eu', region: 'eu-west-1' },
      created_at:  new Date().toISOString(),
      updated_at:  new Date().toISOString(),
    };
    return HttpResponse.json(body);
  }),

  // GET /api/investigations/:taskId/evidence — returns bundle
  http.get(`${BASE}/api/investigations/:taskId/evidence`, ({ params }) => {
    const bundle: EvidenceBundle = {
      incident_id:           'inc-test-001',
      task_id:               params.taskId as string,
      narrative:             {
        report_metadata: { service: 'payment-svc', severity: 'P2', status: 'SYNTHESIS', rca_topic: 'Payment service latency spike' },
        symptoms: 'Elevated p99 latency in payment-svc starting 14:32 UTC.',
        agent_findings: [
          { agent_name: 'Telemetry Agent', focus: 'Metrics & Logs', observation: 'Detected 40x spike in p99 latency starting 14:32 UTC. CPU and memory nominal.' },
          { agent_name: 'Code Agent', focus: 'Deployments & Commits', observation: 'Commit abc123 modified connection pool limits — deployed 14:28 UTC, 4 min before spike.' },
          { agent_name: 'Context Agent', focus: 'Tickets & Runbooks', observation: 'No related runbook found. Jira INFRA-4521 references similar issue in Jan 2025.' },
        ],
        conclusion: { summary: 'Connection pool exhaustion caused by config change in commit abc123.', confidence_level: '87%', action_item: 'Revert commit abc123 and redeploy payment-svc.' },
      },
      confidence_score:      0.87,
      confidence_breakdown:  { citation_strength: 0.9, agent_agreement: 0.85, memory_match_boost: 0.1, rejection_penalty: 0 },
      per_agent_summaries:   [
        { domain: 'telemetry', summary: 'Detected 40× spike in p99 latency starting 14:32 UTC. CPU and memory nominal.', finding_count: 3, status: 'SUCCESS', execution_duration_ms: 2200 },
        { domain: 'code',      summary: 'Commit abc123 modified connection pool limits — deployed 14:28 UTC, 4 min before spike.', finding_count: 2, status: 'SUCCESS', execution_duration_ms: 1800 },
        { domain: 'context',   summary: 'No related runbook found. Jira INFRA-4521 references similar issue in Jan 2025.', finding_count: 1, status: 'PARTIAL', execution_duration_ms: 900 },
      ],
      agent_findings:        [
        {
          finding_id: 'f-001', domain: 'telemetry', type: 'METRIC_ANOMALY',
          description: 'p99 latency exceeded 2s threshold', confidence: 0.95,
          supporting_evidence: [{ ref_id: 'r-001', source_type: 'prometheus', source_id: 'metric:payment_latency', display_label: 'payment_latency p99', url: null, metadata: {} }],
          timeline_ts: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
        },
      ],
      evidence_refs:          [],
      root_cause_candidates: [
        { candidate_id: 'rc-001', description: 'Connection pool limit change in commit abc123', confidence: 0.87, is_primary: true, citations: [] },
        { candidate_id: 'rc-002', description: 'Upstream database slowdown (unconfirmed)', confidence: 0.21, is_primary: false, citations: [] },
      ],
      recommended_actions:   [
        { action_id: 'a-001', description: 'Revert commit abc123 and redeploy', automation: 'Automated', reversible: true, risk: 'Low', estimated_duration_seconds: 120, runbook_ref: 'runbook://deploy/rollback' },
        { action_id: 'a-002', description: 'Increase connection pool limit to 200', automation: 'Manual', reversible: true, risk: 'Med', estimated_duration_seconds: null, runbook_ref: null },
      ],
      prior_incident_matches: [],
      synthesized_at:         new Date().toISOString(),
      iteration:              1,
    };
    return HttpResponse.json(bundle);
  }),

  // POST /api/investigations/:taskId/hitl
  http.post(`${BASE}/api/investigations/:taskId/hitl`, ({ params }) => {
    const body: HITLResponse = {
      status:     'HITL_RESOLVED',
      next_state: 'REMEDIATION',
      task_id:    params.taskId as string,
    };
    return HttpResponse.json(body);
  }),

  // POST /api/investigations/:taskId/remediation
  http.post(`${BASE}/api/investigations/:taskId/remediation`, () => {
    const body: RemediationResponse = { remediation_task_id: 'rem-001', status: 'QUEUED' };
    return HttpResponse.json(body, { status: 202 });
  }),

  // GET /api/incidents/history — paginated history with stats
  http.get(`${BASE}/api/incidents/history`, () => {
    const body: IncidentHistoryPage = {
      items: [
        {
          incident_id:       'inc-test-001',
          title:             'Payment service latency spike',
          severity:          'P2',
          status:            'COMPLETE',
          confidence_score:  0.87,
          created_at:        new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString(),
          resolved_at:       new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
        },
      ],
      page:     1,
      per_page: 20,
      total:    1,
      stats: {
        total_count:                   1,
        avg_time_to_diagnose_seconds:  840,
        avg_confidence_score:          0.87,
        resolved_pct:                  100,
      },
    };
    return HttpResponse.json(body);
  }),
];
