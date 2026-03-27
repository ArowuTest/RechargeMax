/**
 * UserManagementTab
 * Full user management: view, ban, suspend, activate, tier override.
 * Extracted from ComprehensiveAdminPortal for single-responsibility.
 */

import React, { useState, useEffect, useCallback } from 'react';
import { apiClient } from '@/lib/api-client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useToast } from '@/hooks/useToast';
import { Users, Search, RefreshCw, ShieldBan, ShieldCheck, Edit3, Coins } from 'lucide-react';



interface User {
  id: string;
  msisdn: string;
  full_name?: string;
  first_name?: string;
  last_name?: string;
  email?: string;
  is_active: boolean;
  status?: string;
  loyalty_tier?: string;
  total_points?: number;
  points_balance?: number;
  total_spent?: number;
  transaction_count?: number;
  created_at?: string;
}

const TIER_OPTIONS = ['BRONZE', 'SILVER', 'GOLD', 'PLATINUM'];
const STATUS_OPTIONS = [
  { value: 'active',    label: 'Active' },
  { value: 'suspended', label: 'Suspend' },
  { value: 'banned',    label: 'Ban' },
];

const TIER_COLORS: Record<string, string> = {
  BRONZE:   'bg-amber-100 text-amber-800',
  SILVER:   'bg-gray-100 text-gray-700',
  GOLD:     'bg-yellow-100 text-yellow-800',
  PLATINUM: 'bg-purple-100 text-purple-800',
};

const STATUS_COLORS: Record<string, string> = {
  active:    'bg-green-100 text-green-800',
  suspended: 'bg-yellow-100 text-yellow-800',
  banned:    'bg-red-100 text-red-800',
};

function userDisplayName(u: User): string {
  return [u.first_name, u.last_name].filter(Boolean).join(' ')
    || u.full_name
    || `User …${u.msisdn?.slice(-4) ?? ''}`;
}

function userStatus(u: User): string {
  if (u.status) return u.status;
  return u.is_active ? 'active' : 'suspended';
}

function formatCurrency(kobo: number): string {
  return `₦${(kobo / 100).toLocaleString('en-NG', { minimumFractionDigits: 2 })}`;
}

function formatDate(s?: string): string {
  if (!s) return '—';
  return new Date(s).toLocaleDateString('en-NG', { day: '2-digit', month: 'short', year: 'numeric' });
}

export const UserManagementTab: React.FC = () => {
  const { toast } = useToast();
  const [users, setUsers] = useState<User[]>([]);
  const [filtered, setFiltered] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState('');
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalUsers, setTotalUsers] = useState(0);
  const PAGE_SIZE = 20;

  // Edit dialog state
  const [editUser, setEditUser] = useState<User | null>(null);
  const [editStatus, setEditStatus] = useState('');
  const [editTier, setEditTier] = useState('');
  const [editSaving, setEditSaving] = useState(false);

  // Adjust Points dialog state
  const [adjustUser, setAdjustUser] = useState<User | null>(null);
  const [adjustPoints, setAdjustPoints] = useState('');
  const [adjustReason, setAdjustReason] = useState('');
  const [adjustDesc, setAdjustDesc] = useState('');
  const [adjustSaving, setAdjustSaving] = useState(false);

  const fetchUsers = useCallback(async (page = 1, q = '') => {
    setLoading(true);
    try {
      const params: Record<string, any> = { page, per_page: PAGE_SIZE };
      if (q.trim()) params.search = q.trim();
      const res = await apiClient.get('/admin/users/all', { params });
      const data = res.data;
      const list: User[] = Array.isArray(data.data)
        ? data.data
        : Array.isArray(data.data?.users)
          ? data.data.users
          : [];
      setUsers(list);
      setFiltered(list);
      setTotalUsers(data.pagination?.total ?? list.length);
    } catch {
      toast({ title: 'Error', description: 'Failed to load users', variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  }, [toast]);

  useEffect(() => { fetchUsers(1, ''); }, [fetchUsers]);

  // debounce search → server-side
  useEffect(() => {
    const t = setTimeout(() => {
      setCurrentPage(1);
      fetchUsers(1, search);
    }, 400);
    return () => clearTimeout(t);
  }, [search]);

  useEffect(() => {
    fetchUsers(currentPage, search);
  }, [currentPage]);

  // keep filtered in sync (used by older render code below)
  useEffect(() => {
    setFiltered(users);
  }, [users]);

  // client-side filter is now disabled — search is server-side

  const openEdit = (u: User) => {
    setEditUser(u);
    setEditStatus(userStatus(u));
    setEditTier(u.loyalty_tier ?? 'BRONZE');
  };

  const saveEdit = async () => {
    if (!editUser) return;
    setEditSaving(true);
    try {
      // Update status
      const statusRes = await apiClient.put(`/admin/users/${editUser.id}/status`, { status: editStatus });
      const statusData = statusRes.data;
      if (!statusData.success) throw new Error(statusData.message || 'Status update failed');

      // Update loyalty tier via status endpoint (backend accepts loyalty_tier in body)
      if (editTier !== editUser.loyalty_tier) {
        await apiClient.put(`/admin/users/${editUser.id}/status`, { loyalty_tier: editTier });
      }

      toast({ title: 'Updated', description: `${userDisplayName(editUser)} updated successfully` });
      setEditUser(null);
      fetchUsers();
    } catch (e: any) {
      toast({ title: 'Error', description: e.message ?? 'Update failed', variant: 'destructive' });
    } finally {
      setEditSaving(false);
    }
  };

  const handleAdjustPoints = async () => {
    if (!adjustUser) return;
    const pts = parseInt(adjustPoints, 10);
    if (isNaN(pts) || pts === 0) {
      toast({ title: 'Validation', description: 'Enter a non-zero integer (positive to add, negative to deduct)', variant: 'destructive' });
      return;
    }
    if (!adjustReason.trim()) {
      toast({ title: 'Validation', description: 'Reason is required', variant: 'destructive' });
      return;
    }
    setAdjustSaving(true);
    try {
      const res = await apiClient.post(`/admin/users/${adjustUser.id}/adjust-points`, {
        points: pts,
        reason: adjustReason.trim(),
        description: adjustDesc.trim(),
      });
      const data = res.data;
      if (!data.success) throw new Error(data.message ?? data.error ?? 'Failed');
      toast({ title: 'Points adjusted', description: `${pts > 0 ? '+' : ''}${pts} points for ${adjustUser.msisdn}` });
      setAdjustUser(null);
      setAdjustPoints('');
      setAdjustReason('');
      setAdjustDesc('');
      fetchUsers(currentPage, search);
    } catch (e: any) {
      toast({ title: 'Error', description: e.message ?? 'Failed to adjust points', variant: 'destructive' });
    } finally {
      setAdjustSaving(false);
    }
  };

  const quickStatus = async (u: User, newStatus: string) => {
    setActionLoading(u.id);
    try {
      const res = await apiClient.put(`/admin/users/${u.id}/status`, { status: newStatus });
      const data = res.data;
      if (!data.success) throw new Error(data.message);
      toast({ title: 'Done', description: `User ${newStatus}` });
      fetchUsers();
    } catch (e: any) {
      toast({ title: 'Error', description: e.message ?? 'Action failed', variant: 'destructive' });
    } finally {
      setActionLoading(null);
    }
  };

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="w-5 h-5" />
            User Management
          </CardTitle>
          <CardDescription>
            View, search, ban/suspend/activate users, and override loyalty tiers
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* Toolbar */}
          <div className="flex gap-2 mb-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <Input
                placeholder="Search by phone, name or email…"
                value={search}
                onChange={e => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>
            <Button variant="outline" size="sm" onClick={() => fetchUsers()} disabled={loading}>
              <RefreshCw className={`w-4 h-4 mr-1 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>User</TableHead>
                <TableHead>Phone</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Tier</TableHead>
                <TableHead>Points</TableHead>
                <TableHead>Spent</TableHead>
                <TableHead>Joined</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-8 text-gray-400">Loading…</TableCell>
                </TableRow>
              ) : filtered.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-8">
                    <Users className="w-10 h-10 text-gray-300 mx-auto mb-2" />
                    <p className="text-gray-400">{search ? 'No users match your search' : 'No users yet'}</p>
                  </TableCell>
                </TableRow>
              ) : (
                filtered.map(u => {
                  const st = userStatus(u);
                  const tier = u.loyalty_tier ?? 'BRONZE';
                  return (
                    <TableRow key={u.id}>
                      <TableCell>
                        <div>
                          <p className="font-medium text-sm">{userDisplayName(u)}</p>
                          <p className="text-xs text-gray-400">{u.email ?? '—'}</p>
                        </div>
                      </TableCell>
                      <TableCell>
                        <code className="bg-gray-100 px-1.5 py-0.5 rounded text-xs">{u.msisdn}</code>
                      </TableCell>
                      <TableCell>
                        <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${STATUS_COLORS[st] ?? 'bg-gray-100 text-gray-700'}`}>
                          {st.charAt(0).toUpperCase() + st.slice(1)}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${TIER_COLORS[tier] ?? ''}`}>
                          {tier}
                        </span>
                      </TableCell>
                      <TableCell className="text-sm">{(u.total_points ?? u.points_balance ?? 0).toLocaleString()}</TableCell>
                      <TableCell className="text-sm">{formatCurrency(u.total_spent ?? 0)}</TableCell>
                      <TableCell className="text-sm text-gray-500">{formatDate(u.created_at)}</TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          <Button
                            size="sm" variant="outline" title="Edit user"
                            onClick={() => openEdit(u)}
                          >
                            <Edit3 className="w-3 h-3" />
                          </Button>
                          <Button
                            size="sm" variant="outline" title="Adjust points"
                            onClick={() => { setAdjustUser(u); setAdjustPoints(''); setAdjustReason(''); setAdjustDesc(''); }}
                          >
                            <Coins className="w-3 h-3" />
                          </Button>
                          {st === 'active' ? (
                            <Button
                              size="sm" variant="destructive" title="Suspend user"
                              disabled={actionLoading === u.id}
                              onClick={() => quickStatus(u, 'suspended')}
                            >
                              <ShieldBan className="w-3 h-3" />
                            </Button>
                          ) : (
                            <Button
                              size="sm" variant="default" title="Activate user"
                              disabled={actionLoading === u.id}
                              onClick={() => quickStatus(u, 'active')}
                            >
                              <ShieldCheck className="w-3 h-3" />
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>

          <div className="flex items-center justify-between mt-4">
            <p className="text-xs text-gray-400">
              Showing {filtered.length} of {totalUsers} users (page {currentPage} of {Math.ceil(totalUsers / PAGE_SIZE) || 1})
            </p>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" disabled={currentPage <= 1} onClick={() => setCurrentPage(p => Math.max(1, p - 1))}>
                ← Prev
              </Button>
              <Button variant="outline" size="sm" disabled={currentPage >= (Math.ceil(totalUsers / PAGE_SIZE) || 1)} onClick={() => setCurrentPage(p => p + 1)}>
                Next →
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Edit dialog */}
      <Dialog open={!!editUser} onOpenChange={open => !open && setEditUser(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Edit User</DialogTitle>
            <DialogDescription>
              {editUser ? userDisplayName(editUser) : ''} · {editUser?.msisdn}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div>
              <label className="text-sm font-medium mb-1 block">Account Status</label>
              <Select value={editStatus} onValueChange={setEditStatus}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {STATUS_OPTIONS.map(o => (
                    <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div>
              <label className="text-sm font-medium mb-1 block">Loyalty Tier Override</label>
              <Select value={editTier} onValueChange={setEditTier}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TIER_OPTIONS.map(t => (
                    <SelectItem key={t} value={t}>{t}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <p className="text-xs text-gray-400 mt-1">
                Manually override the tier. It will be re-computed on the next recharge.
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditUser(null)}>Cancel</Button>
            <Button onClick={saveEdit} disabled={editSaving}>
              {editSaving ? 'Saving…' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Adjust Points dialog */}
      <Dialog open={!!adjustUser} onOpenChange={open => !open && setAdjustUser(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Coins className="w-4 h-4" />
              Adjust Points
            </DialogTitle>
            <DialogDescription>
              {adjustUser ? `${userDisplayName(adjustUser)} · ${adjustUser.msisdn}` : ''}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div>
              <label className="text-sm font-medium mb-1 block">Points *</label>
              <Input
                type="number"
                placeholder="+50 to add, -20 to deduct"
                value={adjustPoints}
                onChange={e => setAdjustPoints(e.target.value)}
              />
              <p className="text-xs text-gray-400 mt-1">Use positive numbers to add, negative to deduct.</p>
            </div>
            <div>
              <label className="text-sm font-medium mb-1 block">Reason *</label>
              <Input
                placeholder="e.g. Loyalty bonus, Manual correction"
                value={adjustReason}
                onChange={e => setAdjustReason(e.target.value)}
              />
            </div>
            <div>
              <label className="text-sm font-medium mb-1 block">Description (optional)</label>
              <Input
                placeholder="Additional notes…"
                value={adjustDesc}
                onChange={e => setAdjustDesc(e.target.value)}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setAdjustUser(null)}>Cancel</Button>
            <Button onClick={handleAdjustPoints} disabled={adjustSaving}>
              {adjustSaving ? 'Saving…' : 'Apply Adjustment'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default UserManagementTab;
