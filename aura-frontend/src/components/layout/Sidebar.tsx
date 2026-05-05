// TrueStat-inspired navy sidebar with active-item highlight
import React from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { usePathname, useRouter } from 'expo-router';
import { AuraLogo } from '@/components/branding/AuraLogo';
import { colors } from '@/theme/colors';
import { layout, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

interface NavItem { label: string; icon: string; href: string }

const NAV_ITEMS: NavItem[] = [
  { label: 'New Investigation', icon: '＋', href: '/' },
  { label: 'Incident History',  icon: '☰', href: '/history' },
];

export function Sidebar() {
  const pathname = usePathname();
  const router   = useRouter();

  return (
    <View style={styles.sidebar}>
      {/* Brand — full wordmark needs contrast on navy */}
      <Pressable onPress={() => router.push('/')} style={styles.brand} accessibilityRole="button" accessibilityLabel="AURA home">
        <View style={styles.logoWell}>
          <AuraLogo variant="sidebar" />
        </View>
      </Pressable>

      {/* Nav items */}
      <View style={styles.nav}>
        {NAV_ITEMS.map((item) => {
          const active = pathname === item.href;
          return (
            <Pressable
              key={item.href}
              onPress={() => router.push(item.href as never)}
              style={[styles.navItem, active && styles.navItemActive]}
            >
              <Text style={[styles.navIcon, active && styles.navIconActive]}>{item.icon}</Text>
              <Text style={[styles.navLabel, active && styles.navLabelActive]}>
                {item.label}
              </Text>
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  sidebar: {
    width:           layout.sidebarWidth,
    backgroundColor: colors.brand[500],
    paddingTop:      spacing[6],
    paddingBottom:   spacing[6],
  },
  brand: {
    paddingHorizontal: spacing[4],
    marginBottom:   spacing[6],
  },
  logoWell: {
    backgroundColor: 'rgba(255,255,255,0.97)',
    borderRadius:    12,
    paddingVertical:   spacing[2],
    paddingHorizontal: spacing[2],
    alignItems:      'center',
  },
  nav:        { gap: 2, paddingHorizontal: spacing[2] },
  navItem: {
    flexDirection:  'row',
    alignItems:     'center',
    gap:            spacing[3],
    paddingVertical:   10,
    paddingHorizontal: spacing[3],
    borderRadius:   8,
  },
  navItemActive:   { backgroundColor: 'rgba(255,255,255,0.12)' },
  navIcon:         { fontSize: 15, color: 'rgba(255,255,255,0.6)' },
  navIconActive:   { color: '#FFFFFF' },
  navLabel:        { ...typography.navItem, color: 'rgba(255,255,255,0.7)' },
  navLabelActive:  { color: '#FFFFFF', fontWeight: '600' },
});
