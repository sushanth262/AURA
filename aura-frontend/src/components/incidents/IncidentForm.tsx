// Screen 1 — Incident Intake form: TrueStat-style inputs and severity chips
import React, { useState } from 'react';
import {
  Pressable, ScrollView, StyleSheet, Text, TextInput, View,
} from 'react-native';
import { Button } from '@/components/ui/Button';
import { Card } from '@/components/ui/Card';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';
import type {
  Artifact, ArtifactType, IncidentSubmission, Scope, Severity, TimeWindow,
} from '@/types/api';

interface Props {
  onSubmit: (payload: IncidentSubmission) => void;
  loading:  boolean;
}

const SEVERITIES: Severity[] = ['P1', 'P2', 'P3', 'P4'];
const ARTIFACT_TYPES: ArtifactType[] = [
  'STACK_TRACE', 'LOG_EXCERPT', 'ALERT_PAYLOAD', 'METRIC_SNAPSHOT', 'OTHER',
];

const SEV_COLOR: Record<Severity, string> = {
  P1: colors.severity.P1.text,
  P2: colors.severity.P2.text,
  P3: colors.severity.P3.text,
  P4: colors.severity.P4.text,
};

export function IncidentForm({ onSubmit, loading }: Props) {
  const [title, setTitle]       = useState('');
  const [severity, setSeverity] = useState<Severity>('P2');
  const [service, setService]   = useState('');
  const [cluster, setCluster]   = useState('');
  const [region, setRegion]     = useState('');
  const [since, setSince]       = useState('');
  const [symptoms, setSymptoms] = useState('');
  const [artifacts, setArtifacts] = useState<Artifact[]>([]);

  const canSubmit = title.trim().length > 0 && service.trim().length > 0 && symptoms.trim().length > 0;

  function handleSubmit() {
    if (!canSubmit) return;
    const scope: Scope = {
      service: service.trim(),
      cluster: cluster.trim() || null,
      region:  region.trim() || null,
    };
    const timeWindow: TimeWindow = {
      start: since.trim() || new Date(Date.now() - 60 * 60 * 1000).toISOString(),
      end:   null,
    };
    onSubmit({
      title:       title.trim(),
      severity,
      scope,
      time_window: timeWindow,
      symptoms:    symptoms.trim(),
      artifacts:   artifacts.length > 0 ? artifacts : undefined,
    });
  }

  function addArtifact() {
    setArtifacts((prev) => [...prev, { artifact_type: 'OTHER', content: '', source: null }]);
  }

  function updateArtifact(idx: number, patch: Partial<Artifact>) {
    setArtifacts((prev) => prev.map((a, i) => (i === idx ? { ...a, ...patch } : a)));
  }

  function removeArtifact(idx: number) {
    setArtifacts((prev) => prev.filter((_, i) => i !== idx));
  }

  return (
    <Card>
      <View style={styles.section}>
        <Label>Incident Title *</Label>
        <TextInput
          style={styles.input}
          placeholder="e.g. Payment service latency spike"
          placeholderTextColor={colors.text.tertiary}
          value={title}
          onChangeText={setTitle}
          maxLength={200}
        />
        <Text style={styles.charHint}>{title.length}/200</Text>
      </View>

      <View style={styles.section}>
        <Label>Severity *</Label>
        <View style={styles.chipRow}>
          {SEVERITIES.map((s) => (
            <Pressable
              key={s}
              onPress={() => setSeverity(s)}
              style={[
                styles.sevChip,
                severity === s && { backgroundColor: SEV_COLOR[s], borderColor: SEV_COLOR[s] },
              ]}
            >
              <Text style={[styles.sevChipText, severity === s && styles.sevChipTextActive]}>
                {s}
              </Text>
            </Pressable>
          ))}
        </View>
      </View>

      <View style={styles.section}>
        <Label>Scope *</Label>
        <TextInput
          style={styles.input}
          placeholder="Service name (required)"
          placeholderTextColor={colors.text.tertiary}
          value={service}
          onChangeText={setService}
        />
        <View style={styles.row2}>
          <TextInput
            style={[styles.input, styles.flex1]}
            placeholder="Cluster (optional)"
            placeholderTextColor={colors.text.tertiary}
            value={cluster}
            onChangeText={setCluster}
          />
          <View style={{ width: 10 }} />
          <TextInput
            style={[styles.input, styles.flex1]}
            placeholder="Region (optional)"
            placeholderTextColor={colors.text.tertiary}
            value={region}
            onChangeText={setRegion}
          />
        </View>
      </View>

      <View style={styles.section}>
        <Label>Incident Start Time</Label>
        <TextInput
          style={styles.input}
          placeholder="ISO-8601 e.g. 2025-05-03T14:00:00Z  (leave blank = 1h ago)"
          placeholderTextColor={colors.text.tertiary}
          value={since}
          onChangeText={setSince}
        />
      </View>

      <View style={styles.section}>
        <Label>Symptoms / Description *</Label>
        <TextInput
          style={[styles.input, styles.textarea]}
          placeholder="Describe what you're observing…"
          placeholderTextColor={colors.text.tertiary}
          value={symptoms}
          onChangeText={setSymptoms}
          multiline
          maxLength={2000}
          textAlignVertical="top"
        />
        <Text style={styles.charHint}>{symptoms.length}/2000</Text>
      </View>

      {/* Artifact list */}
      <View style={styles.section}>
        <View style={styles.artifactHeader}>
          <Label>Artifacts</Label>
          <Pressable onPress={addArtifact} style={styles.addBtn}>
            <Text style={styles.addBtnText}>+ Add</Text>
          </Pressable>
        </View>
        {artifacts.map((a, i) => (
          <ArtifactRow
            key={i}
            artifact={a}
            onUpdate={(p) => updateArtifact(i, p)}
            onRemove={() => removeArtifact(i)}
          />
        ))}
      </View>

      <Button
        label="Start Investigation"
        onPress={handleSubmit}
        loading={loading}
        disabled={!canSubmit}
        variant="primary"
      />
    </Card>
  );
}

function Label({ children }: { children: React.ReactNode }) {
  return <Text style={styles.label}>{children}</Text>;
}

function ArtifactRow({
  artifact, onUpdate, onRemove,
}: { artifact: Artifact; onUpdate(p: Partial<Artifact>): void; onRemove(): void }) {
  return (
    <View style={styles.artifactRow}>
      <View style={styles.artifactType}>
        <Text style={styles.artifactTypeLabel}>Type</Text>
        {ARTIFACT_TYPES.map((t) => (
          <Pressable
            key={t}
            onPress={() => onUpdate({ artifact_type: t })}
            style={[styles.typeChip, artifact.artifact_type === t && styles.typeChipActive]}
          >
            <Text style={[styles.typeChipText, artifact.artifact_type === t && styles.typeChipTextActive]}>
              {t.replace('_', ' ')}
            </Text>
          </Pressable>
        ))}
      </View>
      <TextInput
        style={[styles.input, styles.textarea]}
        placeholder="Paste content here…"
        placeholderTextColor={colors.text.tertiary}
        value={artifact.content}
        onChangeText={(v) => onUpdate({ content: v })}
        multiline
        textAlignVertical="top"
        maxLength={10240}
      />
      <TextInput
        style={styles.input}
        placeholder="Source URL or label (optional)"
        placeholderTextColor={colors.text.tertiary}
        value={artifact.source ?? ''}
        onChangeText={(v) => onUpdate({ source: v || null })}
      />
      <Pressable onPress={onRemove} style={styles.removeBtn}>
        <Text style={styles.removeBtnText}>Remove artifact</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  section:           { gap: spacing[2], marginBottom: spacing[4] },
  label:             { ...typography.label, color: colors.text.secondary, marginBottom: 2 },
  input: {
    backgroundColor: colors.canvas, borderRadius: radius.md, height: 40,
    paddingHorizontal: 12, borderWidth: 1, borderColor: colors.border.light,
    ...typography.body, color: colors.text.primary,
  },
  textarea:          { height: 96, paddingTop: 10 },
  charHint:          { ...typography.bodySm, color: colors.text.tertiary, textAlign: 'right' },
  chipRow:           { flexDirection: 'row', gap: 8, flexWrap: 'wrap' },
  sevChip: {
    paddingHorizontal: 16, paddingVertical: 7, borderRadius: radius.full,
    borderWidth: 1.5, borderColor: colors.border.medium, backgroundColor: colors.surface,
  },
  sevChipText:       { ...typography.label, color: colors.text.secondary },
  sevChipTextActive: { color: '#FFFFFF' },
  row2:              { flexDirection: 'row', marginTop: spacing[2] },
  flex1:             { flex: 1 },
  artifactHeader:    { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  addBtn:            { paddingHorizontal: 10, paddingVertical: 4 },
  addBtnText:        { ...typography.bodySm, color: colors.brand[500], fontWeight: '600' },
  artifactRow: {
    gap: spacing[2], padding: spacing[3],
    borderWidth: 1, borderColor: colors.border.light,
    borderRadius: radius.md, marginBottom: spacing[3],
    backgroundColor: colors.neutral[50],
  },
  artifactType:         { flexDirection: 'row', alignItems: 'center', flexWrap: 'wrap', gap: 6 },
  artifactTypeLabel:    { ...typography.label, color: colors.text.secondary, marginRight: 4 },
  typeChip: {
    paddingHorizontal: 8, paddingVertical: 3, borderRadius: radius.sm,
    borderWidth: 1, borderColor: colors.border.medium, backgroundColor: colors.surface,
  },
  typeChipActive:     { backgroundColor: colors.brand[500], borderColor: colors.brand[500] },
  typeChipText:       { ...typography.bodySm, color: colors.text.secondary },
  typeChipTextActive: { color: '#FFFFFF' },
  removeBtn:          { alignSelf: 'flex-end', paddingVertical: 4 },
  removeBtnText:      { ...typography.bodySm, color: colors.severity.P1.text },
});
