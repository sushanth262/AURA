import React from 'react';
import { Image, type StyleProp, StyleSheet, View, type ViewStyle } from 'react-native';

// Company mark: aura-frontend/icons/Aura.png
const src = require('../../../icons/Aura.png');

type Variant = 'sidebar' | 'navbar';

interface Props {
  variant?: Variant;
  /** Extra wrapper styles */
  style?: StyleProp<ViewStyle>;
}

/**
 * AURA wordmark + icon artwork — scales for sidebar (wide layout) vs compact top nav.
 */
export function AuraLogo({ variant = 'navbar', style }: Props) {
  const dims =
    variant === 'sidebar'
      ? { width: 188, height: 84 }
      : { width: 120, height: 40 };

  return (
    <View style={[styles.wrap, style]} accessibilityRole="image" accessibilityLabel="AURA company logo">
      <Image
        source={src}
        style={[styles.image, dims]}
        resizeMode="contain"
        accessibilityIgnoresInvertColors
      />
    </View>
  );
}

const styles = StyleSheet.create({
  wrap:   { justifyContent: 'center' },
  image:  { alignSelf: 'center' },
});
