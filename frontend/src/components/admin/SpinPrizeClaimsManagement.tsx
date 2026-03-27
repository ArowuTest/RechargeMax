import React, { useState, useEffect } from 'react';
import { adminApi } from '@/lib/api-client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { useToast } from '@/hooks/useToast';
import { 
  CheckCircle, 
  XCircle, 
  Clock, 
  Download,
  Eye,
  Loader2,
  DollarSign,
  Gift,
  TrendingUp
} from 'lucide-react';

interface SpinClaim {
  id: string;
  spin_code: string;
  msisdn: string;
  user_id: string;
  user_name: string;
  prize_type: string;
  prize_name: string;
  prize_value: number;
  claim_status: string;
  created_at: string;
  claim_date: string;
  bank_details?: {
    account_number: string;
    account_name: string;
    bank_name: string;
  };
}

interface ClaimStatistics {
  overview: {
    total_claims: number;
    pending_review: number;
    approved: number;
    rejected: number;
    auto_claimed: number;
  };
  by_prize_type: {
    [key: string]: {
      total: number;
      total_value: number;
      claimed?: number;
      pending_review?: number;
      approved?: number;
      rejected?: number;
    };
  };
  average_review_time_hours: number;
  total_value_pending: number;
  total_value_approved: number;
  total_value_rejected: number;
}

const SpinPrizeClaimsManagement: React.FC = () => {
  const { toast } = useToast();
  const [claims, setClaims] = useState<SpinClaim[]>([]);
  const [statistics, setStatistics] = useState<ClaimStatistics | null>(null);
  const [loading, setLoading] = useState(false);
  const [selectedClaim, setSelectedClaim] = useState<SpinClaim | null>(null);
  const [showDetailsDialog, setShowDetailsDialog] = useState(false);
  const [showApproveDialog, setShowApproveDialog] = useState(false);
  const [showRejectDialog, setShowRejectDialog] = useState(false);
  
  // Filters
  const [statusFilter, setStatusFilter] = useState('all');
  const [prizeTypeFilter, setPrizeTypeFilter] = useState('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  
  // Approve/Reject form data
  const [adminNotes, setAdminNotes] = useState('');
  const [paymentReference, setPaymentReference] = useState('');
  const [rejectionReason, setRejectionReason] = useState('');
  const [processing, setProcessing] = useState(false);

  useEffect(() => {
    fetchClaims();
    fetchStatistics();
  }, [statusFilter, prizeTypeFilter, searchQuery, currentPage]);

  const fetchClaims = async () => {
    setLoading(true);
    try {
      const params: any = {
        page: currentPage,
        limit: 20,
      };
      
      if (statusFilter !== 'all') params.status = statusFilter;
      if (prizeTypeFilter !== 'all') params.prize_type = prizeTypeFilter;
      if (searchQuery) params.search = searchQuery;

      const response = await adminApi.getSpinClaims(params);
      
      if (response.success) {
        setClaims(response.data.claims || []);
        setTotalPages(response.data.pagination?.total_pages || 1);
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to fetch claims',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchStatistics = async () => {
    try {
      const response = await adminApi.getSpinClaimStatistics();
      if (response.success) {
        setStatistics(response.data);
      }
    } catch (error: any) {
      console.error('Failed to fetch statistics:', error);
    }
  };

  const handleViewDetails = async (claim: SpinClaim) => {
    try {
      const response = await adminApi.getSpinClaimDetails(claim.id);
      if (response.success) {
        setSelectedClaim(response.data.claim);
        setShowDetailsDialog(true);
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to fetch claim details',
        variant: 'destructive',
      });
    }
  };

  const handleApproveClaim = async () => {
    if (!selectedClaim) return;
    
    setProcessing(true);
    try {
      const response = await adminApi.approveSpinClaim(selectedClaim.id, {
        admin_notes: adminNotes,
        payment_reference: paymentReference,
      });
      
      if (response.success) {
        toast({
          title: 'Success',
          description: 'Claim approved successfully',
        });
        setShowApproveDialog(false);
        setShowDetailsDialog(false);
        setAdminNotes('');
        setPaymentReference('');
        fetchClaims();
        fetchStatistics();
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to approve claim',
        variant: 'destructive',
      });
    } finally {
      setProcessing(false);
    }
  };

  const handleRejectClaim = async () => {
    if (!selectedClaim || !rejectionReason.trim()) {
      toast({
        title: 'Error',
        description: 'Rejection reason is required',
        variant: 'destructive',
      });
      return;
    }
    
    setProcessing(true);
    try {
      const response = await adminApi.rejectSpinClaim(selectedClaim.id, {
        rejection_reason: rejectionReason,
        admin_notes: adminNotes,
      });
      
      if (response.success) {
        toast({
          title: 'Success',
          description: 'Claim rejected successfully',
        });
        setShowRejectDialog(false);
        setShowDetailsDialog(false);
        setRejectionReason('');
        setAdminNotes('');
        fetchClaims();
        fetchStatistics();
      }
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to reject claim',
        variant: 'destructive',
      });
    } finally {
      setProcessing(false);
    }
  };

  const handleExportClaims = async () => {
    try {
      const params: any = {};
      if (statusFilter !== 'all') params.status = statusFilter;
      if (prizeTypeFilter !== 'all') params.prize_type = prizeTypeFilter;
      if (searchQuery) params.search = searchQuery;

      await adminApi.exportSpinClaims(params);
      
      toast({
        title: 'Success',
        description: 'Claims exported successfully',
      });
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.message || 'Failed to export claims',
        variant: 'destructive',
      });
    }
  };

  const getStatusBadge = (status: string) => {
    const statusConfig: any = {
      PENDING: { color: 'bg-yellow-500', label: 'Pending' },
      CLAIMED: { color: 'bg-green-500', label: 'Claimed' },
      PENDING_ADMIN_REVIEW: { color: 'bg-blue-500', label: 'Pending Review' },
      APPROVED: { color: 'bg-green-600', label: 'Approved' },
      REJECTED: { color: 'bg-red-500', label: 'Rejected' },
      EXPIRED: { color: 'bg-gray-500', label: 'Expired' },
    };
    
    const config = statusConfig[status] || { color: 'bg-gray-500', label: status };
    return <Badge className={config.color}>{config.label}</Badge>;
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-NG', {
      style: 'currency',
      currency: 'NGN',
    }).format(amount);
  };

  const formatDate = (dateString: string | null | undefined) => {
    if (!dateString) return '—';
    const d = new Date(dateString);
    return isNaN(d.getTime()) ? '—' : d.toLocaleString('en-NG');
  };

  return (
    <div className="space-y-6">
      {/* Statistics Cards */}
      {statistics && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Claims</CardTitle>
              <Gift className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{statistics.overview.total_claims}</div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Pending Review</CardTitle>
              <Clock className="h-4 w-4 text-blue-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-blue-600">{statistics.overview.pending_review}</div>
              <p className="text-xs text-muted-foreground">
                {formatCurrency(statistics.total_value_pending)} pending
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Approved</CardTitle>
              <CheckCircle className="h-4 w-4 text-green-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-600">{statistics.overview.approved}</div>
              <p className="text-xs text-muted-foreground">
                {formatCurrency(statistics.total_value_approved)} approved
              </p>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Avg Review Time</CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{statistics.average_review_time_hours.toFixed(1)}h</div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Filters and Actions */}
      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <div>
              <CardTitle>Spin Prize Claims</CardTitle>
              <CardDescription>Manage user prize claim requests</CardDescription>
            </div>
            <Button onClick={handleExportClaims} variant="outline">
              <Download className="mr-2 h-4 w-4" />
              Export CSV
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4 mb-4">
            <div className="flex-1">
              <Input
                placeholder="Search by MSISDN, spin code, or user name..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[200px]">
                <SelectValue placeholder="Filter by status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Statuses</SelectItem>
                <SelectItem value="PENDING_ADMIN_REVIEW">Pending Review</SelectItem>
                <SelectItem value="APPROVED">Approved</SelectItem>
                <SelectItem value="REJECTED">Rejected</SelectItem>
                <SelectItem value="CLAIMED">Claimed</SelectItem>
              </SelectContent>
            </Select>
            <Select value={prizeTypeFilter} onValueChange={setPrizeTypeFilter}>
              <SelectTrigger className="w-[200px]">
                <SelectValue placeholder="Filter by type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Types</SelectItem>
                <SelectItem value="AIRTIME">Airtime</SelectItem>
                <SelectItem value="DATA">Data</SelectItem>
                <SelectItem value="CASH">Cash</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Claims Table */}
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Spin Code</TableHead>
                  <TableHead>User</TableHead>
                  <TableHead>Prize</TableHead>
                  <TableHead>Value</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Claim Date</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-8">
                      <Loader2 className="h-6 w-6 animate-spin mx-auto" />
                    </TableCell>
                  </TableRow>
                ) : claims.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                      No claims found
                    </TableCell>
                  </TableRow>
                ) : (
                  claims.map((claim) => (
                    <TableRow key={claim.id}>
                      <TableCell className="font-mono text-sm">{claim.spin_code}</TableCell>
                      <TableCell>
                        <div>
                          <div className="font-medium">{claim.user_name || claim.msisdn || 'N/A'}</div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div>
                          <div className="font-medium">{claim.prize_name}</div>
                          <div className="text-sm text-muted-foreground">{claim.prize_type}</div>
                        </div>
                      </TableCell>
                      <TableCell className="font-semibold">{formatCurrency(claim.prize_value)}</TableCell>
                      <TableCell>{getStatusBadge(claim.claim_status)}</TableCell>
                      <TableCell className="text-sm">{formatDate(claim.claim_date ?? claim.created_at)}</TableCell>
                      <TableCell>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleViewDetails(claim)}
                        >
                          <Eye className="h-4 w-4 mr-1" />
                          View
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-4">
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage === 1}
                onClick={() => setCurrentPage(currentPage - 1)}
              >
                Previous
              </Button>
              <span className="py-2 px-4 text-sm">
                Page {currentPage} of {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage === totalPages}
                onClick={() => setCurrentPage(currentPage + 1)}
              >
                Next
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Claim Details Dialog */}
      {selectedClaim && (
        <Dialog open={showDetailsDialog} onOpenChange={setShowDetailsDialog}>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Claim Details - {selectedClaim.spin_code}</DialogTitle>
              <DialogDescription>
                Review claim information and take action
              </DialogDescription>
            </DialogHeader>
            
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-muted-foreground">User</Label>
                  <p className="font-medium">{selectedClaim.user_name || selectedClaim.msisdn || 'N/A'}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Prize</Label>
                  <p className="font-medium">{selectedClaim.prize_name}</p>
                  <p className="text-sm text-muted-foreground">{selectedClaim.prize_type}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Value</Label>
                  <p className="font-medium text-lg">{formatCurrency(selectedClaim.prize_value)}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground">Status</Label>
                  <div className="mt-1">{getStatusBadge(selectedClaim.claim_status)}</div>
                </div>
                <div>
                  <Label className="text-muted-foreground">Claim Date</Label>
                  <p className="text-sm">{formatDate(selectedClaim.claim_date ?? selectedClaim.created_at)}</p>
                </div>
              </div>

              {selectedClaim.bank_details && (
                <div className="border rounded-lg p-4 bg-muted/50">
                  <Label className="text-lg font-semibold mb-2 block">Bank Details</Label>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label className="text-muted-foreground">Account Name</Label>
                      <p className="font-medium">{selectedClaim.bank_details.account_name}</p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground">Account Number</Label>
                      <p className="font-medium font-mono">{selectedClaim.bank_details.account_number}</p>
                    </div>
                    <div>
                      <Label className="text-muted-foreground">Bank Name</Label>
                      <p className="font-medium">{selectedClaim.bank_details.bank_name}</p>
                    </div>
                  </div>
                </div>
              )}

              {selectedClaim.claim_status === 'PENDING_ADMIN_REVIEW' && (
                <div className="flex gap-2 pt-4">
                  <Button
                    className="flex-1"
                    onClick={() => setShowApproveDialog(true)}
                  >
                    <CheckCircle className="mr-2 h-4 w-4" />
                    Approve Claim
                  </Button>
                  <Button
                    variant="destructive"
                    className="flex-1"
                    onClick={() => setShowRejectDialog(true)}
                  >
                    <XCircle className="mr-2 h-4 w-4" />
                    Reject Claim
                  </Button>
                </div>
              )}
            </div>
          </DialogContent>
        </Dialog>
      )}

      {/* Approve Dialog */}
      <Dialog open={showApproveDialog} onOpenChange={setShowApproveDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Approve Claim</DialogTitle>
            <DialogDescription>
              Confirm approval and provide payment details
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="payment_reference">Payment Reference</Label>
              <Input
                id="payment_reference"
                placeholder="Enter payment reference number"
                value={paymentReference}
                onChange={(e) => setPaymentReference(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="admin_notes">Admin Notes (Optional)</Label>
              <Textarea
                id="admin_notes"
                placeholder="Add any notes about this approval..."
                value={adminNotes}
                onChange={(e) => setAdminNotes(e.target.value)}
                rows={3}
              />
            </div>
            <div className="flex gap-2">
              <Button
                className="flex-1"
                onClick={handleApproveClaim}
                disabled={processing}
              >
                {processing ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Approving...
                  </>
                ) : (
                  <>
                    <CheckCircle className="mr-2 h-4 w-4" />
                    Confirm Approval
                  </>
                )}
              </Button>
              <Button
                variant="outline"
                onClick={() => setShowApproveDialog(false)}
                disabled={processing}
              >
                Cancel
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Reject Dialog */}
      <Dialog open={showRejectDialog} onOpenChange={setShowRejectDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Reject Claim</DialogTitle>
            <DialogDescription>
              Provide a reason for rejecting this claim
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label htmlFor="rejection_reason">Rejection Reason *</Label>
              <Textarea
                id="rejection_reason"
                placeholder="Explain why this claim is being rejected..."
                value={rejectionReason}
                onChange={(e) => setRejectionReason(e.target.value)}
                rows={3}
                required
              />
            </div>
            <div>
              <Label htmlFor="admin_notes_reject">Admin Notes (Optional)</Label>
              <Textarea
                id="admin_notes_reject"
                placeholder="Add any additional notes..."
                value={adminNotes}
                onChange={(e) => setAdminNotes(e.target.value)}
                rows={2}
              />
            </div>
            <div className="flex gap-2">
              <Button
                variant="destructive"
                className="flex-1"
                onClick={handleRejectClaim}
                disabled={processing || !rejectionReason.trim()}
              >
                {processing ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Rejecting...
                  </>
                ) : (
                  <>
                    <XCircle className="mr-2 h-4 w-4" />
                    Confirm Rejection
                  </>
                )}
              </Button>
              <Button
                variant="outline"
                onClick={() => setShowRejectDialog(false)}
                disabled={processing}
              >
                Cancel
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default SpinPrizeClaimsManagement;
