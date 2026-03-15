/**
 * Subscription Pricing Configuration Component
 * Allows admin to set and adjust the global price per entry
 */

import React, { useState, useEffect } from 'react';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/hooks/useToast';
import { Loader2, Edit, TrendingUp, TrendingDown, DollarSign, History } from 'lucide-react';
import {
  subscriptionPricingApi,
  type SubscriptionPricing,
} from '@/lib/api-client-extensions';

export default function SubscriptionPricingConfig() {
  const { toast } = useToast();
  const [currentPricing, setCurrentPricing] = useState<SubscriptionPricing | null>(null);
  const [pricingHistory, setPricingHistory] = useState<SubscriptionPricing[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [showUpdateDialog, setShowUpdateDialog] = useState(false);
  const [showHistoryDialog, setShowHistoryDialog] = useState(false);
  const [newPrice, setNewPrice] = useState('');
  const [priceError, setPriceError] = useState('');

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [currentResponse, historyResponse] = await Promise.all([
        subscriptionPricingApi.getCurrent(),
        subscriptionPricingApi.getHistory(),
      ]);

      if (currentResponse.success && currentResponse.data) {
        setCurrentPricing(currentResponse.data);
        setNewPrice(currentResponse.data.price_per_entry.toString());
      }

      if (historyResponse.success && historyResponse.data) {
        setPricingHistory(historyResponse.data);
      }
    } catch (error) {
      console.error('Failed to fetch pricing data:', error);
      toast({
        title: 'Error',
        description: 'Failed to load pricing configuration',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const validatePrice = (): boolean => {
    const price = parseFloat(newPrice);

    if (isNaN(price)) {
      setPriceError('Please enter a valid number');
      return false;
    }

    if (price <= 0) {
      setPriceError('Price must be greater than 0');
      return false;
    }

    if (price > 10000) {
      setPriceError('Price cannot exceed ₦10,000');
      return false;
    }

    setPriceError('');
    return true;
  };

  const handleUpdatePrice = async () => {
    if (!validatePrice()) return;

    const price = parseFloat(newPrice);
    if (currentPricing && price === currentPricing.price_per_entry) {
      toast({
        title: 'No Change',
        description: 'The new price is the same as the current price',
        variant: 'default',
      });
      return;
    }

    if (!confirm(`Are you sure you want to update the price per entry to ₦${price.toFixed(2)}? This will affect all subscription tiers.`)) {
      return;
    }

    setActionLoading(true);
    try {
      const response = await subscriptionPricingApi.update(price);
      if (response.success) {
        toast({
          title: 'Success',
          description: 'Subscription pricing updated successfully',
        });
        await fetchData();
        setShowUpdateDialog(false);
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to update pricing',
        variant: 'destructive',
      });
    } finally {
      setActionLoading(false);
    }
  };

  const calculateTierPrices = (pricePerEntry: number) => {
    return [
      { entries: 1, price: pricePerEntry * 1 },
      { entries: 5, price: pricePerEntry * 5 },
      { entries: 10, price: pricePerEntry * 10 },
      { entries: 20, price: pricePerEntry * 20 },
    ];
  };

  const getPriceChange = (): { percentage: number; direction: 'up' | 'down' | 'same' } | null => {
    if (pricingHistory.length < 2) return null;

    const current = pricingHistory[0]?.price_per_entry ?? 0;
    const previous = pricingHistory[1]?.price_per_entry ?? 0;
    const percentage = ((current - previous) / previous) * 100;

    return {
      percentage: Math.abs(percentage),
      direction: percentage > 0 ? 'up' : percentage < 0 ? 'down' : 'same',
    };
  };

  const priceChange = getPriceChange();

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin" />
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Subscription Pricing</h2>
          <p className="text-muted-foreground">
            Configure the global price per entry for all subscription tiers
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => setShowHistoryDialog(true)}>
            <History className="h-4 w-4 mr-2" />
            View History
          </Button>
          <Button onClick={() => setShowUpdateDialog(true)}>
            <Edit className="h-4 w-4 mr-2" />
            Update Price
          </Button>
        </div>
      </div>

      {/* Current Pricing Card */}
      {currentPricing && (
        <Card className="bg-gradient-to-br from-blue-50 to-indigo-50 border-blue-200">
          <CardContent className="pt-6">
            <div className="flex items-start justify-between">
              <div>
                <div className="flex items-center gap-2 mb-2">
                  <DollarSign className="h-5 w-5 text-blue-600" />
                  <p className="text-sm font-medium text-blue-900">Current Price Per Entry</p>
                </div>
                <p className="text-4xl font-bold text-blue-600 mb-2">
                  ₦{currentPricing.price_per_entry.toFixed(2)}
                </p>
                <p className="text-sm text-blue-700">
                  Effective from {new Date(currentPricing.effective_from).toLocaleDateString()}
                </p>
                {priceChange && priceChange.direction !== 'same' && (
                  <div className="flex items-center gap-2 mt-2">
                    {priceChange.direction === 'up' ? (
                      <TrendingUp className="h-4 w-4 text-green-600" />
                    ) : (
                      <TrendingDown className="h-4 w-4 text-red-600" />
                    )}
                    <span className={`text-sm font-medium ${priceChange.direction === 'up' ? 'text-green-600' : 'text-red-600'}`}>
                      {priceChange.percentage.toFixed(1)}% {priceChange.direction === 'up' ? 'increase' : 'decrease'} from previous
                    </span>
                  </div>
                )}
              </div>

              {/* Example Tier Prices */}
              <div className="bg-white rounded-lg p-4 shadow-sm">
                <p className="text-sm font-medium text-gray-700 mb-3">Example Tier Prices</p>
                <div className="space-y-2">
                  {calculateTierPrices(currentPricing.price_per_entry).map((tier) => (
                    <div key={tier.entries} className="flex items-center justify-between gap-8">
                      <span className="text-sm text-gray-600">{tier.entries} {tier.entries === 1 ? 'entry' : 'entries'}</span>
                      <span className="text-sm font-semibold text-gray-900">₦{tier.price.toFixed(2)}/day</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Impact Information */}
      <Card>
        <CardHeader>
          <CardTitle>Pricing Impact</CardTitle>
          <CardDescription>
            How pricing changes affect subscription tiers
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <div className="bg-blue-100 rounded-full p-2 mt-1">
                <DollarSign className="h-4 w-4 text-blue-600" />
              </div>
              <div>
                <p className="font-medium">Automatic Tier Price Calculation</p>
                <p className="text-sm text-muted-foreground">
                  All tier prices are automatically calculated based on the price per entry. For example, a 5-entry tier will cost 5 × price per entry.
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <div className="bg-green-100 rounded-full p-2 mt-1">
                <TrendingUp className="h-4 w-4 text-green-600" />
              </div>
              <div>
                <p className="font-medium">Existing Subscriptions</p>
                <p className="text-sm text-muted-foreground">
                  Active subscriptions will use the new pricing starting from their next billing cycle. Users will be notified of price changes.
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <div className="bg-purple-100 rounded-full p-2 mt-1">
                <History className="h-4 w-4 text-purple-600" />
              </div>
              <div>
                <p className="font-medium">Price History Tracking</p>
                <p className="text-sm text-muted-foreground">
                  All pricing changes are tracked with timestamps and admin details for audit purposes.
                </p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Update Price Dialog */}
      <Dialog open={showUpdateDialog} onOpenChange={setShowUpdateDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Update Price Per Entry</DialogTitle>
            <DialogDescription>
              Set a new price per entry. This will affect all subscription tiers.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="new_price">New Price Per Entry (₦) *</Label>
              <Input
                id="new_price"
                type="number"
                step="0.01"
                min="0.01"
                max="10000"
                value={newPrice}
                onChange={(e) => {
                  setNewPrice(e.target.value);
                  setPriceError('');
                }}
                placeholder="e.g., 20.00"
                className={priceError ? 'border-red-500' : ''}
              />
              {priceError && (
                <p className="text-red-500 text-sm mt-1">{priceError}</p>
              )}
            </div>

            {newPrice && !priceError && parseFloat(newPrice) > 0 && (
              <div className="bg-gray-50 rounded-lg p-4">
                <p className="text-sm font-medium text-gray-700 mb-3">New Tier Prices Preview</p>
                <div className="space-y-2">
                  {calculateTierPrices(parseFloat(newPrice)).map((tier) => (
                    <div key={tier.entries} className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">{tier.entries} {tier.entries === 1 ? 'entry' : 'entries'}</span>
                      <span className="text-sm font-semibold text-gray-900">₦{tier.price.toFixed(2)}/day</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowUpdateDialog(false)}
              disabled={actionLoading}
            >
              Cancel
            </Button>
            <Button
              onClick={handleUpdatePrice}
              disabled={actionLoading}
            >
              {actionLoading ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Updating...
                </>
              ) : (
                'Update Price'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Price History Dialog */}
      <Dialog open={showHistoryDialog} onOpenChange={setShowHistoryDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Pricing History</DialogTitle>
            <DialogDescription>
              View all historical pricing changes
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-3 max-h-96 overflow-y-auto">
            {pricingHistory.length === 0 ? (
              <p className="text-center text-muted-foreground py-8">No pricing history available</p>
            ) : (
              pricingHistory.map((pricing, index) => (
                <div
                  key={pricing.id}
                  className={`flex items-center justify-between p-4 rounded-lg border ${
                    index === 0 ? 'bg-blue-50 border-blue-200' : 'bg-gray-50'
                  }`}
                >
                  <div>
                    <p className="font-semibold text-lg">₦{pricing.price_per_entry.toFixed(2)}</p>
                    <p className="text-sm text-muted-foreground">
                      {new Date(pricing.effective_from).toLocaleDateString()} at{' '}
                      {new Date(pricing.effective_from).toLocaleTimeString()}
                    </p>
                  </div>
                  <div className="text-right">
                    {index === 0 && (
                      <Badge variant="default">Current</Badge>
                    )}
                    {index > 0 && pricingHistory[index - 1] && (
                      <div className="flex items-center gap-1">
                        {pricing.price_per_entry < (pricingHistory[index - 1]?.price_per_entry ?? 0) ? (
                          <>
                            <TrendingDown className="h-4 w-4 text-red-500" />
                            <span className="text-sm text-red-500">
                              {((((pricingHistory[index - 1]?.price_per_entry ?? 0) - pricing.price_per_entry) / pricing.price_per_entry) * 100).toFixed(1)}% decrease
                            </span>
                          </>
                        ) : pricing.price_per_entry > (pricingHistory[index - 1]?.price_per_entry ?? 0) ? (
                          <>
                            <TrendingUp className="h-4 w-4 text-green-500" />
                            <span className="text-sm text-green-500">
                              {(((pricing.price_per_entry - (pricingHistory[index - 1]?.price_per_entry ?? 0)) / (pricingHistory[index - 1]?.price_per_entry ?? 1)) * 100).toFixed(1)}% increase
                            </span>
                          </>
                        ) : null}
                      </div>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>

          <DialogFooter>
            <Button onClick={() => setShowHistoryDialog(false)}>Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
