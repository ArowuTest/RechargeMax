import React, { useState, useEffect } from 'react';
import { adminApi } from '@/lib/api-client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { useToast } from '@/hooks/use-toast';
import { 
  DollarSign, 
  Plus, 
  Edit, 
  Trash2, 
  Loader2, 
  AlertCircle,
  Shield,
  TrendingUp,
  Clock,
  Calendar,
  CheckCircle,
  XCircle
} from 'lucide-react';

interface TransactionLimit {
  id: string;
  limit_type: string;
  limit_scope: string;
  min_amount: number;
  max_amount: number;
  daily_limit: number | null;
  monthly_limit: number | null;
  is_active: boolean;
  applies_to_user_tier: string | null;
  description: string;
  created_at: string;
  updated_at: string;
}

interface LimitFormData {
  limit_type: string;
  limit_scope: string;
  min_amount: string;
  max_amount: string;
  daily_limit: string;
  monthly_limit: string;
  applies_to_user_tier: string;
  description: string;
}

const TransactionLimitsManagement: React.FC = () => {
  const { toast } = useToast();
  const [limits, setLimits] = useState<TransactionLimit[]>([]);
  const [loading, setLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingLimit, setEditingLimit] = useState<TransactionLimit | null>(null);
  const [formData, setFormData] = useState<LimitFormData>({
    limit_type: 'airtime',
    limit_scope: 'per_transaction',
    min_amount: '100',
    max_amount: '100000',
    daily_limit: '',
    monthly_limit: '',
    applies_to_user_tier: 'all',
    description: ''
  });

  const limitTypes = [
    { value: 'airtime', label: 'Airtime Recharge' },
    { value: 'data', label: 'Data Recharge' },
    { value: 'subscription', label: 'Daily Subscription' },
    { value: 'withdrawal', label: 'Withdrawal' }
  ];

  const limitScopes = [
    { value: 'per_transaction', label: 'Per Transaction' },
    { value: 'daily', label: 'Daily' },
    { value: 'monthly', label: 'Monthly' }
  ];

  const userTiers = [
    { value: 'all', label: 'All Users' },
    { value: 'bronze', label: 'Bronze Tier' },
    { value: 'silver', label: 'Silver Tier' },
    { value: 'gold', label: 'Gold Tier' },
    { value: 'platinum', label: 'Platinum Tier' }
  ];

  useEffect(() => {
    fetchLimits();
  }, []);

  const fetchLimits = async () => {
    try {
      setLoading(true);
      const response = await adminApi.get('/admin/transaction-limits');
      setLimits(response.data.data || []);
    } catch (error) {
      console.error('Failed to fetch limits:', error);
      toast({
        title: 'Error',
        description: 'Failed to load transaction limits',
        variant: 'destructive'
      });
    } finally {
      setLoading(false);
    }
  };

  const handleInputChange = (field: keyof LimitFormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const resetForm = () => {
    setFormData({
      limit_type: 'airtime',
      limit_scope: 'per_transaction',
      min_amount: '100',
      max_amount: '100000',
      daily_limit: '',
      monthly_limit: '',
      applies_to_user_tier: 'all',
      description: ''
    });
    setEditingLimit(null);
  };

  const handleSubmit = async () => {
    try {
      const payload = {
        limit_type: formData.limit_type,
        limit_scope: formData.limit_scope,
        min_amount: parseInt(formData.min_amount) * 100, // Convert to kobo
        max_amount: parseInt(formData.max_amount) * 100, // Convert to kobo
        daily_limit: formData.daily_limit ? parseInt(formData.daily_limit) * 100 : null,
        monthly_limit: formData.monthly_limit ? parseInt(formData.monthly_limit) * 100 : null,
        applies_to_user_tier: formData.applies_to_user_tier === 'all' ? null : formData.applies_to_user_tier,
        description: formData.description
      };

      if (editingLimit) {
        await adminApi.put(`/admin/transaction-limits/${editingLimit.id}`, payload);
        toast({
          title: 'Success',
          description: 'Transaction limit updated successfully'
        });
      } else {
        await adminApi.post('/admin/transaction-limits', payload);
        toast({
          title: 'Success',
          description: 'Transaction limit created successfully'
        });
      }

      setIsDialogOpen(false);
      resetForm();
      fetchLimits();
    } catch (error: any) {
      console.error('Failed to save limit:', error);
      toast({
        title: 'Error',
        description: error.response?.data?.message || 'Failed to save transaction limit',
        variant: 'destructive'
      });
    }
  };

  const handleEdit = (limit: TransactionLimit) => {
    setEditingLimit(limit);
    setFormData({
      limit_type: limit.limit_type,
      limit_scope: limit.limit_scope,
      min_amount: (limit.min_amount / 100).toString(),
      max_amount: (limit.max_amount / 100).toString(),
      daily_limit: limit.daily_limit ? (limit.daily_limit / 100).toString() : '',
      monthly_limit: limit.monthly_limit ? (limit.monthly_limit / 100).toString() : '',
      applies_to_user_tier: limit.applies_to_user_tier || 'all',
      description: limit.description
    });
    setIsDialogOpen(true);
  };

  const handleDelete = async (limitId: string) => {
    if (!confirm('Are you sure you want to deactivate this limit?')) return;

    try {
      await adminApi.delete(`/admin/transaction-limits/${limitId}`);
      toast({
        title: 'Success',
        description: 'Transaction limit deactivated successfully'
      });
      fetchLimits();
    } catch (error) {
      console.error('Failed to delete limit:', error);
      toast({
        title: 'Error',
        description: 'Failed to deactivate transaction limit',
        variant: 'destructive'
      });
    }
  };

  const formatAmount = (kobo: number) => {
    return `₦${(kobo / 100).toLocaleString()}`;
  };

  const getLimitTypeLabel = (type: string) => {
    return limitTypes.find(t => t.value === type)?.label || type;
  };

  const getScopeLabel = (scope: string) => {
    return limitScopes.find(s => s.value === scope)?.label || scope;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Transaction Limits</h2>
          <p className="text-muted-foreground">
            Configure and manage transaction limits for fraud prevention and risk management
          </p>
        </div>
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button onClick={resetForm}>
              <Plus className="mr-2 h-4 w-4" />
              Create Limit
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>
                {editingLimit ? 'Edit Transaction Limit' : 'Create Transaction Limit'}
              </DialogTitle>
              <DialogDescription>
                Configure limits for transaction amounts to prevent fraud and manage risk
              </DialogDescription>
            </DialogHeader>

            <div className="grid gap-4 py-4">
              {/* Limit Type */}
              <div className="grid gap-2">
                <Label htmlFor="limit_type">Limit Type *</Label>
                <Select
                  value={formData.limit_type}
                  onValueChange={(value) => handleInputChange('limit_type', value)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {limitTypes.map(type => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Limit Scope */}
              <div className="grid gap-2">
                <Label htmlFor="limit_scope">Limit Scope *</Label>
                <Select
                  value={formData.limit_scope}
                  onValueChange={(value) => handleInputChange('limit_scope', value)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {limitScopes.map(scope => (
                      <SelectItem key={scope.value} value={scope.value}>
                        {scope.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* User Tier */}
              <div className="grid gap-2">
                <Label htmlFor="applies_to_user_tier">Applies To</Label>
                <Select
                  value={formData.applies_to_user_tier}
                  onValueChange={(value) => handleInputChange('applies_to_user_tier', value)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {userTiers.map(tier => (
                      <SelectItem key={tier.value} value={tier.value}>
                        {tier.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Amount Limits */}
              <div className="grid grid-cols-2 gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="min_amount">Minimum Amount (₦) *</Label>
                  <Input
                    id="min_amount"
                    type="number"
                    value={formData.min_amount}
                    onChange={(e) => handleInputChange('min_amount', e.target.value)}
                    placeholder="100"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="max_amount">Maximum Amount (₦) *</Label>
                  <Input
                    id="max_amount"
                    type="number"
                    value={formData.max_amount}
                    onChange={(e) => handleInputChange('max_amount', e.target.value)}
                    placeholder="100000"
                  />
                </div>
              </div>

              {/* Daily and Monthly Limits */}
              <div className="grid grid-cols-2 gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="daily_limit">Daily Limit (₦)</Label>
                  <Input
                    id="daily_limit"
                    type="number"
                    value={formData.daily_limit}
                    onChange={(e) => handleInputChange('daily_limit', e.target.value)}
                    placeholder="Optional"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="monthly_limit">Monthly Limit (₦)</Label>
                  <Input
                    id="monthly_limit"
                    type="number"
                    value={formData.monthly_limit}
                    onChange={(e) => handleInputChange('monthly_limit', e.target.value)}
                    placeholder="Optional"
                  />
                </div>
              </div>

              {/* Description */}
              <div className="grid gap-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) => handleInputChange('description', e.target.value)}
                  placeholder="Enter a description for this limit"
                  rows={3}
                />
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleSubmit}>
                {editingLimit ? 'Update Limit' : 'Create Limit'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Limits</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{limits.length}</div>
            <p className="text-xs text-muted-foreground">
              Configured transaction limits
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Limits</CardTitle>
            <CheckCircle className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {limits.filter(l => l.is_active).length}
            </div>
            <p className="text-xs text-muted-foreground">
              Currently enforced
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Airtime Limits</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {limits.filter(l => l.limit_type === 'airtime').length}
            </div>
            <p className="text-xs text-muted-foreground">
              Airtime recharge limits
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Data Limits</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {limits.filter(l => l.limit_type === 'data').length}
            </div>
            <p className="text-xs text-muted-foreground">
              Data recharge limits
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Limits Table */}
      <Card>
        <CardHeader>
          <CardTitle>Transaction Limits</CardTitle>
          <CardDescription>
            Manage all configured transaction limits
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : limits.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 text-center">
              <AlertCircle className="h-12 w-12 text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold">No limits configured</h3>
              <p className="text-sm text-muted-foreground">
                Create your first transaction limit to get started
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Type</TableHead>
                  <TableHead>Scope</TableHead>
                  <TableHead>User Tier</TableHead>
                  <TableHead>Min Amount</TableHead>
                  <TableHead>Max Amount</TableHead>
                  <TableHead>Daily Limit</TableHead>
                  <TableHead>Monthly Limit</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {limits.map((limit) => (
                  <TableRow key={limit.id}>
                    <TableCell className="font-medium">
                      {getLimitTypeLabel(limit.limit_type)}
                    </TableCell>
                    <TableCell>{getScopeLabel(limit.limit_scope)}</TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {limit.applies_to_user_tier || 'All Users'}
                      </Badge>
                    </TableCell>
                    <TableCell>{formatAmount(limit.min_amount)}</TableCell>
                    <TableCell>{formatAmount(limit.max_amount)}</TableCell>
                    <TableCell>
                      {limit.daily_limit ? formatAmount(limit.daily_limit) : '-'}
                    </TableCell>
                    <TableCell>
                      {limit.monthly_limit ? formatAmount(limit.monthly_limit) : '-'}
                    </TableCell>
                    <TableCell>
                      {limit.is_active ? (
                        <Badge className="bg-green-600">Active</Badge>
                      ) : (
                        <Badge variant="secondary">Inactive</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEdit(limit)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(limit.id)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default TransactionLimitsManagement;
