import React from 'react';
import { StyleSheet, View, type ViewProps } from 'react-native';
import { shadow, radius } from '@/theme/spacing';

interface Props extends ViewProps {
  tint?: 'none' | 'success' | 'warning' | 'danger';
  padding?: number;
}

const TINT_BG: Record<NonNullable<Props['tint']>, string> = {
  none:    '#FFFFFF',
  success: '#F0FDF4',
  warning: '#FFFBEB',
  danger:  '#FFF1F2',
};

export function Card({ tint = 'none', padding = 16, style, children, ...rest }: Props) {
  return (
    <View
      style={[styles.card, { backgroundColor: TINT_BG[tint], padding }, style]}
      {...rest}
    >
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    borderRadius: radius.lg,
    ...shadow.card,
  },
});
