import { Platform } from 'react-native';

const fontFamily = Platform.select({
  ios: {
    regular: 'System',
    medium:  'System',
    semibold: 'System',
    bold:    'System',
  },
  android: {
    regular:  'Roboto',
    medium:   'Roboto',
    semibold: 'Roboto',
    bold:     'Roboto',
  },
  default: {
    regular:  'System',
    medium:   'System',
    semibold: 'System',
    bold:     'System',
  },
});

export const typography = {
  // ── Metric card large value (e.g. "501 507 ₽" in TrueStat) ──────────────
  metricValue: {
    fontSize: 26,
    fontWeight: '700' as const,
    letterSpacing: -0.5,
    fontFamily: fontFamily?.bold,
  },

  // ── Section heading ──────────────────────────────────────────────────────
  h1: { fontSize: 22, fontWeight: '700' as const, letterSpacing: -0.3, fontFamily: fontFamily?.bold },
  h2: { fontSize: 18, fontWeight: '600' as const, letterSpacing: -0.2, fontFamily: fontFamily?.semibold },
  h3: { fontSize: 15, fontWeight: '600' as const, fontFamily: fontFamily?.semibold },

  // ── Body ─────────────────────────────────────────────────────────────────
  bodyLg: { fontSize: 15, fontWeight: '400' as const, lineHeight: 22 },
  body:   { fontSize: 13, fontWeight: '400' as const, lineHeight: 20 },
  bodySm: { fontSize: 12, fontWeight: '400' as const, lineHeight: 18 },

  // ── Labels (metric card sub-label, table column headers) ─────────────────
  label:   { fontSize: 11, fontWeight: '500' as const, letterSpacing: 0.3, textTransform: 'uppercase' as const },
  labelSm: { fontSize: 10, fontWeight: '500' as const, letterSpacing: 0.2 },

  // ── Delta / percentage (colored inline value) ────────────────────────────
  delta: { fontSize: 12, fontWeight: '600' as const },

  // ── Navigation items ─────────────────────────────────────────────────────
  navItem: { fontSize: 13, fontWeight: '500' as const },

  // ── Button ───────────────────────────────────────────────────────────────
  button:   { fontSize: 14, fontWeight: '600' as const, letterSpacing: 0.1 },
  buttonSm: { fontSize: 12, fontWeight: '600' as const },

  // ── Code/mono (narrative citations, commit SHAs) ─────────────────────────
  mono: { fontSize: 12, fontFamily: Platform.select({ ios: 'Menlo', android: 'monospace', default: 'monospace' }) },
} as const;
