/**
 * Enterprise Logging Infrastructure
 * Comprehensive logging system with multiple transports and structured logging
 */

export interface ILogger {
  debug(message: string, meta?: Record<string, unknown>): void;
  info(message: string, meta?: Record<string, unknown>): void;
  warn(message: string, meta?: Record<string, unknown>): void;
  error(message: string, meta?: Record<string, unknown>): void;
  fatal(message: string, meta?: Record<string, unknown>): void;
}

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
  FATAL = 4
}

export interface LogEntry {
  timestamp: string;
  level: LogLevel;
  message: string;
  meta?: Record<string, unknown>;
  correlationId?: string;
  userId?: string;
  sessionId?: string;
  requestId?: string;
}

export interface LogTransport {
  log(entry: LogEntry): Promise<void>;
}

// Console Transport for development
export class ConsoleTransport implements LogTransport {
  async log(entry: LogEntry): Promise<void> {
    const levelName = LogLevel[entry.level];
    const timestamp = entry.timestamp;
    const message = entry.message;
    const meta = entry.meta ? JSON.stringify(entry.meta, null, 2) : '';
    
    const logMessage = `[${timestamp}] ${levelName}: ${message}${meta ? '\n' + meta : ''}`;
    
    switch (entry.level) {
      case LogLevel.DEBUG:
        console.debug(logMessage);
        break;
      case LogLevel.INFO:
        console.info(logMessage);
        break;
      case LogLevel.WARN:
        console.warn(logMessage);
        break;
      case LogLevel.ERROR:
      case LogLevel.FATAL:
        console.error(logMessage);
        break;
    }
  }
}

// Remote Transport for production logging
export class RemoteTransport implements LogTransport {
  constructor(
    private readonly endpoint: string,
    private readonly apiKey: string
  ) {}

  async log(entry: LogEntry): Promise<void> {
    try {
      await fetch(this.endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiKey}`
        },
        body: JSON.stringify(entry)
      });
    } catch (error) {
      // Fallback to console if remote logging fails
      console.error('Failed to send log to remote endpoint:', error);
      console.error('Original log entry:', entry);
    }
  }
}

// Supabase Transport removed - use RemoteTransport with Go backend API instead
// export class SupabaseTransport implements LogTransport {
//   constructor(private readonly supabaseClient: any) {}
//   async log(entry: LogEntry): Promise<void> {
//     // Removed - use RemoteTransport
//   }
// }

// Main Logger Implementation
export class Logger implements ILogger {
  private transports: LogTransport[] = [];
  private minLevel: LogLevel = LogLevel.INFO;
  private context: Record<string, unknown> = {};

  constructor(
    transports: LogTransport[] = [new ConsoleTransport()],
    minLevel: LogLevel = LogLevel.INFO
  ) {
    this.transports = transports;
    this.minLevel = minLevel;
  }

  setContext(context: Record<string, unknown>): void {
    this.context = { ...this.context, ...context };
  }

  clearContext(): void {
    this.context = {};
  }

  private async log(level: LogLevel, message: string, meta?: Record<string, unknown>): Promise<void> {
    if (level < this.minLevel) {
      return;
    }

    const entry: LogEntry = {
      timestamp: new Date().toISOString(),
      level,
      message,
      meta: { ...this.context, ...meta },
      correlationId: this.generateCorrelationId(),
      userId: this.context.userId as string,
      sessionId: this.context.sessionId as string,
      requestId: this.context.requestId as string
    };

    // Send to all transports in parallel
    const promises = this.transports.map(transport => transport.log(entry));
    await Promise.allSettled(promises);
  }

  debug(message: string, meta?: Record<string, unknown>): void {
    this.log(LogLevel.DEBUG, message, meta);
  }

  info(message: string, meta?: Record<string, unknown>): void {
    this.log(LogLevel.INFO, message, meta);
  }

  warn(message: string, meta?: Record<string, unknown>): void {
    this.log(LogLevel.WARN, message, meta);
  }

  error(message: string, meta?: Record<string, unknown>): void {
    this.log(LogLevel.ERROR, message, meta);
  }

  fatal(message: string, meta?: Record<string, unknown>): void {
    this.log(LogLevel.FATAL, message, meta);
  }

  private generateCorrelationId(): string {
    return `corr_${Date.now()}_${Math.random().toString(36).substring(2)}`;
  }
}

// Logger Factory
export class LoggerFactory {
  private static instance: Logger;

  static createLogger(config?: {
    transports?: LogTransport[];
    minLevel?: LogLevel;
    context?: Record<string, unknown>;
  }): Logger {
    const transports = config?.transports || [new ConsoleTransport()];
    const minLevel = config?.minLevel || LogLevel.INFO;
    
    const logger = new Logger(transports, minLevel);
    
    if (config?.context) {
      logger.setContext(config.context);
    }

    return logger;
  }

  static getDefaultLogger(): Logger {
    if (!this.instance) {
      this.instance = this.createLogger({
        transports: [new ConsoleTransport()],
        minLevel: process.env.NODE_ENV === 'development' ? LogLevel.DEBUG : LogLevel.INFO
      });
    }
    return this.instance;
  }
}

// Export default logger instance
export const logger = LoggerFactory.getDefaultLogger();