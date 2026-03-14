import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { subscriptionApi } from '../../lib/api-client';

interface SubscriptionConfig {
  daily_price: number; // in kobo
  daily_spins: number;
  auto_renewal: boolean;
}

const DailySubscriptionPage: React.FC = () => {
  const navigate = useNavigate();
  const [phoneNumber, setPhoneNumber] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [config, setConfig] = useState<SubscriptionConfig | null>(null);
  const [configLoading, setConfigLoading] = useState(true);

  // Fetch subscription config on mount
  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const result = await subscriptionApi.getConfig();
        if (result?.success && result?.data) {
          setConfig(result.data as SubscriptionConfig);
        }
      } catch (err) {
        console.error('Failed to fetch subscription config:', err);
      } finally {
        setConfigLoading(false);
      }
    };
    fetchConfig();
  }, []);

  // Compute display price in naira
  const dailyPriceNaira = config?.daily_price ? config.daily_price / 100 : 20;
  const dailySpins = config?.daily_spins || 3;

  const handleSubscribe = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      const result = await subscriptionApi.subscribe(phoneNumber || undefined);

      if (result.success) {
        if (result.data?.payment_url) {
          window.location.href = result.data.payment_url;
        } else {
          alert(`✅ Subscription successful! You are now subscribed to the daily draw for ₦${dailyPriceNaira}/day.`);
          navigate('/dashboard');
        }
      } else {
        setError(result.error || 'Failed to subscribe. Please try again.');
      }
    } catch (err: any) {
      console.error('Subscription error:', err);
      const errorMessage =
        err.response?.data?.error?.message ||
        err.response?.data?.message ||
        err.message ||
        'Failed to subscribe. Please try again.';
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            {configLoading ? 'Daily Subscription' : `Daily ₦${dailyPriceNaira} Subscription`}
          </h1>
          <p className="text-xl text-gray-600">
            {configLoading
              ? 'Loading subscription details...'
              : `Subscribe for just ₦${dailyPriceNaira}/day and get ${dailySpins} guaranteed daily draw entries!`}
          </p>
        </div>

        {/* Benefits Section */}
        <div className="bg-white rounded-lg shadow-xl p-8 mb-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">Subscription Benefits</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {[
              {
                title: 'Guaranteed Daily Entries',
                desc: `Get ${dailySpins} automatic entries into all daily draws`,
              },
              {
                title: 'Exclusive Prizes',
                desc: 'Access to subscriber-only prize pools',
              },
              {
                title: 'Bonus Entries',
                desc: 'Earn extra entries for consecutive days',
              },
              {
                title: 'Cancel Anytime',
                desc: 'No long-term commitment required',
              },
            ].map((benefit) => (
              <div key={benefit.title} className="flex items-start">
                <div className="flex-shrink-0">
                  <svg
                    className="h-6 w-6 text-green-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                </div>
                <div className="ml-3">
                  <h3 className="text-lg font-medium text-gray-900">{benefit.title}</h3>
                  <p className="mt-1 text-gray-600">{benefit.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Pricing Card */}
        {!configLoading && (
          <div className="bg-blue-600 text-white rounded-lg shadow-xl p-8 mb-8 text-center">
            <p className="text-lg font-medium mb-2 opacity-90">Daily Subscription Price</p>
            <p className="text-6xl font-extrabold mb-2">
              ₦{dailyPriceNaira}
              <span className="text-2xl font-normal opacity-80">/day</span>
            </p>
            <p className="text-sm opacity-75">
              {config?.auto_renewal ? 'Auto-renews daily · ' : ''}Cancel anytime
            </p>
          </div>
        )}

        {/* Subscription Form */}
        <div className="bg-white rounded-lg shadow-xl p-8 mb-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">Subscribe Now</h2>
          <form onSubmit={handleSubscribe} className="space-y-6">
            <div>
              <label
                htmlFor="phone"
                className="block text-sm font-medium text-gray-700 mb-2"
              >
                Phone Number
              </label>
              <input
                type="tel"
                id="phone"
                value={phoneNumber}
                onChange={(e) => setPhoneNumber(e.target.value)}
                placeholder="0803 123 4567"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                required
              />
            </div>

            {error && (
              <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading || configLoading}
              className="w-full bg-blue-600 text-white py-3 px-6 rounded-lg font-semibold hover:bg-blue-700 transition-colors disabled:bg-gray-400 disabled:cursor-not-allowed"
            >
              {isLoading
                ? 'Processing...'
                : configLoading
                ? 'Loading...'
                : `Subscribe for ₦${dailyPriceNaira}/day`}
            </button>
          </form>
        </div>

        {/* How It Works */}
        <div className="bg-white rounded-lg shadow-xl p-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">How It Works</h2>
          <div className="space-y-4">
            {[
              {
                step: 1,
                title: 'Subscribe',
                desc: `Enter your phone number and subscribe for just ₦${dailyPriceNaira}/day`,
              },
              {
                step: 2,
                title: 'Auto-Renewal',
                desc: 'Your subscription renews automatically every day at midnight',
              },
              {
                step: 3,
                title: 'Win Daily',
                desc: 'Get automatic entries into all daily draws and win amazing prizes',
              },
            ].map((item) => (
              <div key={item.step} className="flex items-start">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-600 text-white rounded-full flex items-center justify-center font-bold">
                  {item.step}
                </div>
                <div className="ml-4">
                  <h3 className="text-lg font-medium text-gray-900">{item.title}</h3>
                  <p className="text-gray-600">{item.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Back Button */}
        <div className="mt-8 text-center">
          <button
            onClick={() => navigate('/')}
            className="text-blue-600 hover:text-blue-800 font-medium"
          >
            ← Back to Home
          </button>
        </div>
      </div>
    </div>
  );
};

export default DailySubscriptionPage;
