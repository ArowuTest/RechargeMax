/**
 * Enterprise Metrics and Monitoring Infrastructure
 * Comprehensive metrics collection and monitoring system
 */

export interface IMetrics {
  incrementCounter(name: string, value?: number, tags?: Record<string, string>): void;
  recordGauge(name: string, value: number, tags?: Record<string, string>): void;
  recordHistogram(name: string, value: number, tags?: Record<string, string>): void;
  recordTimer(name: string, duration: number, tags?: Record<string, string>): void;
  startTimer(name: string, tags?: Record<string, string>): () => void;
}

export interface MetricEntry {
  name: string;
  type: 'counter' | 'gauge' | 'histogram' | 'timer';
  value: number;
  tags?: Record<string, string>;
  timestamp: number;
}

export interface MetricsTransport {
  send(metrics: MetricEntry[]): Promise<void>;
}

// Console Transport for development
export class ConsoleMetricsTransport implements MetricsTransport {
  async send(metrics: MetricEntry[]): Promise<void> {
    metrics.forEach(metric => {
      const tags = metric.tags ? Object.entries(metric.tags).map(([k, v]) => `${k}=${v}`).join(',') : '';
      console.log(`[METRIC] ${metric.name}:${metric.value} (${metric.type}) ${tags} @${new Date(metric.timestamp).toISOString()}`);
    });
  }
}

// Remote Metrics Transport (e.g., DataDog, New Relic)
export class RemoteMetricsTransport implements MetricsTransport {
  constructor(
    private readonly endpoint: string,
    private readonly apiKey: string
  ) {}

  async send(metrics: MetricEntry[]): Promise<void> {
    try {
      await fetch(this.endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiKey}`
        },
        body: JSON.stringify({ metrics })
      });
    } catch (error) {
      console.error('Failed to send metrics to remote endpoint:', error);
    }
  }
}

// Supabase Metrics Transport removed - use RemoteMetricsTransport with Go backend API instead
// export class SupabaseMetricsTransport implements MetricsTransport {
//   constructor(private readonly supabaseClient: any) {}
//   async send(metrics: MetricEntry[]): Promise<void> {
//     // Removed - use RemoteMetricsTransport
//   }
// }

// Main Metrics Implementation
export class Metrics implements IMetrics {
  private transports: MetricsTransport[] = [];
  private buffer: MetricEntry[] = [];
  private flushInterval: number = 10000; // 10 seconds
  private maxBufferSize: number = 100;
  private flushTimer?: NodeJS.Timeout;

  constructor(
    transports: MetricsTransport[] = [new ConsoleMetricsTransport()],
    flushInterval: number = 10000
  ) {
    this.transports = transports;
    this.flushInterval = flushInterval;
    this.startFlushTimer();
  }

  private startFlushTimer(): void {
    this.flushTimer = setInterval(() => {
      this.flush();
    }, this.flushInterval);
  }

  private async flush(): Promise<void> {
    if (this.buffer.length === 0) {
      return;
    }

    const metricsToSend = [...this.buffer];
    this.buffer = [];

    const promises = this.transports.map(transport => transport.send(metricsToSend));
    await Promise.allSettled(promises);
  }

  private addMetric(name: string, type: MetricEntry['type'], value: number, tags?: Record<string, string>): void {
    const metric: MetricEntry = {
      name,
      type,
      value,
      tags,
      timestamp: Date.now()
    };

    this.buffer.push(metric);

    // Flush immediately if buffer is full
    if (this.buffer.length >= this.maxBufferSize) {
      this.flush();
    }
  }

  incrementCounter(name: string, value: number = 1, tags?: Record<string, string>): void {
    this.addMetric(name, 'counter', value, tags);
  }

  recordGauge(name: string, value: number, tags?: Record<string, string>): void {
    this.addMetric(name, 'gauge', value, tags);
  }

  recordHistogram(name: string, value: number, tags?: Record<string, string>): void {
    this.addMetric(name, 'histogram', value, tags);
  }

  recordTimer(name: string, duration: number, tags?: Record<string, string>): void {
    this.addMetric(name, 'timer', duration, tags);
  }

  startTimer(name: string, tags?: Record<string, string>): () => void {
    const startTime = Date.now();
    
    return () => {
      const duration = Date.now() - startTime;
      this.recordTimer(name, duration, tags);
    };
  }

  async destroy(): Promise<void> {
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
    }
    await this.flush();
  }
}

// Metrics Factory
export class MetricsFactory {
  private static instance: Metrics;

  static createMetrics(config?: {
    transports?: MetricsTransport[];
    flushInterval?: number;
  }): Metrics {
    const transports = config?.transports || [new ConsoleMetricsTransport()];
    const flushInterval = config?.flushInterval || 10000;
    
    return new Metrics(transports, flushInterval);
  }

  static getDefaultMetrics(): Metrics {
    if (!this.instance) {
      this.instance = this.createMetrics({
        transports: [new ConsoleMetricsTransport()],
        flushInterval: 10000
      });
    }
    return this.instance;
  }
}

// Export default metrics instance
export const metrics = MetricsFactory.getDefaultMetrics();