export type SessionState = {
  authenticated: boolean
  csrf_token: string
  user_id?: string
  expires_at?: string
}

export type SystemStatus = {
  app_name: string
  environment: string
  database_time: string
  bootstrapped_at: string
  applied_migrations: number
}

export const SOURCE_MEDIA = [
  'BOOK',
  'ESSAY',
  'ARTICLE',
  'PAPER',
  'SCRIPTURE',
  'LECTURE',
  'PODCAST',
  'FILM',
  'TV',
  'CONVERSATION',
  'OTHER',
] as const

export type SourceMedium = (typeof SOURCE_MEDIA)[number]

export type Source = {
  id: string
  title: string
  medium: SourceMedium
  creator?: string
  year?: number
  original_language?: string
  culture_or_context?: string
  notes?: string
  created_at: string
  updated_at: string
  archived_at?: string
}

export type SourceInput = {
  title: string
  medium: SourceMedium
  creator: string
  year: number | null
  original_language: string
  culture_or_context: string
  notes: string
}

export type ListSourcesFilters = {
  q?: string
  medium?: string
  original_language?: string
  sort?: 'recent' | 'title'
  limit?: number
}

export const INQUIRY_STATUSES = ['ACTIVE', 'DORMANT', 'SYNTHESIZED', 'ABANDONED'] as const

export type InquiryStatus = (typeof INQUIRY_STATUSES)[number]

export type Inquiry = {
  id: string
  title: string
  question: string
  status: InquiryStatus
  why_it_matters?: string
  current_view?: string
  open_tensions?: string
  created_at: string
  updated_at: string
  archived_at?: string
  engagement_count: number
  claim_count: number
  synthesis_count: number
  latest_activity?: string
}

export type InquirySummary = {
  id: string
  title: string
  question: string
  status: InquiryStatus
}

export type InquiryInput = {
  title: string
  question: string
  status: InquiryStatus
  why_it_matters: string
  current_view: string
  open_tensions: string
}

export type ListInquiriesFilters = {
  q?: string
  status?: string
  limit?: number
}

export const CLAIM_TYPES = ['OBSERVATION', 'INTERPRETATION', 'PERSONAL_VIEW', 'QUESTION', 'HYPOTHESIS'] as const

export type ClaimType = (typeof CLAIM_TYPES)[number]

export const CLAIM_STATUSES = ['ACTIVE', 'TENTATIVE', 'REVISED', 'ABANDONED'] as const

export type ClaimStatus = (typeof CLAIM_STATUSES)[number]

export type Claim = {
  id: string
  text: string
  claim_type: ClaimType
  confidence?: number
  status: ClaimStatus
  origin_engagement_id?: string
  notes?: string
  created_at: string
  updated_at: string
  archived_at?: string
  origin?: {
    engagement_id: string
    source_id: string
    source_title: string
    source_medium: SourceMedium
    portion_label?: string
  }
}

export type ClaimCreateInput = {
  text: string
  claim_type: ClaimType
  confidence: number | null
  status: ClaimStatus
  origin_engagement_id: string
  notes: string
  inquiry_ids: string[]
}

export type ClaimUpdateInput = {
  text: string
  claim_type: ClaimType
  confidence: number | null
  status: ClaimStatus
  origin_engagement_id: string
  notes: string
}

export const SYNTHESIS_TYPES = ['CHECKPOINT', 'COMPARISON', 'POSITION'] as const

export type SynthesisType = (typeof SYNTHESIS_TYPES)[number]

export type Synthesis = {
  id: string
  title: string
  body: string
  type: SynthesisType
  inquiry_id: string
  notes?: string
  created_at: string
  updated_at: string
  archived_at?: string
  inquiry?: InquirySummary
}

export type SynthesisInput = {
  title: string
  body: string
  type: SynthesisType
  inquiry_id: string
  notes: string
}

export const ACCESS_MODES = [
  'ORIGINAL',
  'TRANSLATION',
  'BILINGUAL',
  'SUBTITLED',
  'LOOKUP_HEAVY',
  'OTHER',
] as const

export type AccessMode = (typeof ACCESS_MODES)[number]

export type Engagement = {
  id: string
  source_id: string
  engaged_at: string
  portion_label?: string
  reflection: string
  why_it_matters?: string
  source_language?: string
  reflection_language?: string
  access_mode?: AccessMode
  revisit_priority?: number
  is_reread_or_rewatch: boolean
  created_at: string
  updated_at: string
  archived_at?: string
  source: {
    id: string
    title: string
    medium: SourceMedium
    creator?: string
  }
}

export type EngagementInput = {
  source_id: string
  engaged_at: string
  portion_label: string
  reflection: string
  why_it_matters: string
  source_language: string
  reflection_language: string
  access_mode: AccessMode | ''
  revisit_priority: number | null
  is_reread_or_rewatch: boolean
}

export type ListEngagementsFilters = {
  source_id?: string
  access_mode?: string
  limit?: number
}

type Envelope<T> = {
  data: T
}

class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

let csrfToken = ''

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)

  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  if (init.method && !['GET', 'HEAD'].includes(init.method.toUpperCase()) && csrfToken) {
    headers.set('X-CSRF-Token', csrfToken)
  }

  const response = await fetch(path, {
    ...init,
    credentials: 'same-origin',
    headers,
  })

  if (response.status === 204) {
    return undefined as T
  }

  const text = await response.text()
  const payload = text ? JSON.parse(text) : undefined

  if (!response.ok) {
    const message = payload?.error?.message ?? 'Request failed'
    throw new ApiError(message, response.status)
  }

  return payload as T
}

export async function getSession(): Promise<SessionState> {
  const response = await request<Envelope<SessionState>>('/api/auth/session')
  csrfToken = response.data.csrf_token
  return response.data
}

export async function login(password: string): Promise<SessionState> {
  if (!csrfToken) {
    await getSession()
  }

  await request('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ password }),
  })

  return getSession()
}

export async function logout(): Promise<SessionState> {
  await request('/api/auth/logout', {
    method: 'POST',
  })

  return getSession()
}

export async function getSystemStatus(): Promise<SystemStatus> {
  const response = await request<Envelope<SystemStatus>>('/api/system/status')
  return response.data
}

export async function listSources(filters: ListSourcesFilters = {}): Promise<Source[]> {
  const query = new URLSearchParams()

  if (filters.q) {
    query.set('q', filters.q)
  }
  if (filters.medium) {
    query.set('medium', filters.medium)
  }
  if (filters.original_language) {
    query.set('original_language', filters.original_language)
  }
  if (filters.sort) {
    query.set('sort', filters.sort)
  }
  if (filters.limit) {
    query.set('limit', String(filters.limit))
  }

  const suffix = query.toString() ? `?${query.toString()}` : ''
  const response = await request<Envelope<Source[]>>(`/api/sources${suffix}`)
  return response.data
}

export async function getSource(id: string): Promise<Source> {
  const response = await request<Envelope<Source>>(`/api/sources/${id}`)
  return response.data
}

export async function createSource(input: SourceInput): Promise<Source> {
  const response = await request<Envelope<Source>>('/api/sources', {
    method: 'POST',
    body: JSON.stringify(input),
  })
  return response.data
}

export async function updateSource(id: string, input: SourceInput): Promise<Source> {
  const response = await request<Envelope<Source>>(`/api/sources/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
  return response.data
}

export async function archiveSource(id: string): Promise<void> {
  await request(`/api/sources/${id}`, {
    method: 'DELETE',
  })
}

export async function listInquiries(filters: ListInquiriesFilters = {}): Promise<Inquiry[]> {
  const query = new URLSearchParams()

  if (filters.q) {
    query.set('q', filters.q)
  }
  if (filters.status) {
    query.set('status', filters.status)
  }
  if (filters.limit) {
    query.set('limit', String(filters.limit))
  }

  const suffix = query.toString() ? `?${query.toString()}` : ''
  const response = await request<Envelope<Inquiry[]>>(`/api/inquiries${suffix}`)
  return response.data
}

export async function getInquiry(id: string): Promise<Inquiry> {
  const response = await request<Envelope<Inquiry>>(`/api/inquiries/${id}`)
  return response.data
}

export async function createInquiry(input: InquiryInput): Promise<Inquiry> {
  const response = await request<Envelope<Inquiry>>('/api/inquiries', {
    method: 'POST',
    body: JSON.stringify(input),
  })
  return response.data
}

export async function updateInquiry(id: string, input: InquiryInput): Promise<Inquiry> {
  const response = await request<Envelope<Inquiry>>(`/api/inquiries/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
  return response.data
}

export async function archiveInquiry(id: string): Promise<void> {
  await request(`/api/inquiries/${id}`, {
    method: 'DELETE',
  })
}

export async function getClaim(id: string): Promise<Claim> {
  const response = await request<Envelope<Claim>>(`/api/claims/${id}`)
  return response.data
}

export async function createClaim(input: ClaimCreateInput): Promise<Claim> {
  const response = await request<Envelope<Claim>>('/api/claims', {
    method: 'POST',
    body: JSON.stringify(input),
  })
  return response.data
}

export async function updateClaim(id: string, input: ClaimUpdateInput): Promise<Claim> {
  const response = await request<Envelope<Claim>>(`/api/claims/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
  return response.data
}

export async function replaceClaimInquiries(id: string, inquiryIDs: string[]): Promise<void> {
  await request(`/api/claims/${id}/inquiries`, {
    method: 'PUT',
    body: JSON.stringify({ inquiry_ids: inquiryIDs }),
  })
}

export async function archiveClaim(id: string): Promise<void> {
  await request(`/api/claims/${id}`, {
    method: 'DELETE',
  })
}

export async function listEngagements(filters: ListEngagementsFilters = {}): Promise<Engagement[]> {
  const query = new URLSearchParams()

  if (filters.source_id) {
    query.set('source_id', filters.source_id)
  }
  if (filters.access_mode) {
    query.set('access_mode', filters.access_mode)
  }
  if (filters.limit) {
    query.set('limit', String(filters.limit))
  }

  const suffix = query.toString() ? `?${query.toString()}` : ''
  const response = await request<Envelope<Engagement[]>>(`/api/engagements${suffix}`)
  return response.data
}

export async function getEngagement(id: string): Promise<Engagement> {
  const response = await request<Envelope<Engagement>>(`/api/engagements/${id}`)
  return response.data
}

export async function createEngagement(input: EngagementInput): Promise<Engagement> {
  const response = await request<Envelope<Engagement>>('/api/engagements', {
    method: 'POST',
    body: JSON.stringify(input),
  })
  return response.data
}

export async function updateEngagement(id: string, input: EngagementInput): Promise<Engagement> {
  const response = await request<Envelope<Engagement>>(`/api/engagements/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
  return response.data
}

export async function archiveEngagement(id: string): Promise<void> {
  await request(`/api/engagements/${id}`, {
    method: 'DELETE',
  })
}

export async function listInquiryEngagements(inquiryId: string, limit = 20): Promise<Engagement[]> {
  const query = new URLSearchParams()
  query.set('limit', String(limit))

  const response = await request<Envelope<Engagement[]>>(`/api/inquiries/${inquiryId}/engagements?${query.toString()}`)
  return response.data
}

export async function listEngagementInquiries(engagementId: string): Promise<InquirySummary[]> {
  const response = await request<Envelope<InquirySummary[]>>(`/api/engagements/${engagementId}/inquiries`)
  return response.data
}

export async function listEngagementClaims(engagementId: string): Promise<Claim[]> {
  const response = await request<Envelope<Claim[]>>(`/api/engagements/${engagementId}/claims`)
  return response.data
}

export async function replaceEngagementInquiries(engagementId: string, inquiryIDs: string[]): Promise<void> {
  await request(`/api/engagements/${engagementId}/inquiries`, {
    method: 'PUT',
    body: JSON.stringify({ inquiry_ids: inquiryIDs }),
  })
}

export async function listInquiryClaims(inquiryId: string): Promise<Claim[]> {
  const response = await request<Envelope<Claim[]>>(`/api/inquiries/${inquiryId}/claims`)
  return response.data
}

export async function listSynthesisEligibleInquiries(limit = 6): Promise<Inquiry[]> {
  const query = new URLSearchParams()
  query.set('limit', String(limit))

  const response = await request<Envelope<Inquiry[]>>(`/api/inquiries/eligible-for-synthesis?${query.toString()}`)
  return response.data
}

export async function listSyntheses(limit = 50): Promise<Synthesis[]> {
  const query = new URLSearchParams()
  query.set('limit', String(limit))

  const response = await request<Envelope<Synthesis[]>>(`/api/syntheses?${query.toString()}`)
  return response.data
}

export async function listInquirySyntheses(inquiryId: string, limit = 50): Promise<Synthesis[]> {
  const query = new URLSearchParams()
  query.set('limit', String(limit))

  const response = await request<Envelope<Synthesis[]>>(`/api/inquiries/${inquiryId}/syntheses?${query.toString()}`)
  return response.data
}

export async function getSynthesis(id: string): Promise<Synthesis> {
  const response = await request<Envelope<Synthesis>>(`/api/syntheses/${id}`)
  return response.data
}

export async function createSynthesis(input: SynthesisInput): Promise<Synthesis> {
  const response = await request<Envelope<Synthesis>>('/api/syntheses', {
    method: 'POST',
    body: JSON.stringify(input),
  })
  return response.data
}

export async function updateSynthesis(id: string, input: SynthesisInput): Promise<Synthesis> {
  const response = await request<Envelope<Synthesis>>(`/api/syntheses/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
  return response.data
}

export async function archiveSynthesis(id: string): Promise<void> {
  await request(`/api/syntheses/${id}`, {
    method: 'DELETE',
  })
}
