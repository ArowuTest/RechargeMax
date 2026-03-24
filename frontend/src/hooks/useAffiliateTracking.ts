import { useEffect } from 'react';
import { apiClient } from '@/lib/api-client';

const STORAGE_KEY = 'rmx_aff';
const TTL_MS = 30 * 24 * 60 * 60 * 1000; // 30 days

interface AffiliateTrackingData {
  affiliate_code: string;
  stored_at: number;
}

/**
 * Reads the ?ref=AFFxxxx param from the URL, persists it to localStorage
 * (30-day TTL), and fires a click-record to the backend.
 *
 * getAffiliateCode() — returns the stored code for use in API payloads, or
 * null if none was ever captured or the TTL has expired.
 */
export const useAffiliateTracking = () => {
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const refCode = urlParams.get('ref');

    if (refCode && refCode.startsWith('AFF')) {
      // Persist with timestamp
      const payload: AffiliateTrackingData = {
        affiliate_code: refCode,
        stored_at: Date.now(),
      };
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(payload));
        // Also keep in sessionStorage for same-session fast reads
        sessionStorage.setItem('affiliate_code', refCode);
      } catch { /* storage unavailable */ }

      // Record click on backend (best-effort, non-blocking)
      apiClient
        .post('/affiliate/track-click', {
          affiliate_code: refCode,
          source: document.referrer || 'direct',
        })
        .catch(() => { /* silently ignore */ });
    }
  }, []);

  /**
   * Returns the captured affiliate code if still within TTL, otherwise null.
   * Checks sessionStorage first (fastest), then localStorage.
   */
  const getAffiliateCode = (): string | null => {
    // Fast path — same browser session
    const session = sessionStorage.getItem('affiliate_code');
    if (session) return session;

    // Persistent path — check localStorage TTL
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return null;
      const data: AffiliateTrackingData = JSON.parse(raw);
      if (Date.now() - data.stored_at > TTL_MS) {
        localStorage.removeItem(STORAGE_KEY);
        return null;
      }
      return data.affiliate_code;
    } catch {
      return null;
    }
  };

  /** Clears attribution data (e.g. after a successful payment). */
  const clearAffiliateCode = () => {
    sessionStorage.removeItem('affiliate_code');
    localStorage.removeItem(STORAGE_KEY);
  };

  return { getAffiliateCode, clearAffiliateCode };
};

// ─────────────────────────────────────────────────────────────────────────────
// Advanced variant (kept for backward compatibility)
// ─────────────────────────────────────────────────────────────────────────────
export const useAdvancedAffiliateTracking = () => {
  const { getAffiliateCode, clearAffiliateCode } = useAffiliateTracking();

  const trackConversion = async (eventType: string, eventData: Record<string, unknown> = {}) => {
    const affiliateCode = getAffiliateCode();
    if (!affiliateCode) return;

    try {
      await apiClient.post('/affiliate/track-conversion', {
        affiliate_code: affiliateCode,
        event_type: eventType,
        event_data: eventData,
        conversion_timestamp: new Date().toISOString(),
      });
    } catch {
      // Queue for retry
      try {
        const retries = JSON.parse(sessionStorage.getItem('aff_retry') || '[]');
        retries.push({ affiliateCode, eventType, eventData, ts: Date.now() });
        sessionStorage.setItem('aff_retry', JSON.stringify(retries.slice(-10)));
      } catch { /* storage unavailable */ }
    }
  };

  return { getAffiliateCode, clearAffiliateCode, trackConversion };
};
