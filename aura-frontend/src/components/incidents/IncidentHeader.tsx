// Compact incident identity strip: ID pill, title, severity badge, status chip
import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { SeverityBadge } from '@/components/ui/SeverityBadge';
import { StatusChip } from '@/components/ui/StatusChip';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { IncidentStateResponse } from '@/types/api';

interface Props {
  incident: IncidentStateResponse;
}

export function IncidentHeader({ incident }: Props) {
  const shortId = incident.incident_id
    ? `INC-${incident.incident_id.slice(-4).toUpperCase()}`
    : 'INC-????';

  return (
    <View style={styles.container}>
      <View style={styles.idPill}>
        <Text style={styles.idText}>{shortId}</Text>
      </View>
      <Text style={styles.title} numberOfLines={2}>{incident.title}</Text>
      <View style={styles.badges}>
        <SeverityBadge severity={incident.severity} size="md" />
        <StatusChip status={incident.status} size="md" />
      </View>
      {incident.scope && (
        <Text style={styles.scope}>
          {incident.scope.service}
          {incident.scope.cluster ? ` · ${incident.scope.cluster}` : ''}
          {incident.scope.region  ? ` · ${incident.scope.region}`  : ''}
        </Text>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: spacing[2] },
  idPill: {
    alignSelf: 'flex-start',
    backgroundColor: colors.brand[50],
    borderRadius: radius.full,
    paddingHorizontal: 10,
    paddingVertical: 3,
  },
  idText:  { ...typography.label, color: colors.brand[500] },
  title:   { ...typography.h2, color: colors.text.primary },
  badges:  { flexDirection: 'row', gap: 8, flexWrap: 'wrap' },
  scope:   { ...typography.bodySm, color: colors.text.secondary },
});
