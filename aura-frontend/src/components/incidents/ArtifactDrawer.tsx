// Overlay A — artifact detail drawer: slides up from bottom, shows raw content
import React from 'react';
import {
  Modal, Pressable, ScrollView, StyleSheet, Text, View,
} from 'react-native';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { Artifact } from '@/types/api';

interface Props {
  artifact: Artifact | null;
  onClose:  () => void;
}

const TYPE_LABEL: Record<string, string> = {
  STACK_TRACE:     'Stack Trace',
  LOG_EXCERPT:     'Log Excerpt',
  ALERT_PAYLOAD:   'Alert Payload',
  METRIC_SNAPSHOT: 'Metric Snapshot',
  OTHER:           'Other',
};

export function ArtifactDrawer({ artifact, onClose }: Props) {
  if (!artifact) return null;

  return (
    <Modal
      visible={!!artifact}
      animationType="slide"
      transparent
      onRequestClose={onClose}
    >
      <Pressable style={styles.scrim} onPress={onClose} />
      <View style={styles.sheet}>
        {/* Handle */}
        <View style={styles.handle} />

        {/* Header */}
        <View style={styles.header}>
          <View>
            <Text style={styles.typeLabel}>{TYPE_LABEL[artifact.artifact_type] ?? artifact.artifact_type}</Text>
            {artifact.source && (
              <Text style={styles.source} numberOfLines={1}>{artifact.source}</Text>
            )}
          </View>
          <Pressable onPress={onClose} style={styles.closeBtn}>
            <Text style={styles.closeBtnText}>✕</Text>
          </Pressable>
        </View>

        {/* Content */}
        <ScrollView style={styles.scroll} contentContainerStyle={styles.scrollContent}>
          <Text style={styles.content} selectable>{artifact.content}</Text>
        </ScrollView>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  scrim: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.45)',
  },
  sheet: {
    position: 'absolute',
    bottom: 0, left: 0, right: 0,
    maxHeight: '70%',
    backgroundColor: colors.surface,
    borderTopLeftRadius: radius.xl,
    borderTopRightRadius: radius.xl,
    paddingBottom: spacing[6],
  },
  handle: {
    alignSelf: 'center',
    marginTop: 10,
    width: 36, height: 4,
    borderRadius: radius.full,
    backgroundColor: colors.border.medium,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    paddingHorizontal: spacing[5],
    paddingTop: spacing[3],
    paddingBottom: spacing[2],
    borderBottomWidth: 1,
    borderBottomColor: colors.border.light,
  },
  typeLabel:    { ...typography.h3, color: colors.text.primary },
  source:       { ...typography.bodySm, color: colors.text.secondary, marginTop: 2 },
  closeBtn:     { padding: 6 },
  closeBtnText: { ...typography.body, color: colors.text.secondary, fontWeight: '600' },
  scroll:       { flex: 1 },
  scrollContent: { padding: spacing[5] },
  content: {
    ...typography.mono,
    color: colors.text.primary,
    lineHeight: 20,
  },
});
