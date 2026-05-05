export type ConfidenceTier = 'HIGH' | 'MEDIUM' | 'LOW' | 'UNKNOWN';

export function getConfidenceTier(score: number | null): ConfidenceTier {
  if (score === null)  return 'UNKNOWN';
  if (score < 0.5)    return 'LOW';
  if (score < 0.75)   return 'MEDIUM';
  return 'HIGH';
}

/** Alias for convenience in tests and components */
export const confidenceTier = getConfidenceTier;

export function getConfidenceColor(score: number | null): string {
  const tier = getConfidenceTier(score);
  if (tier === 'HIGH')   return '#22C55E';
  if (tier === 'MEDIUM') return '#F59E0B';
  return '#EF4444';
}

export function getConfidenceBgColor(score: number | null): string {
  const tier = getConfidenceTier(score);
  if (tier === 'HIGH')   return '#F0FDF4';
  if (tier === 'MEDIUM') return '#FFFBEB';
  return '#FFF1F2';
}

export function formatConfidence(score: number | null): string {
  if (score === null) return '—';
  return `${Math.round(score * 100)}%`;
}
