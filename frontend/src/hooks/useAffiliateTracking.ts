import { useEffect } from 'react';
import { trackAffiliateClick } from '@/lib/api';

// Affiliate tracking hook
export const useAffiliateTracking = () => {
  useEffect(() => {
    const trackClick = async () => {
      // Get affiliate code from URL parameters
      const urlParams = new URLSearchParams(window.location.search);
      const affiliateCode = urlParams.get('ref');
      
      if (affiliateCode) {
        try {
          // Store affiliate code in session storage for later use
          sessionStorage.setItem('affiliate_code', affiliateCode);
          
          // Track the click
          await trackAffiliateClick();
          
        } catch (error) {
          console.error('Failed to track affiliate click:', error);
        }
      }
    };
    
    trackClick();
  }, []);
  
  // Get stored affiliate code for transactions
  const getAffiliateCode = () => {
    return sessionStorage.getItem('affiliate_code');
  };
  
  return { getAffiliateCode };
};

// Generate session ID
const generateSessionId = () => {
  let sessionId = sessionStorage.getItem('session_id');
  if (!sessionId) {
    sessionId = 'sess_' + Date.now() + '_' + Math.random().toString(36).substring(2);
    sessionStorage.setItem('session_id', sessionId);
  }
  return sessionId;
};

// Detect device type
const getDeviceType = () => {
  const userAgent = navigator.userAgent.toLowerCase();
  if (/mobile|android|iphone|ipad|phone/i.test(userAgent)) {
    return 'mobile';
  } else if (/tablet|ipad/i.test(userAgent)) {
    return 'tablet';
  } else {
    return 'desktop';
  }
};

// Enhanced affiliate tracking with more features
export const useAdvancedAffiliateTracking = () => {
  useEffect(() => {
    const trackAdvancedClick = async () => {
      const urlParams = new URLSearchParams(window.location.search);
      const affiliateCode = urlParams.get('ref');
      const campaign = urlParams.get('campaign') || 'direct';
      const source = urlParams.get('source') || 'unknown';
      
      if (affiliateCode) {
        try {
          // Store comprehensive tracking data
          const trackingData = {
            affiliate_code: affiliateCode,
            campaign,
            source,
            session_id: generateSessionId(),
            device_type: getDeviceType(),
            user_agent: navigator.userAgent,
            referrer: document.referrer,
            landing_page: window.location.pathname,
            timestamp: new Date().toISOString(),
            screen_resolution: `${screen.width}x${screen.height}`,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
            language: navigator.language
          };
          
          // Store in session storage
          sessionStorage.setItem('affiliate_tracking', JSON.stringify(trackingData));
          
          // Track the click with enhanced data
          await trackAffiliateClick();
          
        } catch (error) {
          console.error('Failed to track advanced affiliate click:', error);
        }
      }
    };
    
    trackAdvancedClick();
  }, []);
  
  // Get comprehensive tracking data
  const getTrackingData = () => {
    const stored = sessionStorage.getItem('affiliate_tracking');
    return stored ? JSON.parse(stored) : null;
  };
  
  // Get just the affiliate code
  const getAffiliateCode = () => {
    const trackingData = getTrackingData();
    return trackingData?.affiliate_code || null;
  };
  
  // Track conversion events
  const trackConversion = async (eventType: string, eventData: any = {}) => {
    const trackingData = getTrackingData();
    if (!trackingData) return;
    
    try {
      const conversionData = {
        ...trackingData,
        event_type: eventType,
        event_data: eventData,
        conversion_timestamp: new Date().toISOString()
      };
      
      // You would send this to your backend
      
      // Store conversion in session for potential retry
      const conversions = JSON.parse(sessionStorage.getItem('affiliate_conversions') || '[]');
      conversions.push(conversionData);
      sessionStorage.setItem('affiliate_conversions', JSON.stringify(conversions));
      
    } catch (error) {
      console.error('Failed to track conversion:', error);
    }
  };
  
  return { 
    getAffiliateCode, 
    getTrackingData, 
    trackConversion 
  };
};