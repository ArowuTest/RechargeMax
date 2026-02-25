import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: 'NGN',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

export function formatDate(date: string | Date): string {
  return new Intl.DateTimeFormat('en-NG', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(date));
}

export function generateReferralCode(): string {
  return Math.random().toString(36).substring(2, 8).toUpperCase();
}

export function calculateCommission(amount: number, rate: number): number {
  return Math.floor(amount * rate);
}

export function validatePhoneNumber(phone: string): boolean {
  const phoneRegex = /^(\+234|234|0)?[789][01]\d{8}$/;
  return phoneRegex.test(phone);
}

export function formatPhoneNumber(phone: string): string {
  // Remove all non-digits
  const cleaned = phone.replace(/\D/g, '');
  
  // Handle different formats
  if (cleaned.startsWith('234')) {
    return '+' + cleaned;
  } else if (cleaned.startsWith('0')) {
    return '+234' + cleaned.substring(1);
  } else if (cleaned.length === 10) {
    return '+234' + cleaned;
  }
  
  return phone;
}

export function generateTransactionId(): string {
  const timestamp = Date.now().toString();
  const random = Math.random().toString(36).substring(2, 8);
  return `TXN_${timestamp}_${random}`.toUpperCase();
}

export function getSpinEligibility(rechargeAmount: number): boolean {
  return rechargeAmount >= 1000;
}

export function calculateSpinCount(rechargeAmount: number): number {
  if (rechargeAmount < 1000) return 0;
  return Math.floor(rechargeAmount / 1000);
}

export function getDrawEligibility(rechargeAmount: number): boolean {
  return rechargeAmount >= 500;
}

export function calculateDrawEntries(rechargeAmount: number): number {
  if (rechargeAmount < 500) return 0;
  return Math.floor(rechargeAmount / 500);
}

export function validateNigerianPhone(phone: string): boolean {
  // Remove all non-digits
  const cleaned = phone.replace(/\D/g, '');
  
  // Check if it's a valid Nigerian phone number
  // Should be 11 digits starting with 0, or 10 digits, or 13 digits starting with 234
  if (cleaned.startsWith('234') && cleaned.length === 13) {
    return /^234[789][01]\d{8}$/.test(cleaned);
  } else if (cleaned.startsWith('0') && cleaned.length === 11) {
    return /^0[789][01]\d{8}$/.test(cleaned);
  } else if (cleaned.length === 10) {
    return /^[789][01]\d{8}$/.test(cleaned);
  }
  
  return false;
}

export function displayPhoneNumber(phone: string): string {
  // Remove all non-digits
  const cleaned = phone.replace(/\D/g, '');
  
  let formatted = cleaned;
  
  // Normalize to 234 format
  if (cleaned.startsWith('0') && cleaned.length === 11) {
    formatted = '234' + cleaned.substring(1);
  } else if (cleaned.length === 10) {
    formatted = '234' + cleaned;
  } else if (cleaned.startsWith('234')) {
    formatted = cleaned;
  }
  
  // Format as +234 XXX XXX XXXX
  if (formatted.startsWith('234') && formatted.length === 13) {
    return `+234 ${formatted.substring(3, 6)} ${formatted.substring(6, 9)} ${formatted.substring(9)}`;
  }
  
  return phone;
}

export function getNetworkColor(network: string): string {
  const networkColors: Record<string, string> = {
    'MTN': 'bg-yellow-500',
    'GLO': 'bg-green-500',
    'AIRTEL': 'bg-red-500',
    '9MOBILE': 'bg-emerald-600',
  };
  
  return networkColors[network.toUpperCase()] || 'bg-gray-500';
}

export function formatRelativeTime(date: string | Date): string {
  const now = new Date();
  const then = typeof date === 'string' ? new Date(date) : date;
  const diffMs = now.getTime() - then.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) {
    return 'just now';
  } else if (diffMins < 60) {
    return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`;
  } else if (diffHours < 24) {
    return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
  } else if (diffDays < 7) {
    return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
  } else {
    return formatDate(then);
  }
}
