/**
 * Winner Claim Processing Component
 * Enterprise-grade prize claim management with runner-up workflow
 * 
 * Features:
 * - View all winners with comprehensive filtering
 * - Approve/reject claim requests with reason tracking
 * - Process cash payouts with reference management
 * - Manage physical goods shipping with tracking
 * - Runner-up invocation workflow
 * - Automatic winner notification system
 * - Claim statistics dashboard
 * - Audit trail for all claim actions
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/hooks/useToast';
import {
  Loader2,
  CheckCircle2,
  XCircle,
  Clock,
  Package,
  DollarSign,
  Gift,
  Phone,
  User,
  Calendar,
  AlertTriangle,
  RefreshCw,
  Eye,
  Send,
  TrendingUp,
  Users,
} from 'lucide-react';
import {
  winnerClaimApi,
  type Winner,
  type ClaimApprovalRequest,
  type PayoutRequest,
  type ShippingUpdateRequest,
} from '@/lib/api-client-extensions';

interface ClaimStatistics {
  total_winners: number;
  pending_claims: number;
  approved_claims: number;
  rejected_claims: number;
  processing_payouts: number;
  completed_payouts: number;
  shipping_pending: number;
  shipped: number;
}

interface WinnerDetails extends Winner {
  runner_ups?: Winner[];
}

type ClaimStatusFilter = 'all' | 'PENDING' | 'CLAIMED' | 'EXPIRED' | 'PENDING_ADMIN_REVIEW' | 'APPROVED' | 'REJECTED';
type PrizeTypeFilter = 'all' | 'airtime' | 'data' | 'points' | 'cash' | 'physical_goods';

export default function WinnerClaimProcessing() {
  const { toast } = useToast();

  // State
  const [winners, setWinners] = useState<Winner[]>([]);
  const [statistics, setStatistics] = useState<ClaimStatistics | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedWinner, setSelectedWinner] = useState<WinnerDetails | null>(null);
  const [showDetailsDialog, setShowDetailsDialog] = useState(false);
  const [showApprovalDialog, setShowApprovalDialog] = useState(false);
  const [showPayoutDialog, setShowPayoutDialog] = useState(false);
  const [showShippingDialog, setShowShippingDialog] = useState(false);
  const [showRunnerUpDialog, setShowRunnerUpDialog] = useState(false);

  // Filters
  const [claimStatusFilter, setClaimStatusFilter] = useState<ClaimStatusFilter>('all');
  const [prizeTypeFilter, setPrizeTypeFilter] = useState<PrizeTypeFilter>('all');
  const [searchTerm, setSearchTerm] = useState('');

  // Form state
  const [approvalForm, setApprovalForm] = useState({
    action: 'approve' as 'approve' | 'reject',
    reason: '',
    notes: '',
  });

  const [payoutForm, setPayoutForm] = useState({
    payout_method: 'bank_transfer',
    payout_reference: '',
    payout_amount: 0,
    notes: '',
  });

  const [shippingForm, setShippingForm] = useState({
    tracking_number: '',
    courier_service: '',
    estimated_delivery: '',
    notes: '',
  });

  const [runnerUpSelection, setRunnerUpSelection] = useState<string>('');
  const [forfeitReason, setForfeitReason] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalWinners, setTotalWinners] = useState(0);
  const PAGE_SIZE = 20;

  useEffect(() => {
    fetchData();
  }, [claimStatusFilter, prizeTypeFilter]);

  const fetchData = async () => {
    setLoading(true);
    try {
      // Fetch winners with filters
      const winnersResponse = await winnerClaimApi.getWinners(
        currentPage,
        PAGE_SIZE,
        claimStatusFilter !== 'all' ? claimStatusFilter : undefined,
        prizeTypeFilter !== 'all' ? prizeTypeFilter : undefined,
      );

      if (winnersResponse.success && winnersResponse.data) {
        const list = Array.isArray(winnersResponse.data)
          ? winnersResponse.data
          : winnersResponse.data?.items ?? winnersResponse.data?.data ?? [];
        setWinners(list);
        setTotalWinners(winnersResponse.data?.total ?? winnersResponse.total ?? list.length);
      }

      // Fetch statistics
      const statsResponse = await winnerClaimApi.getClaimStatistics();
      if (statsResponse.success && statsResponse.data) {
        setStatistics(statsResponse.data);
      }
    } catch (error) {
      console.error('Failed to fetch winner data:', error);
      toast({
        title: 'Error',
        description: 'Failed to load winner data',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleViewDetails = async (winner: Winner) => {
    try {
      // Fetch full winner details including runner-ups
      const response = await winnerClaimApi.getWinnerDetails(winner.id);
      if (response.success && response.data) {
        setSelectedWinner(response.data);
        setShowDetailsDialog(true);
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load winner details',
        variant: 'destructive',
      });
    }
  };

  const handleOpenApprovalDialog = (winner: Winner) => {
    setSelectedWinner(winner);
    setApprovalForm({
      action: 'approve',
      reason: '',
      notes: '',
    });
    setShowApprovalDialog(true);
  };

  const handleApproveClaim = async () => {
    if (!selectedWinner) return;

    if (!approvalForm.reason.trim()) {
      toast({
        title: 'Reason Required',
        description: 'Please provide a reason for this action',
        variant: 'destructive',
      });
      return;
    }

    try {
      const request: ClaimApprovalRequest = {
        winner_id: selectedWinner.id,
        action: approvalForm.action,
        reason: approvalForm.reason,
        notes: approvalForm.notes,
      };

      const response = await winnerClaimApi.approveClaim(request);

      if (response.success) {
        toast({
          title: 'Success',
          description: `Claim ${approvalForm.action === 'approve' ? 'APPROVED' : 'REJECTED'} successfully`,
        });

        setShowApprovalDialog(false);
        await fetchData();
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to process claim',
        variant: 'destructive',
      });
    }
  };

  const handleOpenPayoutDialog = (winner: Winner) => {
    setSelectedWinner(winner);
    setPayoutForm({
      payout_method: 'bank_transfer',
      payout_reference: '',
      payout_amount: winner.prize_value || 0,
      notes: '',
    });
    setShowPayoutDialog(true);
  };

  const handleProcessPayout = async () => {
    if (!selectedWinner) return;

    if (!payoutForm.payout_reference.trim()) {
      toast({
        title: 'Reference Required',
        description: 'Please provide a payout reference',
        variant: 'destructive',
      });
      return;
    }

    if (payoutForm.payout_amount <= 0) {
      toast({
        title: 'Invalid Amount',
        description: 'Payout amount must be greater than zero',
        variant: 'destructive',
      });
      return;
    }

    try {
      const request: PayoutRequest = {
        winner_id: selectedWinner.id,
        payout_method: payoutForm.payout_method,
        payout_reference: payoutForm.payout_reference,
        payout_amount: payoutForm.payout_amount,
        notes: payoutForm.notes,
      };

      const response = await winnerClaimApi.processPayout(request);

      if (response.success) {
        toast({
          title: 'Success',
          description: 'Payout processed successfully',
        });

        setShowPayoutDialog(false);
        await fetchData();
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to process payout',
        variant: 'destructive',
      });
    }
  };

  const handleOpenShippingDialog = (winner: Winner) => {
    setSelectedWinner(winner);
    setShippingForm({
      tracking_number: winner.tracking_number || '',
      courier_service: '',
      estimated_delivery: '',
      notes: '',
    });
    setShowShippingDialog(true);
  };

  const handleUpdateShipping = async () => {
    if (!selectedWinner) return;

    if (!shippingForm.tracking_number.trim()) {
      toast({
        title: 'Tracking Number Required',
        description: 'Please provide a tracking number',
        variant: 'destructive',
      });
      return;
    }

    try {
      const request: ShippingUpdateRequest = {
        winner_id: selectedWinner.id,
        tracking_number: shippingForm.tracking_number,
        shipping_status: 'shipped',  // Default status
        courier_service: shippingForm.courier_service,
        estimated_delivery: shippingForm.estimated_delivery,
        notes: shippingForm.notes,
      };

      const response = await winnerClaimApi.updateShipping(request);

      if (response.success) {
        toast({
          title: 'Success',
          description: 'Shipping information updated successfully',
        });

        setShowShippingDialog(false);
        await fetchData();
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to update shipping',
        variant: 'destructive',
      });
    }
  };

  const handleOpenRunnerUpDialog = (winner: WinnerDetails) => {
    if (!winner.runner_ups || winner.runner_ups.length === 0) {
      toast({
        title: 'No Runner-Ups',
        description: 'No runner-ups available for this prize',
        variant: 'destructive',
      });
      return;
    }

    setSelectedWinner(winner);
    setRunnerUpSelection('');
    setForfeitReason('');
    setShowRunnerUpDialog(true);
  };

  const handleInvokeRunnerUp = async () => {
    if (!selectedWinner || !runnerUpSelection) return;

    if (!forfeitReason.trim()) {
      toast({
        title: 'Reason Required',
        description: 'Please provide a reason for invoking runner-up',
        variant: 'destructive',
      });
      return;
    }

    try {
      const response = await winnerClaimApi.invokeRunnerUp({
        original_winner_id: selectedWinner.id,
        runner_up_id: runnerUpSelection,
        forfeit_reason: forfeitReason,
      });

      if (response.success) {
        toast({
          title: 'Success',
          description: 'Runner-up invoked successfully. Notifications sent.',
        });

        setShowRunnerUpDialog(false);
        await fetchData();
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to invoke runner-up',
        variant: 'destructive',
      });
    }
  };

  const handleSendNotification = async (winnerId: string) => {
    try {
      const response = await winnerClaimApi.sendNotification(winnerId);

      if (response.success) {
        toast({
          title: 'Success',
          description: 'Notification sent successfully',
        });
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to send notification',
        variant: 'destructive',
      });
    }
  };

  const maskMSISDN = (msisdn: string): string => {
    if (msisdn.length < 7) return msisdn;
    const first3 = msisdn.substring(0, 3);
    const last3 = msisdn.substring(msisdn.length - 3);
    const masked = '*'.repeat(msisdn.length - 6);
    return `${first3}${masked}${last3}`;
  };

  const getClaimStatusBadge = (status: string) => {
    const statusConfig: Record<string, { variant: 'secondary'|'default'|'destructive'|'outline'; icon: React.ElementType; label: string }> = {
      PENDING:              { variant: 'secondary',    icon: Clock,         label: 'Pending' },
      APPROVED:             { variant: 'default',      icon: CheckCircle2,  label: 'Approved' },
      REJECTED:             { variant: 'destructive',  icon: XCircle,       label: 'Rejected' },
      CLAIMED:              { variant: 'default',      icon: CheckCircle2,  label: 'Claimed' },
      EXPIRED:              { variant: 'outline',      icon: Clock,         label: 'Expired' },
      PENDING_ADMIN_REVIEW: { variant: 'secondary',    icon: RefreshCw,     label: 'Under Review' },
    };

    const config = statusConfig[status?.toUpperCase?.() ?? ''] ?? statusConfig['PENDING']!;
    const Icon = config.icon;

    return (
      <Badge variant={config.variant as any} className="flex items-center gap-1">
        <Icon className="h-3 w-3" />
        {config.label}
      </Badge>
    );
  };

  const getPrizeTypeIcon = (prizeType: string) => {
    const icons = {
      airtime: Phone,
      data: TrendingUp,
      points: Gift,
      cash: DollarSign,
      physical_goods: Package,
    };
    return icons[prizeType as keyof typeof icons] || Gift;
  };

  const filteredWinners = winners.filter((winner) => {
    if (searchTerm) {
      const search = searchTerm.toLowerCase();
      return (
        winner.msisdn.includes(search) ||
        winner.prize_name.toLowerCase().includes(search) ||
        winner.prize_type.toLowerCase().includes(search)
      );
    }
    return true;
  });

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
      <div>
        <h2 className="text-2xl font-bold">Winner Claim Processing</h2>
        <p className="text-muted-foreground">
          Manage prize claims, process payouts, and handle shipping
        </p>
      </div>

      {/* Statistics Cards */}
      {statistics && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Winners
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">{statistics.total_winners}</p>
                <Users className="h-8 w-8 text-blue-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Pending Claims
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">{statistics.pending_claims}</p>
                <Clock className="h-8 w-8 text-yellow-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Processing Payouts
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">{statistics.processing_payouts}</p>
                <RefreshCw className="h-8 w-8 text-orange-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Completed
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <p className="text-2xl font-bold">{statistics.completed_payouts}</p>
                <CheckCircle2 className="h-8 w-8 text-green-600" />
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <Label htmlFor="claim-status-filter">Claim Status</Label>
              <Select
                value={claimStatusFilter}
                onValueChange={(value) => setClaimStatusFilter(value as ClaimStatusFilter)}
              >
                <SelectTrigger id="claim-status-filter">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="pending">Pending</SelectItem>
                  <SelectItem value="approved">Approved</SelectItem>
                  <SelectItem value="rejected">Rejected</SelectItem>
                  <SelectItem value="processing">Processing</SelectItem>
                  <SelectItem value="completed">Completed</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="prize-type-filter">Prize Type</Label>
              <Select
                value={prizeTypeFilter}
                onValueChange={(value) => setPrizeTypeFilter(value as PrizeTypeFilter)}
              >
                <SelectTrigger id="prize-type-filter">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Types</SelectItem>
                  <SelectItem value="airtime">Airtime</SelectItem>
                  <SelectItem value="data">Data</SelectItem>
                  <SelectItem value="points">Points</SelectItem>
                  <SelectItem value="cash">Cash</SelectItem>
                  <SelectItem value="physical_goods">Physical Goods</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="search">Search</Label>
              <Input
                id="search"
                placeholder="Search by MSISDN or prize..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Winners Table */}
      <Card>
        <CardHeader>
          <CardTitle>Winners ({filteredWinners.length})</CardTitle>
          <CardDescription>
            Manage prize claims and process payouts
          </CardDescription>
        </CardHeader>
        <CardContent>
          {filteredWinners.length === 0 ? (
            <p className="text-center text-muted-foreground py-8">
              No winners found matching the current filters
            </p>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>MSISDN</TableHead>
                    <TableHead>Prize</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Value</TableHead>
                    <TableHead>Claim Status</TableHead>
                    <TableHead>Draw Date</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredWinners.map((winner) => {
                    const PrizeIcon = getPrizeTypeIcon(winner.prize_type);
                    return (
                      <TableRow key={winner.id}>
                        <TableCell className="font-mono">
                          {maskMSISDN(winner.msisdn)}
                        </TableCell>
                        <TableCell className="font-medium">
                          {winner.prize_name}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            <PrizeIcon className="h-4 w-4 text-muted-foreground" />
                            <span className="capitalize">{winner.prize_type.replace('_', ' ')}</span>
                          </div>
                        </TableCell>
                        <TableCell>
                          {winner.prize_type === 'cash'
                            ? `₦${winner.prize_value?.toLocaleString()}`
                            : winner.prize_value}
                        </TableCell>
                        <TableCell>
                          {getClaimStatusBadge(winner.claim_status || 'PENDING')}
                        </TableCell>
                        <TableCell>
                          {winner.draw_date ? new Date(winner.draw_date).toLocaleDateString() : 'N/A'}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleViewDetails(winner)}
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                            {winner.claim_status === 'PENDING' && (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleOpenApprovalDialog(winner)}
                              >
                                <CheckCircle2 className="h-4 w-4" />
                              </Button>
                            )}
                            {winner.claim_status === 'APPROVED' &&
                              winner.prize_type === 'cash' && (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleOpenPayoutDialog(winner)}
                                >
                                  <DollarSign className="h-4 w-4" />
                                </Button>
                              )}
                            {winner.claim_status === 'APPROVED' &&
                              winner.prize_type === 'physical_goods' && (
                                <Button
                                  variant="ghost"
                                  size="sm"
                                  onClick={() => handleOpenShippingDialog(winner)}
                                >
                                  <Package className="h-4 w-4" />
                                </Button>
                              )}
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleSendNotification(winner.id)}
                            >
                              <Send className="h-4 w-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
            {/* Pagination */}
            <div className="flex items-center justify-between mt-4 px-1">
              <p className="text-xs text-muted-foreground">
                Page {currentPage} · Showing {winners.length} of {totalWinners} claims
              </p>
              <div className="flex gap-2">
                <button className="px-3 py-1 text-xs border rounded disabled:opacity-40"
                  disabled={currentPage <= 1}
                  onClick={() => setCurrentPage(p => Math.max(1, p - 1))}>← Prev</button>
                <button className="px-3 py-1 text-xs border rounded disabled:opacity-40"
                  disabled={winners.length < PAGE_SIZE}
                  onClick={() => setCurrentPage(p => p + 1)}>Next →</button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Winner Details Dialog */}
      <Dialog open={showDetailsDialog} onOpenChange={setShowDetailsDialog}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>Winner Details</DialogTitle>
            <DialogDescription>
              Complete information about the winner and prize
            </DialogDescription>
          </DialogHeader>

          {selectedWinner && (
            <div className="space-y-6">
              {/* Winner Information */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-muted-foreground">MSISDN</Label>
                  <p className="font-mono font-medium">{selectedWinner.msisdn}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Prize</Label>
                  <p className="font-medium">{selectedWinner.prize_name}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Prize Type</Label>
                  <p className="capitalize">{selectedWinner.prize_type.replace('_', ' ')}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Prize Value</Label>
                  <p className="font-medium">
                    {selectedWinner.prize_type === 'cash'
                      ? `₦${selectedWinner.prize_value?.toLocaleString()}`
                      : selectedWinner.prize_value}
                  </p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Claim Status</Label>
                  <div className="mt-1">
                    {getClaimStatusBadge(selectedWinner.claim_status || 'PENDING')}
                  </div>
                </div>
                <div>
                  <Label className="text-muted-foreground">Draw Date</Label>
                  <p>{selectedWinner.draw_date ? new Date(selectedWinner.draw_date).toLocaleDateString() : 'N/A'}</p>
                </div>
              </div>

              {/* Bank Details (for cash prizes) */}
              {selectedWinner.prize_type === 'cash' && selectedWinner.bank_name && (
                <div>
                  <h4 className="font-semibold mb-3">Bank Details</h4>
                  <div className="grid grid-cols-2 gap-4 bg-gray-50 p-4 rounded-lg">
                    <div>
                      <Label className="text-muted-foreground">Bank Name</Label>
                      <p>{selectedWinner.bank_name}</p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground">Account Number</Label>
                      <p className="font-mono">{selectedWinner.account_number}</p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground">Account Name</Label>
                      <p>{selectedWinner.account_name}</p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground">Bank Code</Label>
                      <p className="font-mono">{selectedWinner.bank_code}</p>
                    </div>
                  </div>
                </div>
              )}

              {/* Shipping Details (for physical goods) */}
              {selectedWinner.prize_type === 'physical_goods' && selectedWinner.shipping_address && (
                <div>
                  <h4 className="font-semibold mb-3">Shipping Details</h4>
                  <div className="grid grid-cols-2 gap-4 bg-gray-50 p-4 rounded-lg">
                    <div className="col-span-2">
                      <Label className="text-muted-foreground">Shipping Address</Label>
                      <p>{selectedWinner.shipping_address}</p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground">Phone</Label>
                      <p className="font-mono">{selectedWinner.shipping_phone}</p>
                    </div>
                    {selectedWinner.tracking_number && (
                      <div>
                        <Label className="text-muted-foreground">Tracking Number</Label>
                        <p className="font-mono">{selectedWinner.tracking_number}</p>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Runner-Ups */}
              {selectedWinner.runner_ups && selectedWinner.runner_ups.length > 0 && (
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <h4 className="font-semibold">Runner-Ups ({selectedWinner.runner_ups.length})</h4>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleOpenRunnerUpDialog(selectedWinner)}
                    >
                      <RefreshCw className="h-4 w-4 mr-2" />
                      Invoke Runner-Up
                    </Button>
                  </div>
                  <div className="space-y-2">
                    {selectedWinner.runner_ups.map((runnerUp, index) => (
                      <div
                        key={runnerUp.id}
                        className="flex items-center justify-between bg-gray-50 p-3 rounded-lg"
                      >
                        <div className="flex items-center gap-3">
                          <Badge variant="secondary">{index + 1}</Badge>
                          <span className="font-mono">{maskMSISDN(runnerUp.msisdn)}</span>
                        </div>
                        <Badge variant="outline">{runnerUp.claim_status || 'PENDING'}</Badge>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          <DialogFooter>
            <Button onClick={() => setShowDetailsDialog(false)}>Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Approval Dialog */}
      <Dialog open={showApprovalDialog} onOpenChange={setShowApprovalDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Process Claim</DialogTitle>
            <DialogDescription>
              Approve or reject the prize claim request
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="action">Action</Label>
              <Select
                value={approvalForm.action}
                onValueChange={(value) =>
                  setApprovalForm({ ...approvalForm, action: value as 'approve' | 'reject' })
                }
              >
                <SelectTrigger id="action">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="approve">Approve Claim</SelectItem>
                  <SelectItem value="reject">Reject Claim</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="reason">Reason *</Label>
              <Input
                id="reason"
                placeholder="Enter reason for this action"
                value={approvalForm.reason}
                onChange={(e) =>
                  setApprovalForm({ ...approvalForm, reason: e.target.value })
                }
              />
            </div>

            <div>
              <Label htmlFor="notes">Additional Notes</Label>
              <Textarea
                id="notes"
                placeholder="Optional additional notes"
                value={approvalForm.notes}
                onChange={(e) =>
                  setApprovalForm({ ...approvalForm, notes: e.target.value })
                }
                rows={3}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowApprovalDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleApproveClaim}>
              {approvalForm.action === 'approve' ? (
                <>
                  <CheckCircle2 className="h-4 w-4 mr-2" />
                  Approve
                </>
              ) : (
                <>
                  <XCircle className="h-4 w-4 mr-2" />
                  Reject
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Payout Dialog */}
      <Dialog open={showPayoutDialog} onOpenChange={setShowPayoutDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Process Cash Payout</DialogTitle>
            <DialogDescription>
              Enter payout details and reference
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="payout_method">Payout Method</Label>
              <Select
                value={payoutForm.payout_method}
                onValueChange={(value) =>
                  setPayoutForm({ ...payoutForm, payout_method: value })
                }
              >
                <SelectTrigger id="payout_method">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="bank_transfer">Bank Transfer</SelectItem>
                  <SelectItem value="mobile_money">Mobile Money</SelectItem>
                  <SelectItem value="airtime">Airtime Credit</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="payout_reference">Payout Reference *</Label>
              <Input
                id="payout_reference"
                placeholder="Enter transaction reference"
                value={payoutForm.payout_reference}
                onChange={(e) =>
                  setPayoutForm({ ...payoutForm, payout_reference: e.target.value })
                }
              />
            </div>

            <div>
              <Label htmlFor="payout_amount">Payout Amount (₦) *</Label>
              <Input
                id="payout_amount"
                type="number"
                min="0"
                step="0.01"
                value={payoutForm.payout_amount}
                onChange={(e) =>
                  setPayoutForm({ ...payoutForm, payout_amount: parseFloat(e.target.value) })
                }
              />
            </div>

            <div>
              <Label htmlFor="payout_notes">Notes</Label>
              <Textarea
                id="payout_notes"
                placeholder="Optional payout notes"
                value={payoutForm.notes}
                onChange={(e) =>
                  setPayoutForm({ ...payoutForm, notes: e.target.value })
                }
                rows={3}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowPayoutDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleProcessPayout}>
              <DollarSign className="h-4 w-4 mr-2" />
              Process Payout
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Shipping Dialog */}
      <Dialog open={showShippingDialog} onOpenChange={setShowShippingDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Update Shipping Information</DialogTitle>
            <DialogDescription>
              Enter tracking details for physical goods delivery
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="tracking_number">Tracking Number *</Label>
              <Input
                id="tracking_number"
                placeholder="Enter tracking number"
                value={shippingForm.tracking_number}
                onChange={(e) =>
                  setShippingForm({ ...shippingForm, tracking_number: e.target.value })
                }
              />
            </div>

            <div>
              <Label htmlFor="courier_service">Courier Service</Label>
              <Input
                id="courier_service"
                placeholder="e.g., DHL, FedEx, NIPOST"
                value={shippingForm.courier_service}
                onChange={(e) =>
                  setShippingForm({ ...shippingForm, courier_service: e.target.value })
                }
              />
            </div>

            <div>
              <Label htmlFor="estimated_delivery">Estimated Delivery Date</Label>
              <input
                id="estimated_delivery"
                type="date"
                value={shippingForm.estimated_delivery}
                onChange={(e) =>
                  setShippingForm({ ...shippingForm, estimated_delivery: e.target.value })
                }
                className="w-full px-3 py-2 border rounded-md"
              />
            </div>

            <div>
              <Label htmlFor="shipping_notes">Notes</Label>
              <Textarea
                id="shipping_notes"
                placeholder="Optional shipping notes"
                value={shippingForm.notes}
                onChange={(e) =>
                  setShippingForm({ ...shippingForm, notes: e.target.value })
                }
                rows={3}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowShippingDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateShipping}>
              <Package className="h-4 w-4 mr-2" />
              Update Shipping
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Runner-Up Invocation Dialog */}
      <Dialog open={showRunnerUpDialog} onOpenChange={setShowRunnerUpDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Invoke Runner-Up</DialogTitle>
            <DialogDescription>
              Select a runner-up to replace the current winner
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
              <div className="flex gap-2">
                <AlertTriangle className="h-5 w-5 text-yellow-600 flex-shrink-0" />
                <div className="text-sm text-yellow-800">
                  <p className="font-medium mb-1">Warning: This action will:</p>
                  <ul className="list-disc list-inside space-y-1">
                    <li>Mark the current winner's prize as forfeited</li>
                    <li>Notify the current winner of the forfeit</li>
                    <li>Assign the prize to the selected runner-up</li>
                    <li>Notify the runner-up of their win</li>
                  </ul>
                </div>
              </div>
            </div>

            <div>
              <Label htmlFor="runner_up_select">Select Runner-Up *</Label>
              <Select value={runnerUpSelection} onValueChange={setRunnerUpSelection}>
                <SelectTrigger id="runner_up_select">
                  <SelectValue placeholder="Choose a runner-up" />
                </SelectTrigger>
                <SelectContent>
                  {selectedWinner?.runner_ups?.map((runnerUp, index) => (
                    <SelectItem key={runnerUp.id} value={runnerUp.id}>
                      Runner-Up #{index + 1} - {maskMSISDN(runnerUp.msisdn)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="forfeit_reason">Forfeit Reason *</Label>
              <Textarea
                id="forfeit_reason"
                placeholder="Enter reason for invoking runner-up (e.g., winner did not claim within deadline, invalid bank details, etc.)"
                value={forfeitReason}
                onChange={(e) => setForfeitReason(e.target.value)}
                rows={4}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowRunnerUpDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleInvokeRunnerUp} disabled={!runnerUpSelection || !forfeitReason}>
              <RefreshCw className="h-4 w-4 mr-2" />
              Invoke Runner-Up
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
