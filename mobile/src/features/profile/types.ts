export interface UpdateProfileInput {
  display_name?: string;
  locale?: string;
}

export interface UpdateEmailInput {
  email: string;
  current_password: string;
}

export interface UpdatePasswordInput {
  current_password: string;
  new_password: string;
}

export interface UserResponse {
  id: string;
  email: string;
  display_name: string;
  locale: 'en' | 'ar';
  email_verified: boolean;
  tenant_id: string;
  created_at: string;
}