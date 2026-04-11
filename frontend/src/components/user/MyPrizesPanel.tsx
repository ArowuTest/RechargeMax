import React, { useState, useEffect } from 'react';
import { apiClient } from '@/lib/api-client';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

import { NIGERIAN_BANKS } from '@/constants/banks';
import { Gift, CheckCircle2, Clock, AlertCircle, Loader2, Trophy } from 'lucide-react';

// Claim details required for cash / physical prizes
interface ClaimDetails {
  account_number: string;
  account_name: string;
  bank_name: string;
  bank_code: string;
  address: string;
  phone_number: string;
}

const EMPTY_CLAIM_DETAILS: ClaimDetails = {
  account_number: '',
  account_name: '',
  bank_name: '',
  bank_code: '',
  address: '',
  phone_number: '',
};

interface Prize {
  id: string;
  prizeType: string;
  prizeDescription: string;
  prizeAmount?: number;
  airtimeAmount?: number;
  dataPackage?: string;
  claimStatus: string;
  claimDeadline?: string;
  claimedAt?: string;
  provisionStatus?: string;
  drawName: string;
  createdAt: string;
  fulfillmentMode: string;
}

const MyPrizesPanel: React.FC = () => {
  const [prizes, setPrizes] = useState<Prize[]>([]);
  const [loading, setLoading] = useState(false);
  const [claiming, setClaiming] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [showClaimModal, setShowClaimModal] = useState(false);
  const [selectedPrize, setSelectedPrize] = useState<Prize | null>(null);
  const [claimDetails, setClaimDetails] = useState<ClaimDetails>(EMPTY_CLAIM_DETAILS);

  useEffect(() => {
    fetchMyPrizes();
  }, []);

  const fetchMyPrizes = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get('/user/prizes');
      const data = response.data;
      setPrizes(data.prizes || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load prizes');
    } finally {
      setLoading(false);
    }
  };

  const handleClaimClick = (prize: Prize) => {
    setSelectedPrize(prize);
    setClaimDetails(EMPTY_CLAIM_DETAILS);
    setShowClaimModal(true);
  };

  const needsBankDetails = (prize: Prize) =>
    prize.prizeType === 'CASH' || prize.prizeType === 'GOODS' || prize.prizeType === 'PHYSICAL';

  const handleConfirmClaim = async () => {
    if (!selectedPrize) return;

    // Validate bank details for cash/physical prizes
    if (needsBankDetails(selectedPrize)) {
      if (!claimDetails.account_number || !claimDetails.account_name || !claimDetails.bank_name) {
        setError('Please fill in your bank account number, account name, and bank name.');
        return;
      }
    }

    setClaiming(selectedPrize.id);
    setSuccess(null);
    setError(null);
    setShowClaimModal(false);

    try {
      const body = needsBankDetails(selectedPrize)
        ? {
            account_number: claimDetails.account_number,
            account_name:   claimDetails.account_name,
            bank_name:      claimDetails.bank_name,
            bank_code:      claimDetails.bank_code,
            address:        claimDetails.address,
            phone_number:   claimDetails.phone_number,
          }
        : {};

      await apiClient.post(`/winner/${selectedPrize.id}/claim`, body);

      setSuccess(
        needsBankDetails(selectedPrize)
          ? 'Claim submitted! Bank details saved. Payment will be processed within 24-48 hours.'
          : 'Prize claimed successfully! Your reward will be delivered shortly.',
      );
      setTimeout(() => setSuccess(null), 8000);
      await fetchMyPrizes();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to claim prize');
    } finally {
      setClaiming(null);
      setSelectedPrize(null);
    }
  };

  const getPrizeValue = (prize: Prize) => {
    if (prize.airtimeAmount) {
      return `₦${(prize.airtimeAmount / 100).toFixed(2)} Airtime`;
    }
    if (prize.dataPackage) {
      return `${prize.dataPackage} Data`;
    }
    if (prize.prizeAmount) {
      return `₦${(prize.prizeAmount / 100).toFixed(2)}`;
    }
    return prize.prizeDescription;
  };

  const getStatusBadge = (prize: Prize) => {
    if (prize.claimStatus === 'CLAIMED' || prize.provisionStatus === 'COMPLETED') {
      return <Badge className="bg-green-500">Claimed</Badge>;
    }
    if (prize.claimStatus === 'PENDING_ADMIN_REVIEW') {
      return <Badge className="bg-blue-500">Processing</Badge>;
    }
    if (prize.provisionStatus === 'failed') {
      return <Badge className="bg-red-500">Failed</Badge>;
    }
    if (prize.claimStatus === 'PENDING') {
      return <Badge className="bg-orange-500">Unclaimed</Badge>;
    }
    return <Badge variant="outline">{prize.claimStatus}</Badge>;
  };

  const getDaysLeft = (deadline?: string) => {
    if (!deadline) return null;
    const days = Math.ceil((new Date(deadline).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24));
    return days;
  };

  const unclaimedPrizes = prizes.filter(p => 
    p.claimStatus === 'PENDING' && p.fulfillmentMode === 'manual_claim'
  );
  const claimedPrizes = prizes.filter(p => 
    p.claimStatus === 'CLAIMED' || p.provisionStatus === 'COMPLETED'
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold flex items-center gap-2">
            <Trophy className="w-8 h-8 text-yellow-500" />
            My Prizes
          </h2>
          <p className="text-gray-600 mt-1">
            View and claim your winning prizes
          </p>
        </div>
      </div>

      {/* Alerts */}
      {success && (
        <Alert className="bg-green-50 border-green-200">
          <CheckCircle2 className="w-4 h-4 text-green-600" />
          <AlertDescription className="text-green-800">{success}</AlertDescription>
        </Alert>
      )}
      
      {error && (
        <Alert className="bg-red-50 border-red-200">
          <AlertCircle className="w-4 h-4 text-red-600" />
          <AlertDescription className="text-red-800">{error}</AlertDescription>
        </Alert>
      )}

      {/* Unclaimed Prizes - Prominent Display */}
      {unclaimedPrizes.length > 0 && (
        <Card className="border-2 border-orange-300 bg-orange-50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-orange-800">
              <Gift className="w-6 h-6" />
              Unclaimed Prizes ({unclaimedPrizes.length})
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {unclaimedPrizes.map((prize) => {
                const daysLeft = getDaysLeft(prize.claimDeadline);
                const isUrgent = daysLeft !== null && daysLeft <= 7;
                
                return (
                  <Card key={prize.id} className={`${isUrgent ? 'border-red-300' : ''}`}>
                    <CardContent className="pt-6">
                      <div className="space-y-4">
                        {/* Prize Info */}
                        <div>
                          <div className="flex items-center justify-between mb-2">
                            <Badge variant="outline" className="text-xs">
                              {prize.drawName}
                            </Badge>
                            {getStatusBadge(prize)}
                          </div>
                          <div className="text-2xl font-bold text-gray-800">
                            {getPrizeValue(prize)}
                          </div>
                          <div className="text-sm text-gray-600 mt-1">
                            {prize.prizeDescription}
                          </div>
                        </div>

                        {/* Deadline */}
                        {daysLeft !== null && (
                          <div className={`flex items-center gap-2 text-sm ${
                            isUrgent ? 'text-red-600 font-semibold' : 'text-gray-600'
                          }`}>
                            <Clock className="w-4 h-4" />
                            {daysLeft > 0 
                              ? `${daysLeft} day${daysLeft !== 1 ? 's' : ''} left to claim`
                              : 'Deadline today!'}
                          </div>
                        )}

                        {/* Claim Button */}
                        <Button
                          onClick={() => handleClaimClick(prize)}
                          disabled={claiming === prize.id}
                          className="w-full bg-orange-600 hover:bg-orange-700"
                        >
                          {claiming === prize.id ? (
                            <>
                              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                              Claiming...
                            </>
                          ) : (
                            <>
                              <Gift className="w-4 h-4 mr-2" />
                              CLAIM NOW
                            </>
                          )}
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Claimed Prizes History */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CheckCircle2 className="w-6 h-6 text-green-500" />
            Claimed Prizes ({claimedPrizes.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          {claimedPrizes.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              No claimed prizes yet. Keep playing to win!
            </div>
          ) : (
            <div className="space-y-3">
              {claimedPrizes.map((prize) => (
                <div
                  key={prize.id}
                  className="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <Badge variant="outline" className="text-xs">
                        {prize.drawName}
                      </Badge>
                      {getStatusBadge(prize)}
                    </div>
                    <div className="font-semibold text-gray-800">
                      {getPrizeValue(prize)}
                    </div>
                    <div className="text-sm text-gray-600">
                      {prize.prizeDescription}
                    </div>
                    {prize.claimedAt && (
                      <div className="text-xs text-gray-500 mt-1">
                        Claimed on {new Date(prize.claimedAt).toLocaleDateString()}
                      </div>
                    )}
                  </div>
                  <CheckCircle2 className="w-6 h-6 text-green-500" />
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* No Prizes */}
      {prizes.length === 0 && !loading && (
        <Card>
          <CardContent className="py-12">
            <div className="text-center text-gray-500">
              <Trophy className="w-16 h-16 mx-auto mb-4 text-gray-300" />
              <h3 className="text-xl font-semibold mb-2">No Prizes Yet</h3>
              <p>Keep recharging and spinning to win amazing prizes!</p>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Claim Confirmation Modal */}
      {showClaimModal && selectedPrize && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <Card className="w-full max-w-md mx-4">
            <CardHeader>
              <CardTitle>Claim Your Prize</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="text-center p-6 bg-gradient-to-br from-orange-50 to-yellow-50 rounded-lg">
                  <Gift className="w-16 h-16 mx-auto mb-4 text-orange-500" />
                  <div className="text-3xl font-bold text-gray-800 mb-2">
                    {getPrizeValue(selectedPrize)}
                  </div>
                  <div className="text-gray-600">
                    {selectedPrize.prizeDescription}
                  </div>
                </div>

                {/* Bank / address details for CASH and GOODS prizes */}
                {selectedPrize && needsBankDetails(selectedPrize) ? (
                  <div className="space-y-3">
                    <p className="text-sm font-medium text-gray-700">
                      Please provide your bank details so we can transfer your prize.
                    </p>
                    <div className="grid grid-cols-2 gap-3">
                      <div className="col-span-2">
                        <Label htmlFor="acct-num">Account Number *</Label>
                        <Input
                          id="acct-num"
                          placeholder="0123456789"
                          value={claimDetails.account_number}
                          onChange={(e) => setClaimDetails((p) => ({ ...p, account_number: e.target.value }))}
                        />
                      </div>
                      <div className="col-span-2">
                        <Label htmlFor="acct-name">Account Name *</Label>
                        <Input
                          id="acct-name"
                          placeholder="Full name on account"
                          value={claimDetails.account_name}
                          onChange={(e) => setClaimDetails((p) => ({ ...p, account_name: e.target.value }))}
                        />
                      </div>
                      <div className="col-span-2">
                        <Label htmlFor="bank-select">Bank Name *</Label>
                        <Select
                          value={claimDetails.bank_name}
                          onValueChange={(val) => {
                            const bank = NIGERIAN_BANKS.find((b) => b.name === val);
                            setClaimDetails((p) => ({
                              ...p,
                              bank_name: val,
                              bank_code: bank?.code ?? '',
                            }));
                          }}
                        >
                          <SelectTrigger id="bank-select">
                            <SelectValue placeholder="Select your bank" />
                          </SelectTrigger>
                          <SelectContent>
                            {NIGERIAN_BANKS.map((b) => (
                              <SelectItem key={b.code} value={b.name}>
                                {b.name} ({b.code})
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      {(selectedPrize.prizeType === 'GOODS' ||
                        selectedPrize.prizeType === 'PHYSICAL') && (
                        <div className="col-span-2">
                          <Label htmlFor="address">Delivery Address</Label>
                          <Input
                            id="address"
                            placeholder="Full delivery address"
                            value={claimDetails.address}
                            onChange={(e) => setClaimDetails((p) => ({ ...p, address: e.target.value }))}
                          />
                        </div>
                      )}
                      <div className="col-span-2">
                        <Label htmlFor="phone-claim">Contact Phone</Label>
                        <Input
                          id="phone-claim"
                          placeholder="0801 234 5678"
                          value={claimDetails.phone_number}
                          onChange={(e) => setClaimDetails((p) => ({ ...p, phone_number: e.target.value }))}
                        />
                      </div>
                    </div>
                  </div>
                ) : (
                  <Alert>
                    <AlertDescription>
                      Your prize will be delivered to your registered phone number within a few minutes.
                    </AlertDescription>
                  </Alert>
                )}

                <div className="flex gap-3">
                  <Button
                    variant="outline"
                    onClick={() => {
                      setShowClaimModal(false);
                      setSelectedPrize(null);
                    }}
                    className="flex-1"
                  >
                    Cancel
                  </Button>
                  <Button
                    onClick={handleConfirmClaim}
                    className="flex-1 bg-orange-600 hover:bg-orange-700"
                  >
                    Confirm Claim
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
};

export default MyPrizesPanel;
