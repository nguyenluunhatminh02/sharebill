// User types
export interface BankAccount {
  bank_code: string;
  account_number: string;
  account_name: string;
}

export interface User {
  id: string;
  phone: string;
  display_name: string;
  avatar_url: string;
  bank_accounts: BankAccount[];
  preferred_payment: string;
  created_at: string;
}

// Group types
export type MemberRole = 'admin' | 'member';

export interface GroupMember {
  user_id: string;
  nickname: string;
  display_name: string;
  avatar_url: string;
  role: MemberRole;
  joined_at: string;
}

export interface Group {
  id: string;
  name: string;
  description: string;
  avatar_url: string;
  created_by: string;
  members: GroupMember[];
  invite_code: string;
  is_active: boolean;
  created_at: string;
}

// Bill types
export type SplitType = 'equal' | 'by_item' | 'by_percentage' | 'by_amount';
export type BillStatus = 'pending' | 'settled' | 'cancelled';

export interface BillItem {
  id: string;
  name: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  assigned_to: string[];
}

export interface ExtraCharges {
  tax: number;
  service_charge: number;
  tip: number;
  discount: number;
}

export interface BillSplit {
  user_id: string;
  display_name: string;
  amount: number;
  is_paid: boolean;
  paid_at?: string;
}

export interface Bill {
  id: string;
  group_id: string;
  title: string;
  description: string;
  category: string;
  receipt_image_url: string;
  total_amount: number;
  currency: string;
  paid_by: string;
  paid_by_name: string;
  split_type: SplitType;
  items: BillItem[];
  extra_charges: ExtraCharges;
  splits: BillSplit[];
  status: BillStatus;
  created_at: string;
}

// Transaction types
export type TransactionStatus = 'pending' | 'confirmed' | 'rejected';

export interface Transaction {
  id: string;
  group_id: string;
  from_user: string;
  from_user_name: string;
  to_user: string;
  to_user_name: string;
  amount: number;
  currency: string;
  bill_id?: string;
  type: 'payment' | 'settlement';
  status: TransactionStatus;
  payment_method: string;
  note: string;
  created_at: string;
  confirmed_at?: string;
}

// Settlement types
export interface Settlement {
  from_user_id: string;
  from_user_name: string;
  to_user_id: string;
  to_user_name: string;
  amount: number;
}

export interface Balance {
  user_id: string;
  display_name: string;
  balance: number;
}

// API Response
export interface APIResponse<T> {
  success: boolean;
  message: string;
  data: T;
  error?: string;
}

// Create requests
export interface CreateGroupRequest {
  name: string;
  description?: string;
  avatar_url?: string;
}

export interface CreateBillRequest {
  title: string;
  description?: string;
  category?: string;
  total_amount: number;
  currency: string;
  paid_by: string;
  split_type: SplitType;
  items?: CreateBillItemRequest[];
  extra_charges?: ExtraCharges;
  split_among?: string[];
}

// ===== OCR Types (Phase 2) =====
export type OCRStatus = 'processing' | 'completed' | 'failed' | 'confirmed';

export interface ParsedItem {
  name: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  confidence: number;
}

export interface OCRResult {
  id: string;
  image_url: string;
  raw_text: string;
  parsed_items: ParsedItem[];
  parsed_total: number;
  parsed_tax: number;
  parsed_service_fee: number;
  parsed_discount: number;
  confidence_score: number;
  processing_time_ms: number;
  status: OCRStatus;
  created_at: string;
}

export interface ScanReceiptRequest {
  group_id: string;
  image_url: string;
}

export interface ScanReceiptBase64Request {
  group_id: string;
  image_base64: string;
  file_name?: string;
}

export interface ConfirmOCRRequest {
  title: string;
  items: ParsedItem[];
  total: number;
  tax: number;
  service_fee: number;
  discount: number;
  paid_by: string;
  split_type: SplitType;
  split_among?: string[];
}

export interface ImageUploadResponse {
  url: string;
  file_name: string;
  size: number;
}

export interface CreateBillItemRequest {
  name: string;
  quantity: number;
  unit_price: number;
  total_price?: number;
  assigned_to?: string[];
}

export interface CreateTransactionRequest {
  group_id: string;
  to_user: string;
  amount: number;
  currency: string;
  bill_id?: string;
  payment_method?: string;
  note?: string;
}

// ===== Payment Types (Phase 4) =====
export type ActivityType =
  | 'bill_created'
  | 'bill_deleted'
  | 'bill_updated'
  | 'member_joined'
  | 'member_left'
  | 'payment_sent'
  | 'payment_confirmed'
  | 'payment_rejected'
  | 'group_created'
  | 'settlement_created';

export interface Activity {
  id: string;
  group_id: string;
  group_name?: string;
  user_id: string;
  user_name: string;
  user_avatar?: string;
  type: ActivityType;
  title: string;
  detail: string;
  amount?: number;
  ref_id?: string;
  created_at: string;
  time_ago: string;
}

export interface BankingDeeplink {
  app_name: string;
  scheme: string;
  color: string;
  icon_name: string;
}

export interface PaymentDeeplinkRequest {
  bank_code: string;
  account_number: string;
  account_name?: string;
  amount: number;
  note?: string;
}

export interface PaymentDeeplinkResponse {
  amount: number;
  note: string;
  deeplinks: BankingDeeplink[];
  vietqr_url: string;
}

export interface VietQRRequest {
  bank_id: string;
  account_no: string;
  account_name?: string;
  amount: number;
  description?: string;
  template?: string;
}

export interface VietQRResponse {
  qr_url: string;
  bank_id: string;
  account_number: string;
  account_name: string;
  amount: number;
  description: string;
}

export interface BankInfo {
  id: string;
  name: string;
  code: string;
  short_name: string;
  logo: string;
  color: string;
}

export interface UserPaymentInfo {
  user_id: string;
  display_name: string;
  bank_accounts: BankAccount[];
  preferred_payment: string;
}

// ===== Statistics Types (Phase 5) =====

export interface BillSummary {
  id: string;
  title: string;
  amount: number;
  category: string;
  paid_by_name: string;
  created_at: string;
}

export interface MemberSpendStats {
  user_id: string;
  display_name: string;
  avatar_url: string;
  total_paid: number;
  total_owed: number;
  net_balance: number;
  bill_count: number;
  percentage: number;
}

export interface CategoryStat {
  category: string;
  total: number;
  count: number;
  percentage: number;
  icon: string;
  color: string;
}

export interface MonthlySpend {
  month: string;
  year: number;
  month_num: number;
  total: number;
  bill_count: number;
}

export interface GroupSpendInfo {
  group_id: string;
  group_name: string;
  total: number;
  bill_count: number;
}

export interface GroupStats {
  group_id: string;
  group_name: string;
  total_spent: number;
  total_bills: number;
  total_members: number;
  average_bill: number;
  largest_bill?: BillSummary;
  smallest_bill?: BillSummary;
  member_stats: MemberSpendStats[];
  category_stats: CategoryStat[];
  monthly_trend: MonthlySpend[];
  recent_bills: BillSummary[];
}

export interface UserOverallStats {
  total_groups: number;
  total_spent: number;
  total_owed: number;
  total_bills: number;
  top_groups: GroupSpendInfo[];
  category_stats: CategoryStat[];
  monthly_trend: MonthlySpend[];
}

export interface CategoryInfo {
  key: string;
  label: string;
  icon: string;
  color: string;
}
