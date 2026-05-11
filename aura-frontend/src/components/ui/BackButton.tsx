import React from 'react';
import { Pressable, StyleSheet, Text } from 'react-native';
import { useRouter } from 'expo-router';
import { colors } from '@/theme/colors';
import { typography } from '@/theme/typography';

interface Props {
  label?: string;
  fallbackHref?: string;
}

export function BackButton({ label = 'Back', fallbackHref = '/' }: Props) {
  const router = useRouter();

  return (
    <Pressable
      onPress={() => (router.canGoBack() ? router.back() : router.replace(fallbackHref as never))}
      style={styles.btn}
      accessibilityRole="button"
      accessibilityLabel={label}
    >
      <Text style={styles.arrow}>←</Text>
      <Text style={styles.label}>{label}</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  btn:   { flexDirection: 'row', alignItems: 'center', gap: 6, alignSelf: 'flex-start', paddingVertical: 4 },
  arrow: { fontSize: 18, color: colors.brand[500], fontWeight: '600' },
  label: { ...typography.body, color: colors.brand[500], fontWeight: '600' },
});
