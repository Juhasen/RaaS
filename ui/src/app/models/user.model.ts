export interface User {
  id: string;
  email: string;
  role: 'guest' | 'host' | 'admin';
  created_at: string;
}

export interface TokenResponse {
  access_token: string;
  token_type: string;
  user: User;
}
