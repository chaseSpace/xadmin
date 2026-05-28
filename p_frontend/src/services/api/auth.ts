import { useMutation } from '@tanstack/react-query'
import { apiClient } from './client'
import { loginRequestSchema, type LoginRequestInput } from '../schemas/auth'

type LoginApiData = {
  access_token: string
  expires_at: string
  uid: number
  username: string
  display_name: string
  avatar: string
  session_id: string
}

type LoginApiResponse = {
  code: number
  message: string
  data: LoginApiData
}

export type LoginResponse = {
  token: string
  expiresAt: string
  uid: number
  username: string
  displayName: string
  avatar: string
  sessionId: string
}

export async function login(payload: LoginRequestInput): Promise<LoginResponse> {
  const validatedPayload = loginRequestSchema.parse(payload)
  const response = await apiClient.post<LoginApiResponse>('/auth/login', validatedPayload, {
    skipGlobalErrorTip: true,
    skipUnauthorizedHandler: true,
  })
  return {
    token: response.data.data.access_token,
    expiresAt: response.data.data.expires_at,
    uid: response.data.data.uid,
    username: response.data.data.username,
    displayName: response.data.data.display_name,
    avatar: response.data.data.avatar,
    sessionId: response.data.data.session_id,
  }
}

type CommonApiResponse = {
  code: number
  message: string
}

export async function logout(): Promise<void> {
  await apiClient.post<CommonApiResponse>('/auth/logout')
}

export async function logoutOthers(): Promise<void> {
  await apiClient.post<CommonApiResponse>('/auth/logout_others', {})
}

export async function forceLogout(targetUid: number): Promise<void> {
  await apiClient.post<CommonApiResponse>('/auth/force_logout', { target_uid: targetUid })
}

export async function deactivateAccount(targetUid: number): Promise<void> {
  await apiClient.post<CommonApiResponse>('/auth/deactivate', { target_uid: targetUid })
}

type SessionsApiData = {
  items: Array<{
    session_id: string
    status: string
    login_ip: string
    user_agent: string
    last_seen_at: string
    expired_at: string
    revoked_at: string
    revoked_reason: string
  }>
}

type SessionsApiResponse = {
  code: number
  message: string
  data: SessionsApiData
}

export type UserSessionItem = {
  sessionId: string
  status: string
  loginIp: string
  userAgent: string
  lastSeenAt: string
  expiredAt: string
  revokedAt: string
  revokedReason: string
}

export async function getSessions(
  status?: 'active' | 'revoked' | 'expired',
  pageSize = 10,
): Promise<UserSessionItem[]> {
  const response = await apiClient.get<SessionsApiResponse>('/auth/sessions', {
    params: status ? { status, page_size: pageSize } : { page_size: pageSize },
  })
  return response.data.data.items.map((item) => ({
    sessionId: item.session_id,
    status: item.status,
    loginIp: item.login_ip,
    userAgent: item.user_agent,
    lastSeenAt: item.last_seen_at,
    expiredAt: item.expired_at,
    revokedAt: item.revoked_at,
    revokedReason: item.revoked_reason,
  }))
}

export function useLoginMutation() {
  return useMutation({
    mutationFn: login,
  })
}
