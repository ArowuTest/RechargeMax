import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Textarea } from '@/components/ui/textarea';
import { useToast } from '@/hooks/use-toast';
import { affiliateManagementApi } from '@/lib/api-client-extensions';
import { 
  Users, 
  TrendingUp, 
  DollarSign, 
  Award,
  UserCheck,
  UserX,
  Eye,
  Edit,
  Trash2,
  Plus,
  Download,
  Upload,
  RefreshCw,
  Loader2,
  AlertCircle,
  CheckCircle,
  Clock,
  Star,
  Target,
  BarChart3
} from 'lucide-react';

interface StrategicAffiliate {
  id: string;
  user_id: string;
  full_name: string;
  email: string;
  phone_number: string;
  affiliate_code: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | 'SUSPENDED';
  tier: 'BRONZE' | 'SILVER' | 'GOLD' | 'PLATINUM' | 'DIAMOND';
  commission_rate: number;
  total_referrals: number;
  active_referrals: number;
  total_commission: number;
  pending_commission: number;
  paid_commission: number;
  last_activity: string;
  created_at: string;
  approved_at?: string;
  notes?: string;
}

interface AffiliateStats {
  total_affiliates: number;
  pending_approvals: number;
  active_affiliates: number;
  suspended_affiliates: number;
  total_commission_paid: number;
  pending_commission: number;
  total_referrals: number;
  conversion_rate: number;
}

interface CommissionTier {
  tier: string;
  min_referrals: number;
  commission_rate: number;
  bonus_threshold: number;
  bonus_amount: number;
}

interface StrategicAffiliateAdminDashboardProps {
  sessionToken?: string;
}

const COMMISSION_TIERS: CommissionTier[] = [
  { tier: 'BRONZE', min_referrals: 0, commission_rate: 5, bonus_threshold: 10, bonus_amount: 1000 },
  { tier: 'SILVER', min_referrals: 25, commission_rate: 7.5, bonus_threshold: 25, bonus_amount: 2500 },
  { tier: 'GOLD', min_referrals: 50, commission_rate: 10, bonus_threshold: 50, bonus_amount: 5000 },
  { tier: 'PLATINUM', min_referrals: 100, commission_rate: 12.5, bonus_threshold: 100, bonus_amount: 10000 },
  { tier: 'DIAMOND', min_referrals: 250, commission_rate: 15, bonus_threshold: 250, bonus_amount: 25000 }
];

const StrategicAffiliateAdminDashboard: React.FC<StrategicAffiliateAdminDashboardProps> = ({ sessionToken }) => {
  const { toast } = useToast();
  const [affiliates, setAffiliates] = useState<StrategicAffiliate[]>([]);
  const [stats, setStats] = useState<AffiliateStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string>('');
  const [selectedAffiliate, setSelectedAffiliate] = useState<StrategicAffiliate | null>(null);
  const [showDetailsDialog, setShowDetailsDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [filterStatus, setFilterStatus] = useState<string>('ALL');
  const [filterTier, setFilterTier] = useState<string>('ALL');
  const [searchQuery, setSearchQuery] = useState('');

  // Edit form state
  const [editForm, setEditForm] = useState({
    tier: '',
    commission_rate: 0,
    status: '',
    notes: ''
  });

  useEffect(() => {
    fetchAffiliateData();
  }, []);

  const fetchAffiliateData = async () => {
    try {
      setLoading(true);
      
      const response = await affiliateManagementApi.getAll({ page: 1, limit: 100 });

      if (response.success) {
        setAffiliates((response.data || []) as StrategicAffiliate[]);
        setStats((response as any).stats || {
          total_affiliates: 0,
          pending_approvals: 0,
          active_affiliates: 0,
          suspended_affiliates: 0,
          total_commission_paid: 0,
          pending_commission: 0,
          total_referrals: 0,
          conversion_rate: 0
        });
      } else {
        throw new Error(response.error);
      }
    } catch (error) {
      console.error('Failed to fetch affiliate data:', error);
      toast({
        title: "Error",
        description: "Failed to load affiliate data",
        variant: "destructive"
      });
    } finally {
      setLoading(false);
    }
  };

  const handleAffiliateAction = async (affiliateId: string, action: string, data?: any) => {
    try {
      setActionLoading(`${action}_${affiliateId}`);
      
      let response;
      if (action === 'approve_affiliate') {
        response = await affiliateManagementApi.approve(affiliateId);
      } else if (action === 'reject_affiliate') {
        response = await affiliateManagementApi.reject(affiliateId, data?.reason || 'No reason provided');
      } else if (action === 'update_affiliate') {
        response = await affiliateManagementApi.update(affiliateId, data?.updates || {});
      } else {
        throw new Error('Unknown action');
      }

      if (response.success) {
        await fetchAffiliateData();
        toast({
          title: "Success",
          description: response.message || "Action completed successfully",
        });
        
        if (showEditDialog) {
          setShowEditDialog(false);
        }
      } else {
        throw new Error(response.error);
      }
    } catch (error) {
      console.error(`Failed to ${action} affiliate:`, error);
      toast({
        title: "Action Failed",
        description: error instanceof Error ? error.message : "Failed to complete action",
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleEditAffiliate = async () => {
    if (!selectedAffiliate) return;
    
    await handleAffiliateAction(selectedAffiliate.id, 'update_affiliate', {
      updates: editForm
    });
  };

  const handleExportData = async () => {
    try {
      setActionLoading('export');
      
      const response = await affiliateManagementApi.getAll({ 
        page: 1, 
        limit: 1000,
        status: filterStatus !== 'ALL' ? filterStatus : undefined
      });

      if (response.success && (response as any).export_url) {
        // Create download link
        const link = document.createElement('a');
        link.href = (response as any).export_url;
        link.download = `affiliates_export_${new Date().toISOString().split('T')[0]}.csv`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        toast({
          title: "Export Complete",
          description: "Affiliate data has been exported successfully",
        });
      } else {
        throw new Error('error' in response ? response.error : 'Export failed');
      }
    } catch (error) {
      console.error('Failed to export data:', error);
      toast({
        title: "Export Failed",
        description: "Failed to export affiliate data",
        variant: "destructive"
      });
    } finally {
      setActionLoading('');
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'APPROVED':
        return 'bg-green-100 text-green-800';
      case 'PENDING':
        return 'bg-yellow-100 text-yellow-800';
      case 'REJECTED':
        return 'bg-red-100 text-red-800';
      case 'SUSPENDED':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getTierColor = (tier: string) => {
    switch (tier) {
      case 'DIAMOND':
        return 'bg-purple-100 text-purple-800';
      case 'PLATINUM':
        return 'bg-gray-100 text-gray-800';
      case 'GOLD':
        return 'bg-yellow-100 text-yellow-800';
      case 'SILVER':
        return 'bg-blue-100 text-blue-800';
      case 'BRONZE':
        return 'bg-orange-100 text-orange-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-NG', {
      style: 'currency',
      currency: 'NGN',
      minimumFractionDigits: 0
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  };

  const filteredAffiliates = affiliates.filter(affiliate => {
    const matchesStatus = filterStatus === 'ALL' || affiliate.status === filterStatus;
    const matchesTier = filterTier === 'ALL' || affiliate.tier === filterTier;
    const matchesSearch = !searchQuery || 
      affiliate.full_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      affiliate.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
      affiliate.affiliate_code.toLowerCase().includes(searchQuery.toLowerCase());
    
    return matchesStatus && matchesTier && matchesSearch;
  });

  if (loading) {
    return (
      <Card>
        <CardContent className="p-8 text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4" />
          <p>Loading strategic affiliate dashboard...</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold flex items-center gap-2">
            <Award className="w-6 h-6" />
            Strategic Affiliate Management
          </h2>
          <p className="text-gray-600">
            Manage high-value affiliate partnerships and commission structures
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleExportData}
            disabled={actionLoading === 'export'}
          >
            {actionLoading === 'export' ? (
              <Loader2 className="w-4 h-4 animate-spin mr-2" />
            ) : (
              <Download className="w-4 h-4 mr-2" />
            )}
            Export Data
          </Button>
          <Button variant="outline" onClick={fetchAffiliateData}>
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Total Affiliates</p>
                  <p className="text-2xl font-bold">{stats.total_affiliates}</p>
                  <p className="text-xs text-blue-600">{stats.pending_approvals} pending</p>
                </div>
                <Users className="w-8 h-8 text-blue-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Total Referrals</p>
                  <p className="text-2xl font-bold">{stats.total_referrals}</p>
                  <p className="text-xs text-green-600">{stats.conversion_rate.toFixed(1)}% conversion</p>
                </div>
                <TrendingUp className="w-8 h-8 text-green-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Commission Paid</p>
                  <p className="text-2xl font-bold">{formatCurrency(stats.total_commission_paid)}</p>
                  <p className="text-xs text-purple-600">{formatCurrency(stats.pending_commission)} pending</p>
                </div>
                <DollarSign className="w-8 h-8 text-purple-600" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-gray-600">Active Affiliates</p>
                  <p className="text-2xl font-bold">{stats.active_affiliates}</p>
                  <p className="text-xs text-orange-600">{stats.suspended_affiliates} suspended</p>
                </div>
                <Award className="w-8 h-8 text-orange-600" />
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      <Tabs defaultValue="affiliates" className="space-y-4">
        <TabsList>
          <TabsTrigger value="affiliates">Affiliates</TabsTrigger>
          <TabsTrigger value="tiers">Commission Tiers</TabsTrigger>
          <TabsTrigger value="analytics">Analytics</TabsTrigger>
        </TabsList>

        {/* Affiliates Tab */}
        <TabsContent value="affiliates">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Strategic Affiliates</CardTitle>
                  <CardDescription>
                    Manage high-value affiliate partnerships
                  </CardDescription>
                </div>
                <div className="flex gap-2">
                  <Input
                    placeholder="Search affiliates..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="w-64"
                  />
                  <Select value={filterStatus} onValueChange={setFilterStatus}>
                    <SelectTrigger className="w-32">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="ALL">All Status</SelectItem>
                      <SelectItem value="PENDING">Pending</SelectItem>
                      <SelectItem value="APPROVED">Approved</SelectItem>
                      <SelectItem value="REJECTED">Rejected</SelectItem>
                      <SelectItem value="SUSPENDED">Suspended</SelectItem>
                    </SelectContent>
                  </Select>
                  <Select value={filterTier} onValueChange={setFilterTier}>
                    <SelectTrigger className="w-32">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="ALL">All Tiers</SelectItem>
                      <SelectItem value="BRONZE">Bronze</SelectItem>
                      <SelectItem value="SILVER">Silver</SelectItem>
                      <SelectItem value="GOLD">Gold</SelectItem>
                      <SelectItem value="PLATINUM">Platinum</SelectItem>
                      <SelectItem value="DIAMOND">Diamond</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {filteredAffiliates.length > 0 ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Affiliate</TableHead>
                      <TableHead>Code</TableHead>
                      <TableHead>Tier</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Referrals</TableHead>
                      <TableHead>Commission</TableHead>
                      <TableHead>Last Activity</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredAffiliates.map((affiliate) => (
                      <TableRow key={affiliate.id}>
                        <TableCell>
                          <div>
                            <div className="font-medium">{affiliate.full_name}</div>
                            <div className="text-sm text-gray-500">{affiliate.email}</div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <code className="bg-gray-100 px-2 py-1 rounded text-sm">
                            {affiliate.affiliate_code}
                          </code>
                        </TableCell>
                        <TableCell>
                          <Badge className={getTierColor(affiliate.tier)}>
                            {affiliate.tier}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className={getStatusColor(affiliate.status)}>
                            {affiliate.status}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            <div>{affiliate.total_referrals} total</div>
                            <div className="text-gray-500">{affiliate.active_referrals} active</div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            <div className="font-medium">{formatCurrency(affiliate.total_commission)}</div>
                            <div className="text-gray-500">{affiliate.commission_rate}% rate</div>
                          </div>
                        </TableCell>
                        <TableCell>{formatDate(affiliate.last_activity)}</TableCell>
                        <TableCell>
                          <div className="flex gap-1">
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => {
                                setSelectedAffiliate(affiliate);
                                setShowDetailsDialog(true);
                              }}
                            >
                              <Eye className="w-3 h-3" />
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => {
                                setSelectedAffiliate(affiliate);
                                setEditForm({
                                  tier: affiliate.tier,
                                  commission_rate: affiliate.commission_rate,
                                  status: affiliate.status,
                                  notes: affiliate.notes || ''
                                });
                                setShowEditDialog(true);
                              }}
                            >
                              <Edit className="w-3 h-3" />
                            </Button>
                            {affiliate.status === 'PENDING' && (
                              <>
                                <Button
                                  size="sm"
                                  onClick={() => handleAffiliateAction(affiliate.id, 'approve_affiliate')}
                                  disabled={actionLoading === `approve_affiliate_${affiliate.id}`}
                                >
                                  {actionLoading === `approve_affiliate_${affiliate.id}` ? (
                                    <Loader2 className="w-3 h-3 animate-spin" />
                                  ) : (
                                    <UserCheck className="w-3 h-3" />
                                  )}
                                </Button>
                                <Button
                                  size="sm"
                                  variant="destructive"
                                  onClick={() => handleAffiliateAction(affiliate.id, 'reject_affiliate')}
                                  disabled={actionLoading === `reject_affiliate_${affiliate.id}`}
                                >
                                  <UserX className="w-3 h-3" />
                                </Button>
                              </>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <div className="text-center py-8">
                  <Users className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-500">No affiliates found</p>
                  <p className="text-sm text-gray-400">
                    {searchQuery || filterStatus !== 'ALL' || filterTier !== 'ALL' 
                      ? 'Try adjusting your filters' 
                      : 'Strategic affiliate applications will appear here'}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Commission Tiers Tab */}
        <TabsContent value="tiers">
          <Card>
            <CardHeader>
              <CardTitle>Commission Tier Structure</CardTitle>
              <CardDescription>
                Tiered commission rates based on performance metrics
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {COMMISSION_TIERS.map((tier) => (
                  <div key={tier.tier} className="p-4 border rounded-lg">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <Badge className={getTierColor(tier.tier)}>
                          {tier.tier}
                        </Badge>
                        <div>
                          <div className="font-medium">
                            {tier.commission_rate}% Commission Rate
                          </div>
                          <div className="text-sm text-gray-500">
                            Minimum {tier.min_referrals} referrals required
                          </div>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="font-medium">
                          {formatCurrency(tier.bonus_amount)} Bonus
                        </div>
                        <div className="text-sm text-gray-500">
                          At {tier.bonus_threshold} referrals
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Analytics Tab */}
        <TabsContent value="analytics">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <BarChart3 className="w-5 h-5" />
                Affiliate Analytics
              </CardTitle>
              <CardDescription>
                Performance metrics and insights
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-center py-8">
                <BarChart3 className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <p className="text-gray-500">Analytics dashboard</p>
                <p className="text-sm text-gray-400">Detailed performance metrics and charts</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Affiliate Details Dialog */}
      <Dialog open={showDetailsDialog} onOpenChange={setShowDetailsDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Affiliate Details</DialogTitle>
            <DialogDescription>
              Comprehensive information about {selectedAffiliate?.full_name}
            </DialogDescription>
          </DialogHeader>
          {selectedAffiliate && (
            <div className="space-y-6">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-sm font-medium text-gray-600">Full Name</Label>
                  <p className="font-medium">{selectedAffiliate.full_name}</p>
                </div>
                <div>
                  <Label className="text-sm font-medium text-gray-600">Email</Label>
                  <p>{selectedAffiliate.email}</p>
                </div>
                <div>
                  <Label className="text-sm font-medium text-gray-600">Phone</Label>
                  <p>{selectedAffiliate.phone_number}</p>
                </div>
                <div>
                  <Label className="text-sm font-medium text-gray-600">Affiliate Code</Label>
                  <code className="bg-gray-100 px-2 py-1 rounded">
                    {selectedAffiliate.affiliate_code}
                  </code>
                </div>
                <div>
                  <Label className="text-sm font-medium text-gray-600">Status</Label>
                  <Badge className={getStatusColor(selectedAffiliate.status)}>
                    {selectedAffiliate.status}
                  </Badge>
                </div>
                <div>
                  <Label className="text-sm font-medium text-gray-600">Tier</Label>
                  <Badge className={getTierColor(selectedAffiliate.tier)}>
                    {selectedAffiliate.tier}
                  </Badge>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-4">
                <div className="text-center p-4 bg-blue-50 rounded-lg">
                  <div className="text-2xl font-bold text-blue-600">
                    {selectedAffiliate.total_referrals}
                  </div>
                  <div className="text-sm text-blue-600">Total Referrals</div>
                </div>
                <div className="text-center p-4 bg-green-50 rounded-lg">
                  <div className="text-2xl font-bold text-green-600">
                    {formatCurrency(selectedAffiliate.total_commission)}
                  </div>
                  <div className="text-sm text-green-600">Total Commission</div>
                </div>
                <div className="text-center p-4 bg-purple-50 rounded-lg">
                  <div className="text-2xl font-bold text-purple-600">
                    {selectedAffiliate.commission_rate}%
                  </div>
                  <div className="text-sm text-purple-600">Commission Rate</div>
                </div>
              </div>

              {selectedAffiliate.notes && (
                <div>
                  <Label className="text-sm font-medium text-gray-600">Notes</Label>
                  <p className="mt-1 p-3 bg-gray-50 rounded-lg text-sm">
                    {selectedAffiliate.notes}
                  </p>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Edit Affiliate Dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Affiliate</DialogTitle>
            <DialogDescription>
              Update affiliate settings and commission structure
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label htmlFor="edit_tier">Tier</Label>
              <Select 
                value={editForm.tier} 
                onValueChange={(value) => setEditForm(prev => ({ ...prev, tier: value }))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="BRONZE">Bronze</SelectItem>
                  <SelectItem value="SILVER">Silver</SelectItem>
                  <SelectItem value="GOLD">Gold</SelectItem>
                  <SelectItem value="PLATINUM">Platinum</SelectItem>
                  <SelectItem value="DIAMOND">Diamond</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="edit_commission">Commission Rate (%)</Label>
              <Input
                id="edit_commission"
                type="number"
                min="0"
                max="50"
                step="0.5"
                value={editForm.commission_rate}
                onChange={(e) => setEditForm(prev => ({ 
                  ...prev, 
                  commission_rate: parseFloat(e.target.value) || 0 
                }))}
              />
            </div>

            <div>
              <Label htmlFor="edit_status">Status</Label>
              <Select 
                value={editForm.status} 
                onValueChange={(value) => setEditForm(prev => ({ ...prev, status: value }))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="PENDING">Pending</SelectItem>
                  <SelectItem value="APPROVED">Approved</SelectItem>
                  <SelectItem value="REJECTED">Rejected</SelectItem>
                  <SelectItem value="SUSPENDED">Suspended</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div>
              <Label htmlFor="edit_notes">Notes</Label>
              <Textarea
                id="edit_notes"
                value={editForm.notes}
                onChange={(e) => setEditForm(prev => ({ ...prev, notes: e.target.value }))}
                placeholder="Add notes about this affiliate..."
                rows={3}
              />
            </div>

            <div className="flex gap-2 pt-4">
              <Button
                onClick={handleEditAffiliate}
                disabled={actionLoading.includes('update_affiliate')}
                className="flex-1"
              >
                {actionLoading.includes('update_affiliate') ? (
                  <Loader2 className="w-4 h-4 animate-spin mr-2" />
                ) : null}
                Save Changes
              </Button>
              <Button
                variant="outline"
                onClick={() => setShowEditDialog(false)}
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

export default StrategicAffiliateAdminDashboard;