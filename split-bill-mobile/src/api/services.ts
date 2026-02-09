import { api } from './client';
import {
  APIResponse,
  Group,
  CreateGroupRequest,
  Bill,
  CreateBillRequest,
  Balance,
  Settlement,
  Transaction,
  CreateTransactionRequest,
  User,
  OCRResult,
  ScanReceiptRequest,
  ScanReceiptBase64Request,
  ConfirmOCRRequest,
  ImageUploadResponse,
  Activity,
  PaymentDeeplinkRequest,
  PaymentDeeplinkResponse,
  VietQRRequest,
  VietQRResponse,
  BankInfo,
  UserPaymentInfo,
  GroupStats,
  UserOverallStats,
  CategoryInfo,
} from '../types';

// ===== Auth API =====
export const authAPI = {
  verifyToken: () => api.post<APIResponse<User>>('/auth/verify-token'),

  getMe: () => api.get<APIResponse<User>>('/auth/me'),

  updateProfile: (data: Partial<User>) =>
    api.put<APIResponse<User>>('/auth/profile', data),
};

// ===== Group API =====
export const groupAPI = {
  create: (data: CreateGroupRequest) =>
    api.post<APIResponse<Group>>('/groups', data),

  list: () => api.get<APIResponse<Group[]>>('/groups'),

  getById: (id: string) => api.get<APIResponse<Group>>(`/groups/${id}`),

  update: (id: string, data: Partial<CreateGroupRequest>) =>
    api.put<APIResponse<Group>>(`/groups/${id}`, data),

  delete: (id: string) => api.delete<APIResponse<null>>(`/groups/${id}`),

  addMember: (groupId: string, userId: string, nickname?: string) =>
    api.post<APIResponse<null>>(`/groups/${groupId}/members`, {
      user_id: userId,
      nickname,
    }),

  removeMember: (groupId: string, userId: string) =>
    api.delete<APIResponse<null>>(`/groups/${groupId}/members/${userId}`),

  join: (inviteCode: string) =>
    api.post<APIResponse<Group>>('/groups/join', {invite_code: inviteCode}),
};

// ===== Bill API =====
export const billAPI = {
  create: (groupId: string, data: CreateBillRequest) =>
    api.post<APIResponse<Bill>>(`/groups/${groupId}/bills`, data),

  listByGroup: (groupId: string) =>
    api.get<APIResponse<Bill[]>>(`/groups/${groupId}/bills`),

  getById: (id: string) => api.get<APIResponse<Bill>>(`/bills/${id}`),

  update: (id: string, data: Partial<Bill>) =>
    api.put<APIResponse<Bill>>(`/bills/${id}`, data),

  delete: (id: string) => api.delete<APIResponse<null>>(`/bills/${id}`),

  getBalances: (groupId: string) =>
    api.get<APIResponse<Balance[]>>(`/groups/${groupId}/balances`),

  getSettlements: (groupId: string) =>
    api.get<APIResponse<Settlement[]>>(`/groups/${groupId}/settlements`),
};

// ===== Transaction API =====
export const transactionAPI = {
  create: (data: CreateTransactionRequest) =>
    api.post<APIResponse<Transaction>>('/transactions', data),

  confirm: (id: string) =>
    api.put<APIResponse<null>>(`/transactions/${id}/confirm`),

  getMyDebts: () =>
    api.get<APIResponse<Transaction[]>>('/users/me/debts'),
};

// ===== OCR API (Phase 2) =====
export const ocrAPI = {
  scanReceipt: (data: ScanReceiptRequest) =>
    api.post<APIResponse<OCRResult>>('/ocr/scan', data),

  scanReceiptBase64: (data: ScanReceiptBase64Request) =>
    api.post<APIResponse<OCRResult>>('/ocr/scan-base64', data),

  getResult: (id: string) =>
    api.get<APIResponse<OCRResult>>(`/ocr/${id}/result`),

  confirm: (id: string, data: ConfirmOCRRequest) =>
    api.post<APIResponse<Bill>>(`/ocr/${id}/confirm`, data),

  getPending: () =>
    api.get<APIResponse<OCRResult[]>>('/ocr/pending'),
};

// ===== Image Upload API (Phase 2) =====
export const uploadAPI = {
  uploadImage: (formData: any) =>
    api.post<APIResponse<ImageUploadResponse>>('/upload/image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),

  uploadBase64: (image: string, fileName?: string) =>
    api.post<APIResponse<ImageUploadResponse>>('/upload/image-base64', {
      image,
      file_name: fileName,
    }),
};

// ===== Payment API (Phase 4) =====
export const paymentAPI = {
  generateDeeplink: (data: PaymentDeeplinkRequest) =>
    api.post<APIResponse<PaymentDeeplinkResponse>>('/payment/deeplink', data),

  generateVietQR: (data: VietQRRequest) =>
    api.post<APIResponse<VietQRResponse>>('/payment/vietqr', data),

  getUserPaymentInfo: (userId: string) =>
    api.get<APIResponse<UserPaymentInfo>>(`/payment/user/${userId}`),

  getSupportedBanks: () =>
    api.get<APIResponse<BankInfo[]>>('/payment/banks'),
};

// ===== Activity API (Phase 4) =====
export const activityAPI = {
  getGroupActivities: (groupId: string, limit?: number) =>
    api.get<APIResponse<Activity[]>>(
      `/groups/${groupId}/activities${limit ? `?limit=${limit}` : ''}`,
    ),

  getMyActivities: (limit?: number) =>
    api.get<APIResponse<Activity[]>>(
      `/activities/me${limit ? `?limit=${limit}` : ''}`,
    ),
};

// ===== Stats API (Phase 5) =====
export const statsAPI = {
  getGroupStats: (groupId: string) =>
    api.get<APIResponse<GroupStats>>(`/groups/${groupId}/stats`),

  getGroupCategoryStats: (groupId: string) =>
    api.get<APIResponse<{categories: any[]; monthly_trend: any[]; total_spent: number}>>(
      `/groups/${groupId}/stats/categories`,
    ),

  getUserStats: () =>
    api.get<APIResponse<UserOverallStats>>('/stats/me'),

  exportGroupSummary: (groupId: string, format?: string) =>
    api.get<any>(
      `/groups/${groupId}/export${format ? `?format=${format}` : ''}`,
      { responseType: format === 'text' ? ('text' as any) : 'json' },
    ),

  getCategories: () =>
    api.get<APIResponse<CategoryInfo[]>>('/categories'),
};
