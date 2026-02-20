import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '@/components/ui/table';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { 
  Plus, 
  Edit, 
  Trash2, 
  Trophy, 
  TrendingUp, 
  CheckCircle, 
  XCircle,
  AlertTriangle,
  RefreshCw,
  Loader2
} from 'lucide-react';
import { SpinTierDialog } from './SpinTierDialog';

interface SpinTier {
  id: string;
  tier_name: string;
  tier_display_name: string;
  min_daily_amount: number;
  max_daily_amount: number | null;
  spins_per_day: number;
  tier_color: string;
  tier_icon: string;
  tier_badge: string;
  tier_description: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface ValidationError {
  tier_name: string;
  issue: string;
  details: string;
}

const SpinTiersManagement: React.FC = () => {
  const [tiers, setTiers] = useState<SpinTier[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([]);
  
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedTier, setSelectedTier] = useState<SpinTier | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [tierToDelete, setTierToDelete] = useState<SpinTier | null>(null);

  useEffect(() => {
    loadTiers();
  }, []);

  const loadTiers = async () => {
    try {
      setLoading(true);
      setError('');
      
      const token = localStorage.getItem('adminToken');
      const response = await fetch('/api/admin/spin-tiers', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('Failed to load spin tiers');
      }

      const data = await response.json();
      setTiers(data.tiers || []);
      
      // Validate configuration
      await validateConfiguration();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load spin tiers');
    } finally {
      setLoading(false);
    }
  };

  const validateConfiguration = async () => {
    try {
      const token = localStorage.getItem('adminToken');
      const response = await fetch('/api/admin/spin-tiers/validate', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (response.ok) {
        const data = await response.json();
        setValidationErrors(data.errors || []);
      }
    } catch (err) {
      console.error('Validation check failed:', err);
    }
  };

  const handleCreate = () => {
    setSelectedTier(null);
    setDialogOpen(true);
  };

  const handleEdit = (tier: SpinTier) => {
    setSelectedTier(tier);
    setDialogOpen(true);
  };

  const handleDelete = (tier: SpinTier) => {
    setTierToDelete(tier);
    setDeleteDialogOpen(true);
  };

  const confirmDelete = async () => {
    if (!tierToDelete) return;

    try {
      setSaving(true);
      setError('');
      
      const token = localStorage.getItem('adminToken');
      const response = await fetch(`/api/admin/spin-tiers/${tierToDelete.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to delete tier');
      }

      setSuccess(`Tier "${tierToDelete.tier_display_name}" deleted successfully`);
      setDeleteDialogOpen(false);
      setTierToDelete(null);
      await loadTiers();
      
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete tier');
    } finally {
      setSaving(false);
    }
  };

  const handleSave = async (tierData: Omit<SpinTier, 'id' | 'created_at' | 'updated_at'>) => {
    try {
      setSaving(true);
      setError('');
      
      const token = localStorage.getItem('adminToken');
      const url = selectedTier 
        ? `/api/admin/spin-tiers/${selectedTier.id}`
        : '/api/admin/spin-tiers';
      
      const method = selectedTier ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(tierData)
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to save tier');
      }

      setSuccess(selectedTier ? 'Tier updated successfully' : 'Tier created successfully');
      setDialogOpen(false);
      await loadTiers();
      
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save tier');
      throw err;
    } finally {
      setSaving(false);
    }
  };

  const formatAmount = (amount: number): string => {
    return `₦${(amount / 100).toLocaleString()}`;
  };

  const getTierBadgeColor = (tierName: string): string => {
    const colors: Record<string, string> = {
      'BRONZE': 'bg-orange-100 text-orange-800 border-orange-300',
      'SILVER': 'bg-gray-100 text-gray-800 border-gray-300',
      'GOLD': 'bg-yellow-100 text-yellow-800 border-yellow-300',
      'PLATINUM': 'bg-blue-100 text-blue-800 border-blue-300',
      'DIAMOND': 'bg-purple-100 text-purple-800 border-purple-300'
    };
    return colors[tierName] || 'bg-gray-100 text-gray-800 border-gray-300';
  };

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
          <span className="ml-3 text-gray-600">Loading spin tiers...</span>
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
            <Trophy className="w-7 h-7 text-blue-600" />
            Spin Tiers Management
          </h2>
          <p className="text-gray-600 mt-1">
            Configure daily recharge tiers and spin rewards
          </p>
        </div>
        <Button onClick={handleCreate} className="gap-2">
          <Plus className="w-4 h-4" />
          Add New Tier
        </Button>
      </div>

      {/* Success Message */}
      {success && (
        <Alert className="bg-green-50 border-green-200">
          <CheckCircle className="w-4 h-4 text-green-600" />
          <AlertDescription className="text-green-800">
            {success}
          </AlertDescription>
        </Alert>
      )}

      {/* Error Message */}
      {error && (
        <Alert className="bg-red-50 border-red-200">
          <XCircle className="w-4 h-4 text-red-600" />
          <AlertDescription className="text-red-800">
            {error}
          </AlertDescription>
        </Alert>
      )}

      {/* Validation Errors */}
      {validationErrors.length > 0 && (
        <Alert className="bg-yellow-50 border-yellow-200">
          <AlertTriangle className="w-4 h-4 text-yellow-600" />
          <AlertDescription className="text-yellow-800">
            <div className="font-semibold mb-2">Configuration Issues Detected:</div>
            <ul className="list-disc list-inside space-y-1">
              {validationErrors.map((err, idx) => (
                <li key={idx}>
                  <strong>{err.tier_name}:</strong> {err.issue} - {err.details}
                </li>
              ))}
            </ul>
          </AlertDescription>
        </Alert>
      )}

      {/* Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Total Tiers</p>
                <p className="text-2xl font-bold">{tiers.length}</p>
              </div>
              <Trophy className="w-8 h-8 text-blue-600" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Active Tiers</p>
                <p className="text-2xl font-bold text-green-600">
                  {tiers.filter(t => t.is_active).length}
                </p>
              </div>
              <CheckCircle className="w-8 h-8 text-green-600" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Max Spins</p>
                <p className="text-2xl font-bold text-purple-600">
                  {Math.max(...tiers.map(t => t.spins_per_day), 0)}
                </p>
              </div>
              <TrendingUp className="w-8 h-8 text-purple-600" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">Issues</p>
                <p className={`text-2xl font-bold ${validationErrors.length > 0 ? 'text-red-600' : 'text-green-600'}`}>
                  {validationErrors.length}
                </p>
              </div>
              {validationErrors.length > 0 ? (
                <AlertTriangle className="w-8 h-8 text-red-600" />
              ) : (
                <CheckCircle className="w-8 h-8 text-green-600" />
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tiers Table */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Configured Tiers</CardTitle>
              <CardDescription>
                Manage spin tier thresholds and rewards
              </CardDescription>
            </div>
            <Button 
              variant="outline" 
              size="sm" 
              onClick={loadTiers}
              className="gap-2"
            >
              <RefreshCw className="w-4 h-4" />
              Refresh
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {tiers.length === 0 ? (
            <div className="text-center py-12">
              <Trophy className="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600 mb-4">No spin tiers configured yet</p>
              <Button onClick={handleCreate} className="gap-2">
                <Plus className="w-4 h-4" />
                Create First Tier
              </Button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Tier</TableHead>
                    <TableHead>Daily Amount Range</TableHead>
                    <TableHead className="text-center">Spins/Day</TableHead>
                    <TableHead className="text-center">Sort Order</TableHead>
                    <TableHead className="text-center">Status</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tiers
                    .sort((a, b) => a.sort_order - b.sort_order)
                    .map((tier) => (
                      <TableRow key={tier.id}>
                        <TableCell>
                          <div className="flex items-center gap-3">
                            <div 
                              className="w-10 h-10 rounded-full flex items-center justify-center text-white font-bold"
                              style={{ backgroundColor: tier.tier_color }}
                            >
                              {tier.tier_icon}
                            </div>
                            <div>
                              <div className="font-semibold">{tier.tier_display_name}</div>
                              <Badge 
                                variant="outline" 
                                className={`text-xs ${getTierBadgeColor(tier.tier_name)}`}
                              >
                                {tier.tier_badge}
                              </Badge>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            <div className="font-medium">
                              {formatAmount(tier.min_daily_amount)} - {tier.max_daily_amount ? formatAmount(tier.max_daily_amount) : 'Unlimited'}
                            </div>
                            <div className="text-gray-500 text-xs">
                              {tier.tier_description}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell className="text-center">
                          <Badge variant="secondary" className="font-semibold">
                            {tier.spins_per_day} {tier.spins_per_day === 1 ? 'spin' : 'spins'}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-center">
                          <span className="text-gray-600">{tier.sort_order}</span>
                        </TableCell>
                        <TableCell className="text-center">
                          {tier.is_active ? (
                            <Badge className="bg-green-100 text-green-800 border-green-300">
                              <CheckCircle className="w-3 h-3 mr-1" />
                              Active
                            </Badge>
                          ) : (
                            <Badge variant="outline" className="bg-gray-100 text-gray-600">
                              <XCircle className="w-3 h-3 mr-1" />
                              Inactive
                            </Badge>
                          )}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleEdit(tier)}
                              className="gap-1"
                            >
                              <Edit className="w-4 h-4" />
                              Edit
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleDelete(tier)}
                              className="gap-1 text-red-600 hover:text-red-700 hover:bg-red-50"
                            >
                              <Trash2 className="w-4 h-4" />
                              Delete
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Tier Dialog */}
      <SpinTierDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        tier={selectedTier}
        existingTiers={tiers}
        onSave={handleSave}
        loading={saving}
      />

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Spin Tier</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the <strong>{tierToDelete?.tier_display_name}</strong> tier?
              This action will soft-delete the tier (mark as inactive) and cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={saving}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              disabled={saving}
              className="bg-red-600 hover:bg-red-700"
            >
              {saving ? (
                <>
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete Tier'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};

export default SpinTiersManagement;
