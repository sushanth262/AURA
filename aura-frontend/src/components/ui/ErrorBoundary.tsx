import React, { Component, type ErrorInfo, type ReactNode } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

interface Props { children: ReactNode }
interface State { error: Error | null }

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack);
  }

  private reset = () => this.setState({ error: null });

  render() {
    if (!this.state.error) return this.props.children;

    return (
      <View style={styles.container}>
        <Text style={styles.heading}>Something went wrong</Text>
        <Text style={styles.message}>{this.state.error.message}</Text>
        <Pressable onPress={this.reset} style={styles.btn} accessibilityRole="button">
          <Text style={styles.btnLabel}>Try again</Text>
        </Pressable>
      </View>
    );
  }
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: spacing[6],
    gap: spacing[3],
    backgroundColor: colors.canvas,
  },
  heading:  { ...typography.h2, color: colors.tints.danger.text },
  message:  { ...typography.body, color: colors.text.secondary, textAlign: 'center' },
  btn:      { backgroundColor: colors.brand[500], paddingVertical: 10, paddingHorizontal: 20, borderRadius: 8, marginTop: spacing[2] },
  btnLabel: { ...typography.body, color: '#FFFFFF', fontWeight: '600' },
});
