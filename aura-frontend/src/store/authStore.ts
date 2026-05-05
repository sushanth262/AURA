import { create } from 'zustand';
import * as SecureStore from 'expo-secure-store';

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
    await SecureStore.setItemAsync('auth_token', token);
    await SecureStore.setItemAsync('auth_user_id', userId);
    await SecureStore.setItemAsync('auth_tenant_id', tenantId);
    set({ token, userId, tenantId });
  },

  clearAuth: async () => {
    await SecureStore.deleteItemAsync('auth_token');
    await SecureStore.deleteItemAsync('auth_user_id');
    await SecureStore.deleteItemAsync('auth_tenant_id');
    set({ token: null, userId: null, tenantId: null });
  },

  hydrate: async () => {
    const [token, userId, tenantId] = await Promise.all([
      SecureStore.getItemAsync('auth_token'),
      SecureStore.getItemAsync('auth_user_id'),
      SecureStore.getItemAsync('auth_tenant_id'),
    ]);
    set({ token, userId, tenantId, isReady: true });
  },
}));
