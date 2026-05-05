import { create } from 'zustand';
import type { TaskProgressEvent, WSEventType } from '@/types/api';

interface InvestigationState {
  // Live WebSocket events keyed by task_id
  events:       Record<string, TaskProgressEvent[]>;
  lastSequence: Record<string, number>;

  appendEvent: (taskId: string, event: TaskProgressEvent) => void;
  clearTask:   (taskId: string) => void;
  getEvents:   (taskId: string) => TaskProgressEvent[];
  latestEventOfType: (taskId: string, type: WSEventType) => TaskProgressEvent | undefined;
}

export const useInvestigationStore = create<InvestigationState>((set, get) => ({
  events:       {},
  lastSequence: {},

  appendEvent: (taskId, event) =>
    set((s) => ({
      events:       { ...s.events,       [taskId]: [...(s.events[taskId] ?? []), event] },
      lastSequence: { ...s.lastSequence, [taskId]: event.sequence_num },
    })),

  clearTask: (taskId) =>
    set((s) => {
      const { [taskId]: _e, ...events }       = s.events;
      const { [taskId]: _s, ...lastSequence } = s.lastSequence;
      return { events, lastSequence };
    }),

  getEvents: (taskId) => get().events[taskId] ?? [],

  latestEventOfType: (taskId, type) =>
    [...(get().events[taskId] ?? [])]
      .reverse()
      .find((e) => e.event_type === type),
}));
