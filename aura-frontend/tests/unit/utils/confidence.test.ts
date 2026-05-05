import { formatConfidence, confidenceTier } from '@/utils/confidence';

describe('formatConfidence', () => {
  it('formats 0.87 as "87%"', () => {
    expect(formatConfidence(0.87)).toBe('87%');
  });

  it('formats 0 as "0%"', () => {
    expect(formatConfidence(0)).toBe('0%');
  });

  it('formats 1 as "100%"', () => {
    expect(formatConfidence(1)).toBe('100%');
  });

  it('returns "—" for null', () => {
    expect(formatConfidence(null)).toBe('—');
  });

  it('rounds to nearest integer', () => {
    expect(formatConfidence(0.756)).toBe('76%');
  });
});

describe('confidenceTier', () => {
  it('returns HIGH for ≥ 0.75', () => {
    expect(confidenceTier(0.75)).toBe('HIGH');
    expect(confidenceTier(1.0)).toBe('HIGH');
  });

  it('returns MEDIUM for 0.5–0.74', () => {
    expect(confidenceTier(0.5)).toBe('MEDIUM');
    expect(confidenceTier(0.74)).toBe('MEDIUM');
  });

  it('returns LOW for < 0.5', () => {
    expect(confidenceTier(0)).toBe('LOW');
    expect(confidenceTier(0.49)).toBe('LOW');
  });

  it('returns UNKNOWN for null', () => {
    expect(confidenceTier(null)).toBe('UNKNOWN');
  });
});
