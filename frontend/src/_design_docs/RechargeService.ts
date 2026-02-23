/**
 * Enterprise Application Services Layer
 * Implements Clean Architecture principles with comprehensive business logic
 */

import { 
  Result, 
  success, 
  failure, 
  Transaction, 
  TransactionId, 
  PhoneNumber, 
  Money, 
  NetworkProvider, 
  RechargeType, 
  PaymentMethod, 
  TransactionStatus,
  User,
  DomainEvent,
  RechargeCompletedEvent,
  InvalidPhoneNumberError,
  RateLimitExceededError,
  NetworkProviderUnavailableError
} from '@/domain/types';

import { ITransactionRepository } from '@/domain/repositories/ITransactionRepository';
import { IUserRepository } from '@/domain/repositories/IUserRepository';
import { IPaymentGateway } from '@/domain/services/IPaymentGateway';
import { ITelecomProvider } from '@/domain/services/ITelecomProvider';
import { IEventPublisher } from '@/domain/services/IEventPublisher';
import { IRateLimiter } from '@/domain/services/IRateLimiter';
import { ILogger } from '@/infrastructure/logging/ILogger';
import { IMetrics } from '@/infrastructure/monitoring/IMetrics';

// ============================================================================
// RECHARGE SERVICE - Core business logic for mobile recharges
// ============================================================================

export interface RechargeRequest {
  readonly phoneNumber: string;
  readonly networkProvider: NetworkProvider;
  readonly rechargeType: RechargeType;
  readonly amount: number;
  readonly paymentMethod: PaymentMethod;
  readonly paymentReference?: string;
  readonly userId?: string;
  readonly affiliateCode?: string;
  readonly dataBundle?: string;
  readonly metadata?: Record<string, unknown>;
}

export interface RechargeResult {
  readonly transactionId: TransactionId;
  readonly status: TransactionStatus;
  readonly pointsEarned: number;
  readonly drawEntries: number;
  readonly spinEligible: boolean;
  readonly telecomReference?: string;
  readonly estimatedDeliveryTime: Date;
}

export class RechargeService {
  constructor(
    private readonly transactionRepository: ITransactionRepository,
    private readonly userRepository: IUserRepository,
    private readonly paymentGateway: IPaymentGateway,
    private readonly telecomProvider: ITelecomProvider,
    private readonly eventPublisher: IEventPublisher,
    private readonly rateLimiter: IRateLimiter,
    private readonly logger: ILogger,
    private readonly metrics: IMetrics
  ) {}

  async processRecharge(request: RechargeRequest): Promise<Result<RechargeResult, Error>> {
    const startTime = Date.now();
    const correlationId = this.generateCorrelationId();
    
    this.logger.info('Processing recharge request', {
      correlationId,
      phoneNumber: request.phoneNumber,
      amount: request.amount,
      networkProvider: request.networkProvider
    });

    try {
      // 1. Validate and normalize input
      const validationResult = await this.validateRequest(request);
      if (validationResult.isFailure()) {
        this.metrics.incrementCounter('recharge.validation.failed');
        return failure(validationResult.error);
      }

      const { phoneNumber, amount, user } = validationResult.value;

      // 2. Check rate limits
      const rateLimitResult = await this.checkRateLimit(phoneNumber);
      if (rateLimitResult.isFailure()) {
        this.metrics.incrementCounter('recharge.rate_limit.exceeded');
        return failure(rateLimitResult.error);
      }

      // 3. Create transaction record
      const transactionId = TransactionId.generate();
      const transaction = await this.createTransaction({
        ...request,
        transactionId,
        phoneNumber,
        amount,
        user
      });

      // 4. Process payment
      const paymentResult = await this.processPayment(transaction, request.paymentReference);
      if (paymentResult.isFailure()) {
        await this.updateTransactionStatus(transactionId, TransactionStatus.FAILED);
        this.metrics.incrementCounter('recharge.payment.failed');
        return failure(paymentResult.error);
      }

      // 5. Execute telecom recharge
      const telecomResult = await this.executeTelecomRecharge(transaction);
      if (telecomResult.isFailure()) {
        await this.handleTelecomFailure(transaction, telecomResult.error);
        this.metrics.incrementCounter('recharge.telecom.failed');
        return failure(telecomResult.error);
      }

      // 6. Complete transaction and calculate rewards
      const completionResult = await this.completeTransaction(
        transaction, 
        telecomResult.value.reference
      );

      // 7. Publish domain events
      await this.publishRechargeCompletedEvent(transaction, completionResult);

      // 8. Record metrics
      const duration = Date.now() - startTime;
      this.metrics.recordHistogram('recharge.duration', duration);
      this.metrics.incrementCounter('recharge.success');

      this.logger.info('Recharge completed successfully', {
        correlationId,
        transactionId: transactionId.toString(),
        duration
      });

      return success({
        transactionId,
        status: TransactionStatus.SUCCESS,
        pointsEarned: completionResult.pointsEarned,
        drawEntries: completionResult.drawEntries,
        spinEligible: completionResult.spinEligible,
        telecomReference: telecomResult.value.reference,
        estimatedDeliveryTime: new Date(Date.now() + 30000) // 30 seconds
      });

    } catch (error) {
      const duration = Date.now() - startTime;
      this.logger.error('Recharge processing failed', {
        correlationId,
        error: error instanceof Error ? error.message : String(error),
        duration
      });
      
      this.metrics.incrementCounter('recharge.error');
      return failure(error instanceof Error ? error : new Error(String(error)));
    }
  }

  private async validateRequest(request: RechargeRequest): Promise<Result<{
    phoneNumber: PhoneNumber;
    amount: Money;
    user?: User;
  }, Error>> {
    try {
      // Validate phone number
      const phoneNumber = PhoneNumber.create(request.phoneNumber);
      
      // Validate amount
      if (request.amount < 50 || request.amount > 50000) {
        return failure(new Error('Amount must be between ₦50 and ₦50,000'));
      }
      const amount = Money.create(request.amount);

      // Get user if authenticated
      let user: User | undefined;
      if (request.userId) {
        const userResult = await this.userRepository.findById(request.userId);
        if (userResult.isSuccess()) {
          user = userResult.value;
        }
      }

      return success({ phoneNumber, amount, user });
    } catch (error) {
      if (error instanceof Error && error.message.includes('Invalid Nigerian phone number')) {
        return failure(new InvalidPhoneNumberError(request.phoneNumber));
      }
      return failure(error instanceof Error ? error : new Error(String(error)));
    }
  }

  private async checkRateLimit(phoneNumber: PhoneNumber): Promise<Result<void, RateLimitExceededError>> {
    const key = `recharge:${phoneNumber.toString()}`;
    const limit = 10; // 10 recharges per hour
    const window = 3600; // 1 hour in seconds

    const isAllowed = await this.rateLimiter.isAllowed(key, limit, window);
    if (!isAllowed) {
      return failure(new RateLimitExceededError(limit, '1 hour'));
    }

    return success(undefined);
  }

  private async createTransaction(params: {
    transactionId: TransactionId;
    phoneNumber: PhoneNumber;
    amount: Money;
    networkProvider: NetworkProvider;
    rechargeType: RechargeType;
    paymentMethod: PaymentMethod;
    user?: User;
    affiliateCode?: string;
    metadata?: Record<string, unknown>;
  }): Promise<Transaction> {
    const { user, ...rest } = params;
    
    // Calculate points and draw entries
    const pointsEarned = this.calculatePoints(params.amount, user);
    const drawEntries = this.calculateDrawEntries(pointsEarned);
    const loyaltyMultiplier = user ? this.getLoyaltyMultiplier(user.loyaltyTier) : 1.0;

    const transaction: Transaction = {
      id: params.transactionId,
      userId: user?.id,
      phoneNumber: params.phoneNumber,
      networkProvider: params.networkProvider,
      rechargeType: params.rechargeType,
      amount: params.amount,
      paymentMethod: params.paymentMethod,
      status: TransactionStatus.PENDING,
      pointsEarned,
      drawEntries,
      loyaltyMultiplier,
      affiliateCommission: Money.zero(),
      createdAt: new Date(),
      metadata: params.metadata || {}
    };

    await this.transactionRepository.save(transaction);
    return transaction;
  }

  private async processPayment(
    transaction: Transaction, 
    paymentReference?: string
  ): Promise<Result<{ reference: string; gatewayReference: string }, Error>> {
    try {
      const result = await this.paymentGateway.verifyPayment({
        amount: transaction.amount,
        reference: paymentReference || transaction.id.toString(),
        method: transaction.paymentMethod
      });

      if (result.isSuccess()) {
        await this.updateTransactionStatus(transaction.id, TransactionStatus.PROCESSING);
        return success(result.value);
      }

      return failure(result.error);
    } catch (error) {
      return failure(error instanceof Error ? error : new Error(String(error)));
    }
  }

  private async executeTelecomRecharge(
    transaction: Transaction
  ): Promise<Result<{ reference: string; balance?: number }, Error>> {
    try {
      const result = await this.telecomProvider.processRecharge({
        phoneNumber: transaction.phoneNumber,
        networkProvider: transaction.networkProvider,
        rechargeType: transaction.rechargeType,
        amount: transaction.amount
      });

      if (result.isSuccess()) {
        return success(result.value);
      }

      return failure(new NetworkProviderUnavailableError(transaction.networkProvider));
    } catch (error) {
      return failure(error instanceof Error ? error : new Error(String(error)));
    }
  }

  private async completeTransaction(
    transaction: Transaction,
    telecomReference: string
  ): Promise<{
    pointsEarned: number;
    drawEntries: number;
    spinEligible: boolean;
  }> {
    // Update transaction status
    await this.updateTransactionStatus(transaction.id, TransactionStatus.SUCCESS, {
      telecomReference,
      completedAt: new Date()
    });

    // Update user statistics if authenticated
    if (transaction.userId) {
      await this.updateUserStatistics(transaction.userId, transaction.amount, transaction.pointsEarned);
    }

    // Process affiliate commission if applicable
    if (transaction.affiliateUserId) {
      await this.processAffiliateCommission(transaction);
    }

    // Add entries to active draws
    await this.addToActiveDraws(transaction);

    return {
      pointsEarned: transaction.pointsEarned,
      drawEntries: transaction.drawEntries,
      spinEligible: transaction.amount.getAmount() >= 1000
    };
  }

  private async publishRechargeCompletedEvent(
    transaction: Transaction,
    result: { pointsEarned: number; drawEntries: number }
  ): Promise<void> {
    const event: RechargeCompletedEvent = {
      eventId: this.generateEventId(),
      eventType: 'RechargeCompleted',
      aggregateId: transaction.id.toString(),
      aggregateType: 'Transaction',
      eventData: {
        transactionId: transaction.id.toString(),
        userId: transaction.userId,
        phoneNumber: transaction.phoneNumber.toString(),
        amount: transaction.amount.getAmount(),
        networkProvider: transaction.networkProvider,
        pointsEarned: result.pointsEarned,
        drawEntries: result.drawEntries
      },
      occurredAt: new Date(),
      version: 1
    };

    await this.eventPublisher.publish(event);
  }

  // Helper methods
  private calculatePoints(amount: Money, user?: User): number {
    const basePoints = amount.getAmount(); // 1:1 ratio
    const multiplier = user ? this.getLoyaltyMultiplier(user.loyaltyTier) : 1.0;
    return Math.floor(basePoints * multiplier);
  }

  private calculateDrawEntries(points: number): number {
    return Math.floor(points / 200); // Every 200 points = 1 entry
  }

  private getLoyaltyMultiplier(tier: string): number {
    const multipliers = {
      'BRONZE': 1.0,
      'SILVER': 1.2,
      'GOLD': 1.5,
      'PLATINUM': 2.0
    };
    return multipliers[tier as keyof typeof multipliers] || 1.0;
  }

  private generateCorrelationId(): string {
    return `corr_${Date.now()}_${Math.random().toString(36).substring(2)}`;
  }

  private generateEventId(): string {
    return `evt_${Date.now()}_${Math.random().toString(36).substring(2)}`;
  }

  // Placeholder methods - to be implemented
  private async updateTransactionStatus(
    transactionId: TransactionId, 
    status: TransactionStatus, 
    updates?: Record<string, unknown>
  ): Promise<void> {
    // Implementation would update the transaction in the repository
  }

  private async handleTelecomFailure(transaction: Transaction, error: Error): Promise<void> {
    // Implementation would handle telecom failures, potentially scheduling retries
  }

  private async updateUserStatistics(userId: string, amount: Money, points: number): Promise<void> {
    // Implementation would update user's total recharge amount, points, etc.
  }

  private async processAffiliateCommission(transaction: Transaction): Promise<void> {
    // Implementation would calculate and record affiliate commission
  }

  private async addToActiveDraws(transaction: Transaction): Promise<void> {
    // Implementation would add entries to active draws
  }
}

// ============================================================================
// DRAW SERVICE - Manages prize draws and entries
// ============================================================================

export class DrawService {
  constructor(
    private readonly logger: ILogger,
    private readonly metrics: IMetrics
  ) {}

  // Implementation for draw management
}

// ============================================================================
// GAMIFICATION SERVICE - Handles achievements, streaks, and rewards
// ============================================================================

export class GamificationService {
  constructor(
    private readonly logger: ILogger,
    private readonly metrics: IMetrics
  ) {}

  // Implementation for gamification features
}

// ============================================================================
// NOTIFICATION SERVICE - Manages user notifications
// ============================================================================

export class NotificationService {
  constructor(
    private readonly logger: ILogger,
    private readonly metrics: IMetrics
  ) {}

  // Implementation for notifications
}