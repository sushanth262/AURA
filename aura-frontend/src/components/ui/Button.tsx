import React from 'react';
import { ActivityIndicator, Pressable, StyleSheet, Text, type PressableProps } from 'react-native';
import { colors } from '@/theme/colors';
import { radius } from '@/theme/spacing';
import { typography } from '@/theme/typography';

interface Props extends PressableProps {
  label:    string;
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  size?:    'sm' | 'md' | 'lg';
  loading?: boolean;
  icon?:    string;
}

const VARIANTS = {
  primary:   { bg: colors.brand[500],   text: '#FFFFFF',              border: 'transparent' },
  secondary: { bg: '#FFFFFF',           text: colors.brand[500],      border: colors.brand[500] },
  danger:    { bg: colors.danger[500],  text: '#FFFFFF',              border: 'transparent' },
  ghost:     { bg: 'transparent',       text: colors.text.secondary,  border: 'transparent' },
};

const SIZES = {
  sm: { paddingVertical: 6,  paddingHorizontal: 12, fontSize: 12, borderRadius: radius.sm },
  md: { paddingVertical: 10, paddingHorizontal: 18, fontSize: 14, borderRadius: radius.md },
  lg: { paddingVertical: 14, paddingHorizontal: 24, fontSize: 15, borderRadius: radius.md },
};

export function Button({ label, variant = 'primary', size = 'md', loading, icon, disabled, style, ...rest }: Props) {
  const v = VARIANTS[variant];
  const s = SIZES[size];
  const isDisabled = disabled || loading;

  return (
    <Pressable
      style={({ pressed }) => [
        styles.base,
        { backgroundColor: v.bg, borderColor: v.border, borderWidth: v.border !== 'transparent' ? 1.5 : 0,
          paddingVertical: s.paddingVertical, paddingHorizontal: s.paddingHorizontal, borderRadius: s.borderRadius },
        isDisabled && styles.disabled,
        pressed && styles.pressed,
        style as object,
      ]}
      disabled={isDisabled}
      {...rest}
    >
      {loading
        ? <ActivityIndicator size="small" color={v.text} />
        : <Text style={[typography.button, { color: v.text, fontSize: s.fontSize }]}>
            {icon ? `${icon}  ${label}` : label}
          </Text>
      }
    </Pressable>
  );
}

const styles = StyleSheet.create({
  base:     { flexDirection: 'row', alignItems: 'center', justifyContent: 'center' },
  disabled: { opacity: 0.5 },
  pressed:  { opacity: 0.85 },
});
