// Overlay B — evidence deep-link drawer: shows source metadata and external link
import React from 'react';
import { Linking, Modal, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import { Button } from '@/components/ui/Button';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type { EvidenceRef } from '@/types/api';

interface Props {
  item:    EvidenceRef | null;
  onClose: () => void;
}

export function EvidenceDeepLinkDrawer({ item: evidenceRef, onClose }: Props) {
  if (!evidenceRef) return null;

  const hasUrl = !!evidenceRef.url;

  async function openUrl() {
    if (evidenceRef?.url) {
      await Linking.openURL(evidenceRef.url);
    }
  }

  return (
    <Modal
      visible={!!evidenceRef}
      animationType="slide"
      transparent
      onRequestClose={onClose}
    >
      <Pressable style={styles.scrim} onPress={onClose} />
      <View style={styles.sheet}>
        <View style={styles.handle} />

        <View style={styles.header}>
          <View style={{ flex: 1 }}>
            <Text style={styles.label}>{evidenceRef.source_type}</Text>
            <Text style={styles.title} numberOfLines={2}>{evidenceRef.display_label}</Text>
          </View>
          <Pressable onPress={onClose} style={styles.closeBtn}>
            <Text style={styles.closeBtnText}>✕</Text>
          </Pressable>
        </View>

        <ScrollView style={styles.scroll} contentContainerStyle={styles.scrollContent}>
          <View style={styles.row}>
            <Text style={styles.metaKey}>Source ID</Text>
            <Text style={styles.metaValue} selectable>{evidenceRef.source_id}</Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.metaKey}>Ref ID</Text>
            <Text style={styles.metaValue} selectable>{evidenceRef.ref_id}</Text>
          </View>

          {evidenceRef.metadata && Object.keys(evidenceRef.metadata).length > 0 && (
            <View style={styles.metadataBlock}>
              <Text style={styles.metaHeading}>Metadata</Text>
              {Object.entries(evidenceRef.metadata).map(([k, v]) => (
                <View key={k} style={styles.row}>
                  <Text style={styles.metaKey}>{k}</Text>
                  <Text style={styles.metaValue}>{String(v)}</Text>
                </View>
              ))}
            </View>
          )}

          {evidenceRef.url && (
            <View style={styles.urlBox}>
              <Text style={styles.metaKey}>URL</Text>
              <Text style={styles.urlText} selectable numberOfLines={2}>{evidenceRef.url}</Text>
            </View>
          )}
        </ScrollView>

        {hasUrl && (
          <View style={styles.footer}>
            <Button label="Open Source →" onPress={openUrl} variant="primary" />
          </View>
        )}
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  scrim: { flex: 1, backgroundColor: 'rgba(0,0,0,0.45)' },
  sheet: {
    position: 'absolute', bottom: 0, left: 0, right: 0, maxHeight: '65%',
    backgroundColor: colors.surface, borderTopLeftRadius: radius.xl,
    borderTopRightRadius: radius.xl, paddingBottom: spacing[6],
  },
  handle: {
    alignSelf: 'center', marginTop: 10, width: 36, height: 4,
    borderRadius: radius.full, backgroundColor: colors.border.medium,
  },
  header: {
    flexDirection: 'row', alignItems: 'flex-start', justifyContent: 'space-between',
    padding: spacing[5], borderBottomWidth: 1, borderBottomColor: colors.border.light,
  },
  label:       { ...typography.label, color: colors.brand[500], marginBottom: 2 },
  title:       { ...typography.h3, color: colors.text.primary },
  closeBtn:    { padding: 6 },
  closeBtnText:{ ...typography.body, color: colors.text.secondary, fontWeight: '600' },
  scroll:      { flex: 1 },
  scrollContent:{ padding: spacing[5], gap: spacing[3] },
  row:         { flexDirection: 'row', gap: 12, alignItems: 'flex-start' },
  metaKey:     { ...typography.label, color: colors.text.secondary, width: 96 },
  metaValue:   { ...typography.body, color: colors.text.primary, flex: 1 },
  metadataBlock:{ gap: spacing[2], marginTop: spacing[2] },
  metaHeading: { ...typography.label, color: colors.text.secondary, marginBottom: spacing[1] },
  urlBox:      { gap: 4 },
  urlText:     { ...typography.bodySm, color: colors.brand[500] },
  footer:      { paddingHorizontal: spacing[5], paddingTop: spacing[3] },
});
