import { create } from 'zustand';
import { deleteStoredValue, getStoredValue, setStoredValue } from '@/store/tokenStorage';

interface AuthState {
  token:    string | null;
  userId:   string | null;
  tenantId: string | null;
  isReady:  boolean;
  setToken: (token: string, userId: string, tenantId: string) => Promise<void>;
  clearAuth: () => Promise<void>;
  hydrate:  () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  token:    null,
  userId:   null,
  tenantId: null,
  isReady:  false,

  setToken: async (token, userId, tenantId) => {
    await setStoredValue('auth_token', token);
    await setStoredValue('auth_user_id', userId);
    await setStoredValue('auth_tenant_id', tenantId);
    set({ token, userId, tenantId });
  },

  clearAuth: async () => {
    await deleteStoredValue('auth_token');
    await deleteStoredValue('auth_user_id');
    await deleteStoredValue('auth_tenant_id');
    set({ token: null, userId: null, tenantId: null });
  },

  hydrate: async () => {
    const [token, userId, tenantId] = await Promise.all([
      getStoredValue('auth_token'),
      getStoredValue('auth_user_id'),
      getStoredValue('auth_tenant_id'),
    ]);
    set({ token, userId, tenantId, isReady: true });
  },
}));
