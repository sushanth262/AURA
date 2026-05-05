// Re-auth PIN/password confirmation before HITL approval is submitted
import React, { useState } from 'react';
import { StyleSheet, Text, TextInput, View } from 'react-native';
import { Button } from '@/components/ui/Button';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

interface Props {
  onConfirm: (pin: string) => void;
  loading:   boolean;
  error?:    string | null;
}

export function ReAuthInput({ onConfirm, loading, error }: Props) {
  const [pin, setPin] = useState('');

  return (
    <View style={styles.container}>
      <Text style={styles.label}>Confirm your identity to approve</Text>
      <TextInput
        style={[styles.input, !!error && styles.inputError]}
        placeholder="Enter your PIN or password"
        placeholderTextColor={colors.text.tertiary}
        value={pin}
        onChangeText={setPin}
        secureTextEntry
        returnKeyType="done"
        onSubmitEditing={() => pin && onConfirm(pin)}
      />
      {error && <Text style={styles.error}>{error}</Text>}
      <Button
        label="Confirm Approval"
        onPress={() => onConfirm(pin)}
        disabled={!pin || loading}
        loading={loading}
        variant="primary"
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: spacing[3] },
  label:     { ...typography.label, color: colors.text.secondary },
  input: {
    backgroundColor: colors.canvas, borderRadius: radius.md, height: 44,
    paddingHorizontal: 12, borderWidth: 1, borderColor: colors.border.light,
    ...typography.body, color: colors.text.primary,
  },
  inputError: { borderColor: colors.tints.danger.text },
  error:      { ...typography.bodySm, color: colors.tints.danger.text },
});
