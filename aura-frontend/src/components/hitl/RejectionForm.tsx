// Rejection form: category multi-select + free-text reason (Screen 4a)
import React, { useState } from 'react';
import { Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import { Button } from '@/components/ui/Button';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { RejectionCategory } from '@/types/api';

interface Props {
  onSubmit: (reason: string, categories: RejectionCategory[]) => void;
  loading:  boolean;
}

const CATEGORIES: Array<{ value: RejectionCategory; label: string }> = [
  { value: 'WRONG_SERVICE',        label: 'Wrong service' },
  { value: 'INCORRECT_TIME_WINDOW',label: 'Incorrect time window' },
  { value: 'MISSING_DATA_SOURCE',  label: 'Missing data source' },
  { value: 'ROOT_CAUSE_KNOWN',     label: 'Root cause already known' },
  { value: 'OTHER',                label: 'Other' },
];

export function RejectionForm({ onSubmit, loading }: Props) {
  const [selected, setSelected]   = useState<RejectionCategory[]>([]);
  const [reason, setReason]       = useState('');

  function toggle(cat: RejectionCategory) {
    setSelected((prev) =>
      prev.includes(cat) ? prev.filter((c) => c !== cat) : [...prev, cat],
    );
  }

  const canSubmit = reason.trim().length > 0;

  return (
    <View style={styles.container}>
      <Text style={styles.heading}>Rejection Reason</Text>
      <Text style={styles.hint}>
        Explain why you're rejecting this analysis. The Supervisor will replan based on your feedback.
      </Text>

      <Text style={styles.sectionLabel}>Categories (optional)</Text>
      <View style={styles.chips}>
        {CATEGORIES.map(({ value, label }) => {
          const active = selected.includes(value);
          return (
            <Pressable
              key={value}
              onPress={() => toggle(value)}
              style={[styles.chip, active && styles.chipActive]}
            >
              <Text style={[styles.chipText, active && styles.chipTextActive]}>{label}</Text>
            </Pressable>
          );
        })}
      </View>

      <Text style={styles.sectionLabel}>Explanation *</Text>
      <TextInput
        style={[styles.input, styles.textarea]}
        placeholder="Describe what's wrong or what's missing…"
        placeholderTextColor={colors.text.tertiary}
        value={reason}
        onChangeText={setReason}
        multiline
        textAlignVertical="top"
        maxLength={1000}
      />
      <Text style={styles.charHint}>{reason.length}/1000</Text>

      <Button
        label="Submit Rejection"
        onPress={() => onSubmit(reason.trim(), selected)}
        disabled={!canSubmit || loading}
        loading={loading}
        variant="danger"
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container:    { gap: spacing[3] },
  heading:      { ...typography.h3, color: colors.text.primary },
  hint:         { ...typography.bodySm, color: colors.text.secondary },
  sectionLabel: { ...typography.label, color: colors.text.secondary },
  chips:        { flexDirection: 'row', flexWrap: 'wrap', gap: 8 },
  chip: {
    paddingHorizontal: 12, paddingVertical: 6, borderRadius: radius.full,
    borderWidth: 1, borderColor: colors.border.medium, backgroundColor: colors.surface,
  },
  chipActive:     { backgroundColor: colors.tints.danger.bg, borderColor: colors.tints.danger.text },
  chipText:       { ...typography.bodySm, color: colors.text.secondary },
  chipTextActive: { color: colors.tints.danger.text, fontWeight: '600' },
  input: {
    backgroundColor: colors.canvas, borderRadius: radius.md,
    paddingHorizontal: 12, paddingVertical: 10,
    borderWidth: 1, borderColor: colors.border.light,
    ...typography.body, color: colors.text.primary,
  },
  textarea: { height: 96 },
  charHint: { ...typography.bodySm, color: colors.text.tertiary, textAlign: 'right' },
});
