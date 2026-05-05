// TrueStat-style filter row: search input + dropdown chips
import React, { useState } from 'react';
import { Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { HistoryFilters, InvestigationStatus, Severity } from '@/types/api';

interface Props {
  filters:  HistoryFilters;
  onChange: (f: HistoryFilters) => void;
}

const SEVERITIES: Array<Severity | undefined> = [undefined, 'P1', 'P2', 'P3', 'P4'];
const STATUSES: Array<InvestigationStatus | undefined> = [undefined, 'COMPLETE', 'HITL_PENDING', 'REPLANNING', 'FAILED'];

export function FilterBar({ filters, onChange }: Props) {
  const [q, setQ] = useState(filters.q ?? '');

  const applyQ = () => onChange({ ...filters, q: q || undefined, page: 1 });

  return (
    <View style={styles.bar}>
      {/* Search */}
      <TextInput
        style={styles.search}
        placeholder="🔍  Filter by title or service…"
        placeholderTextColor={colors.text.tertiary}
        value={q}
        onChangeText={setQ}
        onSubmitEditing={applyQ}
        returnKeyType="search"
      />

      {/* Severity chips */}
      <View style={styles.chipGroup}>
        <Text style={styles.chipGroupLabel}>Severity</Text>
        {SEVERITIES.map((s) => (
          <Chip
            key={s ?? 'all'}
            label={s ?? 'All'}
            active={filters.severity === s}
            onPress={() => onChange({ ...filters, severity: s, page: 1 })}
          />
        ))}
      </View>

      {/* Status chips */}
      <View style={styles.chipGroup}>
        <Text style={styles.chipGroupLabel}>Status</Text>
        {STATUSES.map((st) => (
          <Chip
            key={st ?? 'all'}
            label={st === 'COMPLETE' ? 'Resolved' : st === 'HITL_PENDING' ? 'Awaiting' : st ?? 'All'}
            active={filters.status === st}
            onPress={() => onChange({ ...filters, status: st, page: 1 })}
          />
        ))}
      </View>
    </View>
  );
}

function Chip({ label, active, onPress }: { label: string; active: boolean; onPress(): void }) {
  return (
    <Pressable
      onPress={onPress}
      style={[styles.chip, active && styles.chipActive]}
    >
      <Text style={[styles.chipText, active && styles.chipTextActive]}>{label}</Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  bar:            { gap: spacing[3] },
  search: {
    backgroundColor: colors.surface, borderRadius: radius.md, height: 38,
    paddingHorizontal: 12, borderWidth: 1, borderColor: colors.border.light,
    ...typography.body, color: colors.text.primary,
  },
  chipGroup:      { flexDirection: 'row', alignItems: 'center', flexWrap: 'wrap', gap: 6 },
  chipGroupLabel: { ...typography.label, color: colors.text.secondary, marginRight: 4 },
  chip: {
    paddingHorizontal: 10, paddingVertical: 5, borderRadius: radius.full,
    borderWidth: 1, borderColor: colors.border.medium, backgroundColor: colors.surface,
  },
  chipActive:     { backgroundColor: colors.brand[500], borderColor: colors.brand[500] },
  chipText:       { ...typography.bodySm, color: colors.text.secondary },
  chipTextActive: { color: '#FFFFFF', fontWeight: '600' },
});
