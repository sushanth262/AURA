// Ranked list of root-cause candidates with confidence bars and citation counts
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Card } from '@/components/ui/Card';
import { ConfidenceBar } from '@/components/ui/ConfidenceBar';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { RootCauseCandidate } from '@/types/api';

interface Props {
  candidates: RootCauseCandidate[];
}

export function RootCauseCandidateList({ candidates }: Props) {
  const sorted = [...candidates].sort((a, b) => b.confidence - a.confidence);

  return (
    <Card>
      <Text style={styles.heading}>Root Cause Candidates</Text>
      {sorted.length === 0 ? (
        <Text style={styles.empty}>No candidates identified.</Text>
      ) : (
        sorted.map((c, i) => <CandidateRow key={c.candidate_id} candidate={c} rank={i + 1} />)
      )}
    </Card>
  );
}

function CandidateRow({ candidate, rank }: { candidate: RootCauseCandidate; rank: number }) {
  return (
    <View style={[styles.row, candidate.is_primary && styles.rowPrimary]}>
      <View style={styles.rankCol}>
        <Text style={[styles.rank, candidate.is_primary && styles.rankPrimary]}>#{rank}</Text>
        {candidate.is_primary && <Text style={styles.primaryTag}>Primary</Text>}
      </View>
      <View style={styles.body}>
        <Text style={styles.description}>{candidate.description}</Text>
        <ConfidenceBar score={candidate.confidence} showLabel={false} />
        <Text style={styles.citations}>
          {candidate.citations.length} citation{candidate.citations.length !== 1 ? 's' : ''}
        </Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  heading:     { ...typography.h3, color: colors.text.primary, marginBottom: spacing[3] },
  empty:       { ...typography.bodySm, color: colors.text.tertiary, textAlign: 'center', paddingVertical: spacing[4] },
  row: {
    flexDirection: 'row', gap: 12, paddingVertical: spacing[3],
    borderBottomWidth: 1, borderBottomColor: colors.border.light,
  },
  rowPrimary:  { backgroundColor: colors.brand[50], marginHorizontal: -16, paddingHorizontal: 16, borderRadius: radius.md },
  rankCol:     { alignItems: 'center', width: 44 },
  rank:        { ...typography.h3, color: colors.text.tertiary },
  rankPrimary: { color: colors.brand[500] },
  primaryTag:  { ...typography.bodySm, color: colors.brand[500], fontWeight: '700', marginTop: 2 },
  body:        { flex: 1, gap: spacing[2] },
  description: { ...typography.body, color: colors.text.primary },
  citations:   { ...typography.bodySm, color: colors.brand[500] },
});
