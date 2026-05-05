// Design language: TrueStat-inspired — navy sidebar, light-blue-gray canvas,
// color-coded metric cards (green=good, red=critical, amber=warning).

export const colors = {
  // ── Brand navy (sidebar, primary CTAs) ──────────────────────────────────
  brand: {
    50:  '#EEF2FF',
    100: '#E0E7FF',
    200: '#C7D2FE',
    500: '#1B2B65',  // sidebar fill, primary buttons
    600: '#152253',
    700: '#0F1940',
    800: '#0A1130',
  },

  // ── Success / resolved / high-confidence ────────────────────────────────
  success: {
    50:  '#F0FDF4',  // metric card background (light green like TrueStat)
    100: '#DCFCE7',
    200: '#BBF7D0',
    500: '#22C55E',  // text, icons
    600: '#16A34A',
    700: '#15803D',
  },

  // ── Warning / investigating / P2-P3 ─────────────────────────────────────
  warning: {
    50:  '#FFFBEB',  // metric card background (light amber)
    100: '#FEF3C7',
    200: '#FDE68A',
    500: '#F59E0B',
    600: '#D97706',
    700: '#B45309',
  },

  // ── Danger / critical / P1 / failed / low-confidence ────────────────────
  danger: {
    50:  '#FFF1F2',  // metric card background (light red-pink like TrueStat)
    100: '#FFE4E6',
    200: '#FECDD3',
    500: '#EF4444',
    600: '#DC2626',
    700: '#B91C1C',
  },

  // ── Info / awaiting-review / HITL ────────────────────────────────────────
  info: {
    50:  '#EFF6FF',
    100: '#DBEAFE',
    500: '#3B82F6',
    600: '#2563EB',
    700: '#1D4ED8',
  },

  // ── Severity chips ───────────────────────────────────────────────────────
  severity: {
    P1: { bg: '#FFE4E6', text: '#B91C1C', dot: '#EF4444' },
    P2: { bg: '#FEF3C7', text: '#B45309', dot: '#F59E0B' },
    P3: { bg: '#FEF3C7', text: '#B45309', dot: '#F59E0B' },
    P4: { bg: '#DCFCE7', text: '#15803D', dot: '#22C55E' },
  },

  // ── Investigation status chips ───────────────────────────────────────────
  status: {
    QUEUED:           { bg: '#F1F5F9', text: '#475569' },
    INTAKE:           { bg: '#EFF6FF', text: '#2563EB' },
    PLANNING:         { bg: '#EFF6FF', text: '#2563EB' },
    RETRIEVING:       { bg: '#EFF6FF', text: '#2563EB' },
    SYNTHESIS:        { bg: '#EFF6FF', text: '#2563EB' },
    HITL_PENDING:     { bg: '#FEF3C7', text: '#B45309' },
    REPLANNING:       { bg: '#FEF3C7', text: '#B45309' },
    REMEDIATION:      { bg: '#EFF6FF', text: '#2563EB' },
    MEMORY_WRITEBACK: { bg: '#EFF6FF', text: '#2563EB' },
    COMPLETE:         { bg: '#DCFCE7', text: '#15803D' },
    PARTIAL_EVIDENCE: { bg: '#FEF3C7', text: '#B45309' },
    FAILED:           { bg: '#FFE4E6', text: '#B91C1C' },
  },

  // ── Canvas & surfaces ────────────────────────────────────────────────────
  canvas:  '#EEF2F7',  // main page background (TrueStat blue-gray)
  surface: '#FFFFFF',  // card backgrounds
  overlay: 'rgba(15, 25, 64, 0.4)',

  // ── Borders ──────────────────────────────────────────────────────────────
  border: {
    light:  '#E2E8F0',
    medium: '#CBD5E1',
    strong: '#94A3B8',
  },

  // ── Text ─────────────────────────────────────────────────────────────────
  text: {
    primary:   '#1C2B3A',
    secondary: '#6B7A99',
    tertiary:  '#9CA3AF',
    inverse:   '#FFFFFF',
    link:      '#2563EB',
  },

  // ── Neutral scale ────────────────────────────────────────────────────────
  neutral: {
    50:  '#F8FAFC',
    100: '#F1F5F9',
    200: '#E2E8F0',
    300: '#CBD5E1',
    400: '#94A3B8',
    500: '#64748B',
    600: '#475569',
    700: '#334155',
    800: '#1E293B',
    900: '#0F172A',
  },

  // ── Semantic tint aliases (convenience for status banners, cards) ─────────
  tints: {
    success: { bg: '#F0FDF4', text: '#15803D' },
    warning: { bg: '#FFFBEB', text: '#B45309' },
    danger:  { bg: '#FFF1F2', text: '#B91C1C' },
    info:    { bg: '#EFF6FF', text: '#2563EB' },
  },
} as const;

export type Colors = typeof colors;
