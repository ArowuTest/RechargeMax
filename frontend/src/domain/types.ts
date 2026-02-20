/**
 * Enterprise Domain Types
 * Comprehensive type definitions following Domain-Driven Design principles
 */

// ============================================================================
// VALUE OBJECTS - Immutable domain primitives
// ============================================================================

export class PhoneNumber {
  private constructor(private readonly value: string) {
    if (!this.isValid(value)) {
      throw new Error(`Invalid Nigerian phone number: ${value}`);
    }
  }

  static create(value: string): PhoneNumber {
    const normalized = this.normalize(value);
    return new PhoneNumber(normalized);
  }

  private static normalize(value: string): string {
    const cleaned = value.replace(/[\s\-\+]/g, '');
    if (cleaned.startsWith('0')) {
      return '234' + cleaned.substring(1);
    }
    if (cleaned.startsWith('234')) {
      return cleaned;
    }
    throw new Error(`Cannot normalize phone number: ${value}`);
  }

  private isValid(value: string): boolean {
    const phoneRegex = /^234[789][01]\d{8}$/;
    return phoneRegex.test(value);
  }

  toString(): string {
    return this.value;
  }

  toDisplayFormat(): string {
    const local = this.value.substring(3);
    return `0${local.substring(0, 3)} ${local.substring(3, 6)} ${local.substring(6)}`;
  }

  equals(other: PhoneNumber): boolean {
    return this.value === other.value;
  }
}

export class Money {
  private constructor(
    private readonly amount: number,
    private readonly currency: Currency = Currency.NGN
  ) {
    if (amount < 0) {
      throw new Error('Money amount cannot be negative');
    }
    if (!Number.isFinite(amount)) {
      throw new Error('Money amount must be finite');
    }
  }

  static create(amount: number, currency: Currency = Currency.NGN): Money {
    return new Money(Math.round(amount * 100) / 100, currency); // Round to 2 decimal places
  }

  static zero(currency: Currency = Currency.NGN): Money {
    return new Money(0, currency);
  }

  getAmount(): number {
    return this.amount;
  }

  getCurrency(): Currency {
    return this.currency;
  }

  add(other: Money): Money {
    this.ensureSameCurrency(other);
    return new Money(this.amount + other.amount, this.currency);
  }

  subtract(other: Money): Money {
    this.ensureSameCurrency(other);
    return new Money(this.amount - other.amount, this.currency);
  }

  multiply(factor: number): Money {
    return new Money(this.amount * factor, this.currency);
  }

  isGreaterThan(other: Money): boolean {
    this.ensureSameCurrency(other);
    return this.amount > other.amount;
  }

  isLessThan(other: Money): boolean {
    this.ensureSameCurrency(other);
    return this.amount < other.amount;
  }

  equals(other: Money): boolean {
    return this.amount === other.amount && this.currency === other.currency;
  }

  private ensureSameCurrency(other: Money): void {
    if (this.currency !== other.currency) {
      throw new Error(`Currency mismatch: ${this.currency} vs ${other.currency}`);
    }
  }

  toString(): string {
    return `${this.currency}${this.amount.toLocaleString('en-NG', { minimumFractionDigits: 2 })}`;
  }
}

export class TransactionId {
  private constructor(private readonly value: string) {}

  static generate(): TransactionId {
    const timestamp = Date.now();
    const random = Math.random().toString(36).substring(2, 15);
    return new TransactionId(`TXN_${timestamp}_${random}`);
  }

  static fromString(value: string): TransactionId {
    if (!value || value.length < 10) {
      throw new Error('Invalid transaction ID');
    }
    return new TransactionId(value);
  }

  toString(): string {
    return this.value;
  }

  equals(other: TransactionId): boolean {
    return this.value === other.value;
  }
}

// ============================================================================
// ENUMERATIONS - Domain constants
// ============================================================================

export enum Currency {
  NGN = '₦',
  USD = '$',
  EUR = '€'
}

export enum NetworkProvider {
  MTN = 'MTN',
  AIRTEL = 'AIRTEL',
  GLO = 'GLO',
  NINE_MOBILE = '9MOBILE'
}

export enum RechargeType {
  AIRTIME = 'AIRTIME',
  DATA = 'DATA'
}

export enum TransactionStatus {
  PENDING = 'PENDING',
  PROCESSING = 'PROCESSING',
  SUCCESS = 'SUCCESS',
  FAILED = 'FAILED',
  CANCELLED = 'CANCELLED',
  REFUNDED = 'REFUNDED'
}

export enum PaymentMethod {
  CARD = 'CARD',
  BANK_TRANSFER = 'BANK_TRANSFER',
  USSD = 'USSD',
  MOBILE_MONEY = 'MOBILE_MONEY'
}

export enum UserRole {
  GUEST = 'GUEST',
  CUSTOMER = 'CUSTOMER',
  AFFILIATE = 'AFFILIATE',
  ADMIN = 'ADMIN',
  SUPER_ADMIN = 'SUPER_ADMIN'
}

export enum LoyaltyTier {
  BRONZE = 'BRONZE',
  SILVER = 'SILVER',
  GOLD = 'GOLD',
  PLATINUM = 'PLATINUM'
}

export enum AchievementType {
  RECHARGE_COUNT = 'RECHARGE_COUNT',
  RECHARGE_AMOUNT = 'RECHARGE_AMOUNT',
  STREAK = 'STREAK',
  REFERRAL = 'REFERRAL',
  TIER = 'TIER',
  SPIN_WIN = 'SPIN_WIN'
}

export enum NotificationType {
  RECHARGE_SUCCESS = 'RECHARGE_SUCCESS',
  RECHARGE_FAILED = 'RECHARGE_FAILED',
  PRIZE_WON = 'PRIZE_WON',
  DRAW_REMINDER = 'DRAW_REMINDER',
  ACHIEVEMENT_UNLOCKED = 'ACHIEVEMENT_UNLOCKED',
  TIER_UPGRADED = 'TIER_UPGRADED',
  AFFILIATE_COMMISSION = 'AFFILIATE_COMMISSION',
  SYSTEM_MAINTENANCE = 'SYSTEM_MAINTENANCE'
}

// ============================================================================
// DOMAIN ENTITIES - Core business objects
// ============================================================================

export interface User {
  readonly id: string;
  readonly phoneNumber: PhoneNumber;
  readonly email?: string;
  readonly fullName?: string;
  readonly role: UserRole;
  readonly loyaltyTier: LoyaltyTier;
  readonly totalRechargeAmount: Money;
  readonly totalPoints: number;
  readonly currentStreak: number;
  readonly longestStreak: number;
  readonly lastRechargeDate?: Date;
  readonly createdAt: Date;
  readonly updatedAt: Date;
  readonly isActive: boolean;
}

export interface Transaction {
  readonly id: TransactionId;
  readonly userId?: string;
  readonly phoneNumber: PhoneNumber;
  readonly networkProvider: NetworkProvider;
  readonly rechargeType: RechargeType;
  readonly amount: Money;
  readonly paymentMethod: PaymentMethod;
  readonly status: TransactionStatus;
  readonly pointsEarned: number;
  readonly drawEntries: number;
  readonly loyaltyMultiplier: number;
  readonly affiliateUserId?: string;
  readonly affiliateCommission: Money;
  readonly telecomReference?: string;
  readonly paymentReference?: string;
  readonly gatewayReference?: string;
  readonly createdAt: Date;
  readonly completedAt?: Date;
  readonly metadata: Record<string, unknown>;
}

export interface Draw {
  readonly id: string;
  readonly name: string;
  readonly type: DrawType;
  readonly prizePool: Money;
  readonly numberOfWinners: number;
  readonly startTime: Date;
  readonly endTime: Date;
  readonly drawTime?: Date;
  readonly status: DrawStatus;
  readonly totalEntries: number;
  readonly winners?: DrawWinner[];
  readonly externalDrawId?: string;
}

export interface DrawEntry {
  readonly id: string;
  readonly drawId: string;
  readonly userId?: string;
  readonly phoneNumber: PhoneNumber;
  readonly entriesCount: number;
  readonly sourceType: EntrySourceType;
  readonly sourceTransactionId?: TransactionId;
  readonly createdAt: Date;
}

export interface Affiliate {
  readonly id: string;
  readonly userId: string;
  readonly uniqueCode: string;
  readonly totalClicks: number;
  readonly totalReferrals: number;
  readonly totalCommission: Money;
  readonly pendingCommission: Money;
  readonly paidCommission: Money;
  readonly isActive: boolean;
  readonly createdAt: Date;
  readonly updatedAt: Date;
}

export interface SpinResult {
  readonly id: string;
  readonly userId?: string;
  readonly phoneNumber: PhoneNumber;
  readonly transactionId: TransactionId;
  readonly prizeName: string;
  readonly prizeType: PrizeType;
  readonly prizeValue: number;
  readonly status: PrizeStatus;
  readonly createdAt: Date;
  readonly fulfilledAt?: Date;
}

export interface Achievement {
  readonly id: string;
  readonly key: string;
  readonly name: string;
  readonly description: string;
  readonly type: AchievementType;
  readonly criteriaValue: number;
  readonly pointsReward: number;
  readonly icon: string;
  readonly isActive: boolean;
}

export interface UserAchievement {
  readonly id: string;
  readonly userId: string;
  readonly achievementId: string;
  readonly currentProgress: number;
  readonly isCompleted: boolean;
  readonly pointsAwarded: number;
  readonly unlockedAt?: Date;
}

// ============================================================================
// DOMAIN EVENTS - Business events for event-driven architecture
// ============================================================================

export interface DomainEvent {
  readonly eventId: string;
  readonly eventType: string;
  readonly aggregateId: string;
  readonly aggregateType: string;
  readonly eventData: Record<string, unknown>;
  readonly occurredAt: Date;
  readonly version: number;
}

export interface RechargeCompletedEvent extends DomainEvent {
  readonly eventType: 'RechargeCompleted';
  readonly eventData: {
    readonly transactionId: string;
    readonly userId?: string;
    readonly phoneNumber: string;
    readonly amount: number;
    readonly networkProvider: NetworkProvider;
    readonly pointsEarned: number;
    readonly drawEntries: number;
  };
}

export interface PrizeWonEvent extends DomainEvent {
  readonly eventType: 'PrizeWon';
  readonly eventData: {
    readonly userId?: string;
    readonly phoneNumber: string;
    readonly prizeName: string;
    readonly prizeValue: number;
    readonly prizeType: PrizeType;
  };
}

export interface AchievementUnlockedEvent extends DomainEvent {
  readonly eventType: 'AchievementUnlocked';
  readonly eventData: {
    readonly userId: string;
    readonly achievementId: string;
    readonly achievementName: string;
    readonly pointsAwarded: number;
  };
}

// ============================================================================
// ADDITIONAL ENUMS AND TYPES
// ============================================================================

export enum DrawType {
  DAILY = 'DAILY',
  WEEKLY = 'WEEKLY',
  MONTHLY = 'MONTHLY',
  SPECIAL = 'SPECIAL'
}

export enum DrawStatus {
  UPCOMING = 'UPCOMING',
  ACTIVE = 'ACTIVE',
  DRAWING = 'DRAWING',
  COMPLETED = 'COMPLETED',
  CANCELLED = 'CANCELLED'
}

export enum EntrySourceType {
  RECHARGE = 'RECHARGE',
  SUBSCRIPTION = 'SUBSCRIPTION',
  BONUS = 'BONUS',
  REFERRAL = 'REFERRAL'
}

export enum PrizeType {
  AIRTIME = 'AIRTIME',
  DATA = 'DATA',
  CASH = 'CASH',
  POINTS = 'POINTS'
}

export enum PrizeStatus {
  PENDING = 'PENDING',
  FULFILLED = 'FULFILLED',
  FAILED = 'FAILED',
  EXPIRED = 'EXPIRED'
}

export interface DrawWinner {
  readonly userId?: string;
  readonly phoneNumber: PhoneNumber;
  readonly prizeAmount: Money;
  readonly position: number;
}

// ============================================================================
// RESULT PATTERN - For error handling without exceptions
// ============================================================================

export type Result<T, E = Error> = Success<T> | Failure<E>;

export class Success<T> {
  constructor(public readonly value: T) {}
  
  isSuccess(): this is Success<T> {
    return true;
  }
  
  isFailure(): this is Failure<never> {
    return false;
  }
}

export class Failure<E> {
  constructor(public readonly error: E) {}
  
  isSuccess(): this is Success<never> {
    return false;
  }
  
  isFailure(): this is Failure<E> {
    return true;
  }
}

export const success = <T>(value: T): Success<T> => new Success(value);
export const failure = <E>(error: E): Failure<E> => new Failure(error);

// ============================================================================
// DOMAIN ERRORS - Specific business errors
// ============================================================================

export abstract class DomainError extends Error {
  abstract readonly code: string;
  abstract readonly statusCode: number;
}

export class InvalidPhoneNumberError extends DomainError {
  readonly code = 'INVALID_PHONE_NUMBER';
  readonly statusCode = 400;
  
  constructor(phoneNumber: string) {
    super(`Invalid phone number: ${phoneNumber}`);
  }
}

export class InsufficientFundsError extends DomainError {
  readonly code = 'INSUFFICIENT_FUNDS';
  readonly statusCode = 400;
  
  constructor(available: Money, required: Money) {
    super(`Insufficient funds: ${available} available, ${required} required`);
  }
}

export class TransactionNotFoundError extends DomainError {
  readonly code = 'TRANSACTION_NOT_FOUND';
  readonly statusCode = 404;
  
  constructor(transactionId: TransactionId) {
    super(`Transaction not found: ${transactionId}`);
  }
}

export class RateLimitExceededError extends DomainError {
  readonly code = 'RATE_LIMIT_EXCEEDED';
  readonly statusCode = 429;
  
  constructor(limit: number, window: string) {
    super(`Rate limit exceeded: ${limit} requests per ${window}`);
  }
}

export class NetworkProviderUnavailableError extends DomainError {
  readonly code = 'NETWORK_PROVIDER_UNAVAILABLE';
  readonly statusCode = 503;
  
  constructor(provider: NetworkProvider) {
    super(`Network provider unavailable: ${provider}`);
  }
}