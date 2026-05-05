import React, { type ReactElement } from 'react';
import { ScrollView, StyleSheet, View, type ViewProps } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { colors } from '@/theme/colors';
import { layout, spacing } from '@/theme/spacing';

interface Props extends ViewProps {
  scrollable?: boolean;
  padded?:     boolean;
  refreshControl?: ReactElement;
}

export function ScreenContainer({ scrollable = true, padded = true, refreshControl, style, children, ...rest }: Props) {
  const content = (
    <View
      style={[styles.inner, padded && styles.padded, style]}
      {...rest}
    >
      {children}
    </View>
  );

  return (
    <SafeAreaView style={styles.safe} edges={['bottom']}>
      {scrollable
        ? (
          <ScrollView
            style={styles.scroll}
            contentContainerStyle={styles.scrollContent}
            showsVerticalScrollIndicator={false}
            refreshControl={refreshControl}
          >
            {content}
          </ScrollView>
        )
        : content
      }
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe:          { flex: 1, backgroundColor: colors.canvas },
  scroll:        { flex: 1, backgroundColor: colors.canvas },
  scrollContent: { flexGrow: 1 },
  inner:         { flex: 1, backgroundColor: colors.canvas },
  padded:        { padding: layout.screenPaddingH, gap: spacing[4] },
});
