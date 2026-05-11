import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

interface Props {
  children: string;
}

interface Block {
  type: 'h2' | 'h3' | 'paragraph';
  text: string;
}

function parseBlocks(md: string): Block[] {
  const blocks: Block[] = [];
  const raw = md.split(/\n{2,}/);

  for (const chunk of raw) {
    const trimmed = chunk.trim();
    if (!trimmed) continue;

    if (trimmed.startsWith('## ')) {
      blocks.push({ type: 'h2', text: trimmed.slice(3).trim() });
    } else if (trimmed.startsWith('### ')) {
      blocks.push({ type: 'h3', text: trimmed.slice(4).trim() });
    } else {
      blocks.push({ type: 'paragraph', text: trimmed.replace(/\n/g, ' ') });
    }
  }
  return blocks;
}

function renderInline(text: string): React.ReactNode[] {
  const parts = text.split(/(\*\*[^*]+\*\*)/g);
  return parts.map((part, i) => {
    if (part.startsWith('**') && part.endsWith('**')) {
      return (
        <Text key={i} style={styles.bold}>
          {part.slice(2, -2)}
        </Text>
      );
    }
    return <Text key={i}>{part}</Text>;
  });
}

export function MarkdownText({ children }: Props) {
  const blocks = parseBlocks(children);

  return (
    <View style={styles.container}>
      {blocks.map((block, i) => {
        switch (block.type) {
          case 'h2':
            return (
              <Text key={i} style={styles.h2}>
                {renderInline(block.text)}
              </Text>
            );
          case 'h3':
            return (
              <Text key={i} style={styles.h3}>
                {renderInline(block.text)}
              </Text>
            );
          case 'paragraph':
            return (
              <Text key={i} style={styles.paragraph}>
                {renderInline(block.text)}
              </Text>
            );
        }
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: spacing[2] },
  h2:        { ...typography.h2, color: colors.text.primary, marginTop: spacing[1] },
  h3:        { ...typography.h3, color: colors.text.primary, marginTop: spacing[2], marginBottom: spacing[1] },
  paragraph: { ...typography.body, color: colors.text.primary, lineHeight: 22 },
  bold:      { fontWeight: '700' },
});
