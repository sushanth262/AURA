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

  // RN-web: flex:1 does not get a definite height unless an ancestor sets one; without this the main column can be 0px tall (blank UI).
  const webShell =
    Platform.OS === 'web'
      ? ({ minHeight: '100vh' as const, width: '100%' as const })
      : null;

  if (wide) {
    return (
      <View style={[styles.row, webShell]}>
        <Sidebar />
        <View style={styles.content}>{children}</View>
      </View>
    );
  }

  return (
    <View style={[styles.col, webShell]}>
      <NavBar />
      <View style={styles.content}>{children}</View>
    </View>
  );
}

const styles = StyleSheet.create({
  row:     { flex: 1, flexDirection: 'row', backgroundColor: colors.canvas },
  col:     { flex: 1, flexDirection: 'column', backgroundColor: colors.canvas },
  // minWidth 0 lets the main pane shrink correctly beside the sidebar on web flex layouts.
  content: {
    flex:           1,
    minWidth:       0,
    minHeight:      0,
    alignSelf:      'stretch',
    maxWidth:       layout.maxContentWidth,
    width:          '100%',
    ...(Platform.OS === 'web' ? { flexGrow: 1 as const } : {}),
  },
});
