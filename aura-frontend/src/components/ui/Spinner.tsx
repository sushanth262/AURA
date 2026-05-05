import React from 'react';
import { ActivityIndicator, StyleSheet, View } from 'react-native';
import { colors } from '@/theme/colors';

interface Props { size?: 'small' | 'large'; fullscreen?: boolean }

export function Spinner({ size = 'large', fullscreen = false }: Props) {
  if (fullscreen) {
    return (
      <View style={styles.fullscreen}>
        <ActivityIndicator size={size} color={colors.brand[500]} />
      </View>
    );
  }
  return <ActivityIndicator size={size} color={colors.brand[500]} />;
}

const styles = StyleSheet.create({
  fullscreen: { flex: 1, justifyContent: 'center', alignItems: 'center' },
});
