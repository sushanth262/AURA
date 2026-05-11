import { create } from 'zustand';
import type { Artifact, Severity } from '@/types/api';

interface IncidentDraft {
  title: string;
  severity: Severity;
  service: string;
  cluster: string;
  region: string;
  since: string;
  symptoms: string;
  artifacts: Artifact[];
}

const INITIAL: IncidentDraft = {
  title: '',
  severity: 'P2',
  service: '',
  cluster: '',
  region: '',
  since: '',
  symptoms: '',
  artifacts: [],
};

interface IncidentDraftStore {
  draft: IncidentDraft;
  setField: <K extends keyof IncidentDraft>(key: K, value: IncidentDraft[K]) => void;
  reset: () => void;
}

export const useIncidentDraftStore = create<IncidentDraftStore>((set) => ({
  draft: { ...INITIAL },
  setField: (key, value) =>
    set((state) => ({ draft: { ...state.draft, [key]: value } })),
  reset: () => set({ draft: { ...INITIAL } }),
}));
