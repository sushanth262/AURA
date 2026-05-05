import React from 'react';
import { render, screen } from '@testing-library/react-native';
import { MetricCard } from '@/components/ui/MetricCard';

describe('MetricCard', () => {
  it('renders label and value', () => {
    render(<MetricCard label="Total (30d)" value="42" tint="none" />);
    expect(screen.getByText('Total (30d)')).toBeTruthy();
    expect(screen.getByText('42')).toBeTruthy();
  });

  it('renders up delta arrow', () => {
    render(<MetricCard label="Resolved" value="95%" tint="success" trend="up" delta="+5%" />);
    expect(screen.getByText('↑ +5%')).toBeTruthy();
  });

  it('renders down delta arrow', () => {
    render(<MetricCard label="Errors" value="12%" tint="danger" trend="down" delta="-3%" />);
    expect(screen.getByText('↓ -3%')).toBeTruthy();
  });

  it('does not render delta when not provided', () => {
    render(<MetricCard label="Score" value="87%" tint="none" />);
    expect(screen.queryByText(/↑|↓/)).toBeNull();
  });
});
