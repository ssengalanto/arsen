// Client-cached, non-sensitive profile. Never carries any token.
export interface UserProfile {
  id: string;
  email: string;
  emailVerified: boolean;
  createdAt: string;
}
