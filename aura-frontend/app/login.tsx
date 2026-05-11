// Dev / demo login — obtains JWT from aura-bff-api (see docs/BFF_AUTH_LOGIN.md).
import React, { useState } from 'react';
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { fetchDevToken } from '@/api/auth';
import { ScreenContainer } from '@/components/layout/ScreenContainer';
import { useAuthStore } from '@/store/authStore';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

export default function LoginScreen() {
  const router = useRouter();
  const setToken = useAuthStore((s) => s.setToken);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const signInDemoOperator = async () => {
    setError(null);
    setLoading(true);
    try {
      const tok = await fetchDevToken({
        sub:       'demo-operator',
        roles:     ['operator'],
        tenant_id: 'demo',
      });
      await setToken(tok.access_token, tok.sub, tok.tenant_id);
      router.replace('/' as never);
    }
    catch (e: unknown) {
      const msg = e && typeof e === 'object' && 'message' in e ? String((e as { message: string }).message) : 'Login failed';
      setError(msg);
    }
    finally {
      setLoading(false);
    }
  };

  const signInDemoViewer = async () => {
    setError(null);
    setLoading(true);
    try {
      const tok = await fetchDevToken({
        sub:       'demo-viewer',
        roles:     ['viewer'],
        tenant_id: 'demo',
      });
      await setToken(tok.access_token, tok.sub, tok.tenant_id);
      router.replace('/' as never);
    }
    catch (e: unknown) {
      const msg = e && typeof e === 'object' && 'message' in e ? String((e as { message: string }).message) : 'Login failed';
      setError(msg);
    }
    finally {
      setLoading(false);
    }
  };

  return (
    <ScreenContainer>
      <Text style={styles.title}>Sign in</Text>
      <Text style={styles.sub}>
        Demo tokens are minted by the AURA backend when AUTH_DEV_MOCK=true. Use an operator for investigations;
        viewer cannot create incidents (AuthZ returns 403).
      </Text>

      {error && (
        <View style={styles.err}>
          <Text style={styles.errText}>{error}</Text>
        </View>
      )}

      <Pressable
        style={[styles.btn, styles.btnPrimary]}
        onPress={signInDemoOperator}
        disabled={loading}
        accessibilityRole="button"
        accessibilityLabel="Sign in as demo operator"
      >
        {loading
          ? <ActivityIndicator color="#fff" />
          : <Text style={styles.btnPrimaryLabel}>Demo operator (can investigate)</Text>}
      </Pressable>

      <Pressable
        style={[styles.btn, styles.btnGhost]}
        onPress={signInDemoViewer}
        disabled={loading}
        accessibilityRole="button"
        accessibilityLabel="Sign in as demo viewer"
      >
        <Text style={styles.btnGhostLabel}>Demo viewer (read-only)</Text>
      </Pressable>

      <Text style={styles.hint}>
        API base: {process.env.EXPO_PUBLIC_API_BASE_URL ?? 'http://localhost:8080/v1'}
      </Text>
    </ScreenContainer>
  );
}

const styles = StyleSheet.create({
  title:   { ...typography.h1, color: colors.text.primary, marginBottom: spacing[2] },
  sub:     { ...typography.body, color: colors.text.secondary, marginBottom: spacing[5] },
  err:     {
    backgroundColor: colors.tints.danger.bg,
    padding: spacing[3],
    borderRadius: 8,
    marginBottom: spacing[4],
  },
  errText: { ...typography.body, color: colors.tints.danger.text },
  btn:     {
    paddingVertical: spacing[3],
    paddingHorizontal: spacing[4],
    borderRadius: 8,
    marginBottom: spacing[3],
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 48,
  },
  btnPrimary:       { backgroundColor: colors.brand[500] },
  btnPrimaryLabel:  { ...typography.body, color: '#FFFFFF', fontWeight: '600' },
  btnGhost:         { borderWidth: 1, borderColor: colors.border.medium },
  btnGhostLabel:    { ...typography.body, color: colors.text.primary },
  hint:             { ...typography.bodySm, color: colors.text.tertiary, marginTop: spacing[4] },
});
