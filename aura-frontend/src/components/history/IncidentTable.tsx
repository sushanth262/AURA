// TrueStat-inspired data table: color-coded rows, sortable columns, pagination
import React from 'react';
import { Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { useRouter } from 'expo-router';
import { Card } from '@/components/ui/Card';
import { SeverityBadge } from '@/components/ui/SeverityBadge';
import { StatusChip } from '@/components/ui/StatusChip';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import { formatAge, formatConfidence } from '@/utils/formatting';
import { formatConfidence as fmtConf } from '@/utils/confidence';
import type { IncidentHistoryPage, IncidentSummary } from '@/types/api';

interface Props {
  page:     IncidentHistoryPage;
  onPage:   (p: number) => void;
}

export function IncidentTable({ page, onPage }: Props) {
  const router = useRouter();

  return (
    <Card padding={0}>
      <ScrollView horizontal showsHorizontalScrollIndicator={false}>
        <View style={{ minWidth: 680 }}>
          {/* Header */}
          <View style={[styles.row, styles.header]}>
            {['ID', 'Title', 'Severity', 'Status', 'Confidence', 'Age'].map((col) => (
              <Text key={col} style={[styles.cell, styles.headerCell, COL_FLEX[col] && { flex: COL_FLEX[col] }]}>
                {col}
              </Text>
            ))}
          </View>

          {/* Rows */}
          {page.items.length === 0 ? (
            <View style={styles.emptyWrap}>
              <Text style={styles.emptyText}>No incidents to display.</Text>
              <Text style={styles.emptyHint}>Adjust filters or check back after new investigations complete.</Text>
            </View>
          ) : (
            page.items.map((item, idx) => (
              <IncidentRow
                key={item.incident_id}
                item={item}
                idx={idx}
                onPress={() => router.push(`/investigations/${item.incident_id}/evidence` as never)}
              />
            ))
          )}
        </View>
      </ScrollView>

      {/* Pagination */}
      <View style={styles.pagination}>
        <Pressable onPress={() => onPage(page.page - 1)} disabled={page.page <= 1} style={styles.pageBtn}>
          <Text style={[styles.pageBtnText, page.page <= 1 && styles.disabled]}>← Previous</Text>
        </Pressable>
        <Text style={styles.pageInfo}>
          Page {page.page} of {Math.ceil(page.total / page.per_page) || 1}
        </Text>
        <Pressable
          onPress={() => onPage(page.page + 1)}
          disabled={page.page * page.per_page >= page.total}
          style={styles.pageBtn}
        >
          <Text style={[styles.pageBtnText, page.page * page.per_page >= page.total && styles.disabled]}>
            Next →
          </Text>
        </Pressable>
      </View>
    </Card>
  );
}

function IncidentRow({ item, idx, onPress }: { item: IncidentSummary; idx: number; onPress(): void }) {
  return (
    <Pressable
      onPress={onPress}
      style={({ pressed }) => [
        styles.row,
        idx % 2 === 1 && styles.rowAlt,
        pressed && styles.rowPressed,
      ]}
    >
      <Text style={[styles.cell, { flex: COL_FLEX.ID }]} numberOfLines={1}>
        INC-{item.incident_id.slice(-4).toUpperCase()}
      </Text>
      <Text style={[styles.cell, { flex: COL_FLEX.Title }]} numberOfLines={1}>
        {item.title}
      </Text>
      <View style={[styles.cell, { flex: COL_FLEX.Severity }]}>
        <SeverityBadge severity={item.severity} size="sm" />
      </View>
      <View style={[styles.cell, { flex: COL_FLEX.Status }]}>
        <StatusChip status={item.status} size="sm" />
      </View>
      <Text style={[styles.cell, styles.confCell, { flex: COL_FLEX.Confidence }]}>
        {fmtConf(item.confidence_score ?? null)}
      </Text>
      <Text style={[styles.cell, styles.ageCell, { flex: COL_FLEX.Age }]}>
        {formatAge(item.created_at)}
      </Text>
    </Pressable>
  );
}

const COL_FLEX: Record<string, number> = {
  ID: 1, Title: 3, Severity: 1, Status: 1.5, Confidence: 1, Age: 0.7,
};

const styles = StyleSheet.create({
  row:         { flexDirection: 'row', alignItems: 'center', paddingHorizontal: 16, paddingVertical: 10, borderBottomWidth: 1, borderBottomColor: colors.border.light },
  rowAlt:      { backgroundColor: colors.neutral[50] },
  rowPressed:  { backgroundColor: colors.brand[50] },
  header:      { backgroundColor: colors.neutral[100], paddingVertical: 8 },
  cell:        { flex: 1, paddingRight: 8, ...typography.body, color: colors.text.primary },
  headerCell:  { ...typography.label, color: colors.text.secondary },
  confCell:    { fontWeight: '600', color: colors.text.primary },
  ageCell:     { color: colors.text.secondary },
  pagination:  { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', padding: 12, borderTopWidth: 1, borderTopColor: colors.border.light },
  pageBtn:     { paddingHorizontal: 12, paddingVertical: 6 },
  pageBtnText: { ...typography.bodySm, color: colors.brand[500], fontWeight: '600' },
  pageInfo:    { ...typography.bodySm, color: colors.text.secondary },
  disabled:    { color: colors.text.tertiary },
  emptyWrap:   { paddingVertical: 32, paddingHorizontal: 16, borderBottomWidth: 1, borderBottomColor: colors.border.light },
  emptyText:   { ...typography.body, color: colors.text.secondary, textAlign: 'center' },
  emptyHint:   { ...typography.bodySm, color: colors.text.tertiary, textAlign: 'center', marginTop: 6 },
});
