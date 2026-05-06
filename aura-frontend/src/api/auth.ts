import axios from 'axios';

export interface DevTokenResponse {
  access_token: string;
  token_type:   string;
  expires_in:   number;
  sub:          string;
  tenant_id:    string;
  roles:        string[];
}

const base = () => process.env.EXPO_PUBLIC_API_BASE_URL ?? 'http://localhost:8080/v1';

/** Mint a demo JWT from the BFF (`aura-bff-api`, AUTH_DEV_MOCK=true only). No Authorization header. */
export async function fetchDevToken(body?: {
  sub?: string;
  roles?: string[];
  tenant_id?: string;
}): Promise<DevTokenResponse> {
  const { data } = await axios.post<DevTokenResponse>(`${base()}/auth/dev-token`, body ?? {}, {
    headers: { 'Content-Type': 'application/json' },
    timeout: 15_000,
  });
  return data;
}
