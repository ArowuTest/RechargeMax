/**
 * UserManagementTab
 * Full user management: view, ban, suspend, activate, tier override.
 * Extracted from ComprehensiveAdminPortal for single-responsibility.
 */

import React, { useState, useEffect, useCallback } from 'react';
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
import { Users, Search, RefreshCw, ShieldBan, ShieldCheck, Edit3, Eye } from 'lucide-react';

const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api/v1';

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

  // Edit dialog state
  const [editUser, setEditUser] = useState<User | null>(null);
  const [editStatus, setEditStatus] = useState('');
  const [editTier, setEditTier] = useState('');
  const [editSaving, setEditSaving] = useState(false);

  const fetchUsers = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch(`${API_BASE}/admin/users/all`, { credentials: 'include' });
      const data = await res.json();
      const list: User[] = Array.isArray(data.data)
        ? data.data
        : Array.isArray(data.data?.users)
          ? data.data.users
          : [];
      setUsers(list);
    } catch {
      toast({ title: 'Error', description: 'Failed to load users', variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  }, [toast]);

  useEffect(() => { fetchUsers(); }, [fetchUsers]);

  useEffect(() => {
    if (!search.trim()) {
      setFiltered(users);
      return;
    }
    const q = search.toLowerCase();
    setFiltered(users.filter(u =>
      u.msisdn?.includes(q) ||
      userDisplayName(u).toLowerCase().includes(q) ||
      u.email?.toLowerCase().includes(q)
    ));
  }, [search, users]);

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
      const statusRes = await fetch(`${API_BASE}/admin/users/${editUser.id}/status`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: editStatus }),
      });
      const statusData = await statusRes.json();
      if (!statusData.success) throw new Error(statusData.message || 'Status update failed');

      // Update loyalty tier via status endpoint (backend accepts loyalty_tier in body)
      if (editTier !== editUser.loyalty_tier) {
        await fetch(`${API_BASE}/admin/users/${editUser.id}/status`, {
          method: 'PUT',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ loyalty_tier: editTier }),
        });
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

  const quickStatus = async (u: User, newStatus: string) => {
    setActionLoading(u.id);
    try {
      const res = await fetch(`${API_BASE}/admin/users/${u.id}/status`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus }),
      });
      const data = await res.json();
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
            <Button variant="outline" size="sm" onClick={fetchUsers} disabled={loading}>
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

          <p className="text-xs text-gray-400 mt-3">
            Showing {filtered.length} of {users.length} users
          </p>
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
    </>
  );
};

export default UserManagementTab;
