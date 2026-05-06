// Compact top nav for narrow/mobile viewports
import React from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { AuraLogo } from '@/components/branding/AuraLogo';
import { colors } from '@/theme/colors';
import { layout, shadow, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import { useAuthStore } from '@/store/authStore';

export function NavBar() {
  const router    = useRouter();
  const userId    = useAuthStore((s) => s.userId);
  const token     = useAuthStore((s) => s.token);
  const clearAuth = useAuthStore((s) => s.clearAuth);
  const signedIn  = Boolean(token);

  return (
    <View style={styles.bar}>
      <Pressable onPress={() => router.push('/')} hitSlop={8} accessibilityRole="button" accessibilityLabel="AURA home">
        <AuraLogo variant="navbar" />
      </Pressable>
      <View style={styles.right}>
        <Pressable
          onPress={() => (signedIn ? clearAuth() : router.push('/login' as never))}
          style={styles.btn}
          accessibilityRole="button"
          accessibilityLabel={signedIn ? 'Sign out' : 'Sign in'}
        >
          <Text style={styles.btnText}>{signedIn ? 'Sign out' : 'Sign in'}</Text>
        </Pressable>
        <Pressable onPress={() => router.push('/history')} style={styles.btn}>
          <Text style={styles.btnText}>Incident History</Text>
        </Pressable>
        {signedIn && userId ? (
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>{userId[0].toUpperCase()}</Text>
          </View>
        ) : null}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  bar: {
    height:          layout.navBarHeight,
    backgroundColor: colors.surface,
    flexDirection:   'row',
    alignItems:      'center',
    justifyContent:  'space-between',
    paddingHorizontal: spacing[4],
    borderBottomWidth: 1,
    borderBottomColor: colors.border.light,
    ...shadow.card,
  },
  right:   { flexDirection: 'row', alignItems: 'center', gap: spacing[3] },
  btn:     { paddingHorizontal: 12, paddingVertical: 6, borderRadius: 6, backgroundColor: colors.neutral[100] },
  btnText: { ...typography.bodySm, color: colors.text.secondary, fontWeight: '500' },
  avatar:  {
    width: 30, height: 30, borderRadius: 15,
    backgroundColor: colors.brand[500],
    justifyContent: 'center', alignItems: 'center',
  },
  avatarText: { color: '#FFFFFF', fontWeight: '700', fontSize: 13 },
});
