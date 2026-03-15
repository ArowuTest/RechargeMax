/**
 * Subscription Tier Management Component
 * Full CRUD for managing subscription tiers (entry bundles)
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
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
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/hooks/useToast';
import { Loader2, Plus, Edit, Trash2, ArrowUp, ArrowDown } from 'lucide-react';
import {
  subscriptionTierApi,
  subscriptionPricingApi,
  type SubscriptionTier,
  type SubscriptionPricing,
} from '@/lib/api-client-extensions';

export default function SubscriptionTierManagement() {
  const { toast } = useToast();
  const [tiers, setTiers] = useState<SubscriptionTier[]>([]);
  const [currentPricing, setCurrentPricing] = useState<SubscriptionPricing | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState('');
  const [showTierDialog, setShowTierDialog] = useState(false);
  const [editingTier, setEditingTier] = useState<SubscriptionTier | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    entries_count: 1,
    description: '',
    is_active: true,
    display_order: 0,
  });
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [tiersResponse, pricingResponse] = await Promise.all([
        subscriptionTierApi.getAll(),
        subscriptionPricingApi.getCurrent(),
      ]);

      if (tiersResponse.success && tiersResponse.data) {
        // Sort by display_order
        const sortedTiers = tiersResponse.data.sort((a, b) => a.display_order - b.display_order);
        setTiers(sortedTiers);
      }

      if (pricingResponse.success && pricingResponse.data) {
        setCurrentPricing(pricingResponse.data);
      }
    } catch (error) {
      console.error('Failed to fetch subscription data:', error);
      toast({
        title: 'Error',
        description: 'Failed to load subscription tiers',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (tier?: SubscriptionTier) => {
    if (tier) {
      setEditingTier(tier);
      setFormData({
        name: tier.name,
        entries_count: tier.entries_count,
        description: tier.description || '',
        is_active: tier.is_active,
        display_order: tier.display_order,
      });
    } else {
      setEditingTier(null);
      setFormData({
        name: '',
        entries_count: 1,
        description: '',
        is_active: true,
        display_order: tiers.length,
      });
    }
    setFormErrors({});
    setShowTierDialog(true);
  };

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};

    if (!formData.name.trim()) {
      errors.name = 'Tier name is required';
    }

    if (formData.entries_count < 1) {
      errors.entries_count = 'Entries count must be at least 1';
    }

    if (formData.display_order < 0) {
      errors.display_order = 'Display order must be 0 or greater';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSave = async () => {
    if (!validateForm()) return;

    setActionLoading(editingTier ? `update-${editingTier.id}` : 'create');
    try {
      if (editingTier) {
        const response = await subscriptionTierApi.update(editingTier.id, formData);
        if (response.success) {
          toast({
            title: 'Success',
            description: 'Subscription tier updated successfully',
          });
          await fetchData();
          setShowTierDialog(false);
        } else {
          throw new Error(response.message || 'Failed to update tier');
        }
      } else {
        const response = await subscriptionTierApi.create(formData);
        if (response.success) {
          toast({
            title: 'Success',
            description: 'Subscription tier created successfully',
          });
          await fetchData();
          setShowTierDialog(false);
        } else {
          throw new Error(response.message || 'Failed to create tier');
        }
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Operation failed',
        variant: 'destructive',
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this tier?')) return;

    setActionLoading(`delete-${id}`);
    try {
      const response = await subscriptionTierApi.delete(id);
      if (response.success) {
        toast({
          title: 'Success',
          description: 'Subscription tier deleted successfully',
        });
        await fetchData();
      } else {
        throw new Error(response.message || 'Failed to delete tier');
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Delete failed',
        variant: 'destructive',
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleMoveOrder = async (tier: SubscriptionTier, direction: 'up' | 'down') => {
    const currentIndex = tiers.findIndex(t => t.id === tier.id);
    if (
      (direction === 'up' && currentIndex === 0) ||
      (direction === 'down' && currentIndex === tiers.length - 1)
    ) {
      return;
    }

    const newOrder = direction === 'up' ? tier.display_order - 1 : tier.display_order + 1;
    setActionLoading(`move-${tier.id}`);
    
    try {
      const response = await subscriptionTierApi.update(tier.id, {
        ...tier,
        display_order: newOrder,
      });
      
      if (response.success) {
        await fetchData();
      } else {
        throw new Error(response.message || 'Failed to update order');
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to update order',
        variant: 'destructive',
      });
    } finally {
      setActionLoading('');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Subscription Tiers</CardTitle>
              <CardDescription>
                Manage subscription tiers (entry bundles) for daily draws
              </CardDescription>
            </div>
            <Button onClick={() => handleOpenDialog()}>
              <Plus className="mr-2 h-4 w-4" />
              Add Tier
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {currentPricing && (
            <div className="mb-4 p-4 bg-blue-50 rounded-lg">
              <p className="text-sm font-medium">
                Current Price: ₦{(currentPricing.price_per_entry / 100).toFixed(2)} per entry
              </p>
            </div>
          )}

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Order</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Entries</TableHead>
                <TableHead>Price</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tiers.map((tier, index) => (
                <TableRow key={tier.id}>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={index === 0 || actionLoading === `move-${tier.id}`}
                        onClick={() => handleMoveOrder(tier, 'up')}
                      >
                        <ArrowUp className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={index === tiers.length - 1 || actionLoading === `move-${tier.id}`}
                        onClick={() => handleMoveOrder(tier, 'down')}
                      >
                        <ArrowDown className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                  <TableCell className="font-medium">{tier.name}</TableCell>
                  <TableCell>{tier.entries_count}</TableCell>
                  <TableCell>
                    {currentPricing
                      ? `₦${((tier.entries_count * currentPricing.price_per_entry) / 100).toFixed(2)}`
                      : 'N/A'}
                  </TableCell>
                  <TableCell className="max-w-xs truncate">{tier.description}</TableCell>
                  <TableCell>
                    <Badge variant={tier.is_active ? 'default' : 'secondary'}>
                      {tier.is_active ? 'Active' : 'Inactive'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleOpenDialog(tier)}
                        disabled={!!actionLoading}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDelete(tier.id)}
                        disabled={actionLoading === `delete-${tier.id}`}
                      >
                        {actionLoading === `delete-${tier.id}` ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Trash2 className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={showTierDialog} onOpenChange={setShowTierDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingTier ? 'Edit Tier' : 'Create New Tier'}</DialogTitle>
            <DialogDescription>
              Configure the subscription tier details
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">Tier Name</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g., Basic, Premium, VIP"
              />
              {formErrors.name && (
                <p className="text-sm text-red-500">{formErrors.name}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="entries_count">Number of Entries</Label>
              <Input
                id="entries_count"
                type="number"
                min="1"
                value={formData.entries_count}
                onChange={(e) =>
                  setFormData({ ...formData, entries_count: parseInt(e.target.value) || 1 })
                }
              />
              {formErrors.entries_count && (
                <p className="text-sm text-red-500">{formErrors.entries_count}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Describe this tier..."
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="display_order">Display Order</Label>
              <Input
                id="display_order"
                type="number"
                min="0"
                value={formData.display_order}
                onChange={(e) =>
                  setFormData({ ...formData, display_order: parseInt(e.target.value) || 0 })
                }
              />
              {formErrors.display_order && (
                <p className="text-sm text-red-500">{formErrors.display_order}</p>
              )}
            </div>

            <div className="flex items-center space-x-2">
              <Switch
                id="is_active"
                checked={formData.is_active}
                onCheckedChange={(checked) => setFormData({ ...formData, is_active: checked })}
              />
              <Label htmlFor="is_active">Active</Label>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowTierDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={!!actionLoading}>
              {actionLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                'Save'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
