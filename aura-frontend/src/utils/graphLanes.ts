import type { AgentDomain, GraphLane, GraphManifest, TaskProgressEvent } from '@/types/api';
import { colors } from '@/theme/colors';

export type Lane = { domain: AgentDomain; label: string; color: string };

export const DEFAULT_LANES: Lane[] = [
  { domain: 'supervisor', label: 'Supervisor',       color: colors.brand[500] },
  { domain: 'telemetry',  label: 'Telemetry / RCA',  color: '#3B82F6' },
  { domain: 'code',       label: 'Code / Fix',        color: '#8B5CF6' },
  { domain: 'context',    label: 'Context / Docs',    color: '#10B981' },
];

const DOMAIN_LABEL: Record<string, string> = {
  supervisor:     'Supervisor',
  telemetry:      'Telemetry / RCA',
  code:           'Code / Fix',
  context:        'Context / Docs',
  communications: 'Communications',
};

const DOMAIN_COLOR: Record<string, string> = {
  supervisor:     colors.brand[500],
  telemetry:      '#3B82F6',
  code:           '#8B5CF6',
  context:        '#10B981',
  communications: '#F59E0B',
};

function laneFromManifestEntry(entry: GraphLane): Lane | null {
  const domain = String(entry.domain ?? '').trim() as AgentDomain;
  if (!domain) return null;
  return {
    domain,
    label: entry.label || DOMAIN_LABEL[domain] || domain,
    color: entry.color || DOMAIN_COLOR[domain] || colors.brand[500],
  };
}

/** Latest GRAPH_PLANNED manifest from WS events, if any. */
export function graphManifestFromEvents(events: TaskProgressEvent[]): GraphManifest | null {
  for (let i = events.length - 1; i >= 0; i--) {
    const e = events[i];
    if (e.event_type !== 'GRAPH_PLANNED') continue;
    const raw = e.payload?.graph_manifest;
    if (!raw || typeof raw !== 'object') continue;
    return raw as GraphManifest;
  }
  return null;
}

/** Swimlanes from GRAPH_PLANNED manifest, or default four-lane layout. */
export function buildLanesFromEvents(events: TaskProgressEvent[]): Lane[] {
  const manifest = graphManifestFromEvents(events);
  if (!manifest?.lanes?.length) {
    return DEFAULT_LANES;
  }
  const lanes: Lane[] = [];
  for (const entry of manifest.lanes) {
    const lane = laneFromManifestEntry(entry);
    if (lane) lanes.push(lane);
  }
  return lanes.length > 0 ? lanes : DEFAULT_LANES;
}

export function domainLabel(domain: string): string {
  return DOMAIN_LABEL[domain] ?? domain;
}

export function domainColor(domain: string): string {
  return DOMAIN_COLOR[domain] ?? colors.brand[500];
}
