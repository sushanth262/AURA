// Adaptive layout: sidebar on wide screens, NavBar on narrow/mobile
import React from 'react';
import { Platform, StyleSheet, useWindowDimensions, View } from 'react-native';
import { Sidebar } from './Sidebar';
import { NavBar } from './NavBar';
import { colors } from '@/theme/colors';
import { layout } from '@/theme/spacing';

interface Props { children: React.ReactNode }

export function AppShell({ children }: Props) {
  const { width } = useWindowDimensions();
  const wide = Platform.OS === 'web' && width >= 768;

  if (wide) {
    return (
      <View style={styles.row}>
        <Sidebar />
        <View style={styles.content}>{children}</View>
      </View>
    );
  }

  return (
    <View style={styles.col}>
      <NavBar />
      <View style={styles.content}>{children}</View>
    </View>
  );
}

const styles = StyleSheet.create({
  row:     { flex: 1, flexDirection: 'row', backgroundColor: colors.canvas },
  col:     { flex: 1, flexDirection: 'column', backgroundColor: colors.canvas },
  content: { flex: 1, maxWidth: layout.maxContentWidth },
});
