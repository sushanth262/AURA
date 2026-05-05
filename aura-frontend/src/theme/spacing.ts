export const spacing = {
  0:   0,
  1:   4,
  2:   8,
  3:   12,
  4:   16,
  5:   20,
  6:   24,
  8:   32,
  10:  40,
  12:  48,
  16:  64,
} as const;

export const radius = {
  sm:   6,
  md:   10,
  lg:   14,
  xl:   20,
  full: 9999,
} as const;

export const shadow = {
  // Subtle navy-tinted shadow (matches TrueStat card elevation)
  card: {
    shadowColor:   '#1B2B65',
    shadowOffset:  { width: 0, height: 2 },
    shadowOpacity: 0.07,
    shadowRadius:  8,
    elevation:     2,
  },
  modal: {
    shadowColor:   '#0F1940',
    shadowOffset:  { width: 0, height: 8 },
    shadowOpacity: 0.18,
    shadowRadius:  24,
    elevation:     12,
  },
} as const;

// Sidebar dimensions
export const layout = {
  sidebarWidth:    220,
  navBarHeight:    56,
  cardPadding:     16,
  screenPaddingH:  20,
  screenPaddingV:  20,
  maxContentWidth: 1280,
} as const;
