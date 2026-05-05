// HITL approve screen: selectable action checklist with risk badges
import React, { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { ActionRisk, RecommendedAction } from '@/types/api';

interface Props {
  actions:   RecommendedAction[];
  selected:  string[];
  onToggle:  (actionId: string) => void;
}

const RISK_COLOR: Record<ActionRisk, string> = {
  Low: colors.tints.success.text,
  Med: colors.tints.warning.text,
  High: colors.tints.danger.text,
};
const RISK_BG: Record<ActionRisk, string> = {
  Low:  colors.tints.success.bg,
  Med:  colors.tints.warning.bg,
  High: colors.tints.danger.bg,
};

export function ActionChecklist({ actions, selected, onToggle }: Props) {
  return (
    <Card>
      <Text style={styles.heading}>Recommended Actions</Text>
      <Text style={styles.hint}>Select the actions to approve for automated execution.</Text>
      {actions.length === 0 && (
        <Text style={styles.empty}>No actions recommended.</Text>
      )}
      {actions.map((action) => {
        const checked = selected.includes(action.action_id);
        return (
          <Pressable
            key={action.action_id}
            onPress={() => onToggle(action.action_id)}
            style={[styles.row, checked && styles.rowChecked]}
          >
            <View style={[styles.checkbox, checked && styles.checkboxChecked]}>
              {checked && <Text style={styles.checkmark}>✓</Text>}
            </View>
            <View style={styles.body}>
              <View style={styles.meta}>
                <Text style={styles.description}>{action.description}</Text>
                <View style={styles.badges}>
                  <View style={[styles.badge, { backgroundColor: RISK_BG[action.risk] }]}>
                    <Text style={[styles.badgeText, { color: RISK_COLOR[action.risk] }]}>
                      {action.risk} Risk
                    </Text>
                  </View>
                  <View style={[styles.badge, { backgroundColor: colors.neutral[100] }]}>
                    <Text style={[styles.badgeText, { color: colors.text.secondary }]}>
                      {action.automation}
                    </Text>
                  </View>
                  {action.reversible && (
                    <View style={[styles.badge, { backgroundColor: colors.tints.success.bg }]}>
                      <Text style={[styles.badgeText, { color: colors.tints.success.text }]}>Reversible</Text>
                    </View>
                  )}
                </View>
              </View>
              {action.estimated_duration_seconds != null && (
                <Text style={styles.duration}>
                  ~{action.estimated_duration_seconds < 60
                    ? `${action.estimated_duration_seconds}s`
                    : `${Math.ceil(action.estimated_duration_seconds / 60)}m`}
                </Text>
              )}
              {action.runbook_ref && (
                <Text style={styles.runbook}>Runbook: {action.runbook_ref}</Text>
              )}
            </View>
          </Pressable>
        );
      })}
    </Card>
  );
}

const styles = StyleSheet.create({
  heading:      { ...typography.h3, color: colors.text.primary, marginBottom: spacing[1] },
  hint:         { ...typography.bodySm, color: colors.text.secondary, marginBottom: spacing[3] },
  empty:        { ...typography.bodySm, color: colors.text.tertiary, textAlign: 'center', paddingVertical: spacing[4] },
  row: {
    flexDirection: 'row', gap: 12, paddingVertical: spacing[3],
    borderBottomWidth: 1, borderBottomColor: colors.border.light,
    alignItems: 'flex-start',
  },
  rowChecked:     { backgroundColor: colors.brand[50], marginHorizontal: -16, paddingHorizontal: 16 },
  checkbox: {
    width: 22, height: 22, borderRadius: radius.sm,
    borderWidth: 2, borderColor: colors.border.medium,
    alignItems: 'center', justifyContent: 'center', marginTop: 1,
    backgroundColor: colors.surface,
  },
  checkboxChecked: { backgroundColor: colors.brand[500], borderColor: colors.brand[500] },
  checkmark:       { color: '#FFFFFF', fontSize: 13, fontWeight: '700' },
  body:            { flex: 1, gap: spacing[1] },
  meta:            { gap: spacing[1] },
  description:     { ...typography.body, color: colors.text.primary },
  badges:          { flexDirection: 'row', gap: 6, flexWrap: 'wrap' },
  badge:           { paddingHorizontal: 7, paddingVertical: 2, borderRadius: radius.full },
  badgeText:       { ...typography.bodySm, fontWeight: '600' },
  duration:        { ...typography.bodySm, color: colors.text.secondary },
  runbook:         { ...typography.bodySm, color: colors.brand[500] },
});
