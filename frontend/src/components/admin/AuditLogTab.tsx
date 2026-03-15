/**
 * AuditLogTab
 * Paginated view of admin_activity_logs (audit trail).
 * Pulled from the /admin/audit-logs endpoint.
 */

import React, { useState, useEffect, useCallback } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useToast } from '@/hooks/useToast';
import { ClipboardList, RefreshCw, ChevronLeft, ChevronRight } from 'lucide-react';

import apiClient from '@/lib/api-client';
const PAGE_SIZE = 50;

interface AuditLog {
  id: string;
  admin_user_id?: string;
  action: string;
  entity_type?: string;
  entity_id?: string;
  new_value?: any;
  ip_address?: string;
  created_at: string;
}

function formatDate(s: string): string {
  return new Date(s).toLocaleString('en-NG', {
    day: '2-digit', month: 'short', year: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  });
}

const ACTION_COLORS: Record<string, string> = {
  CREATE: 'bg-green-100 text-green-800',
  UPDATE: 'bg-blue-100 text-blue-800',
  DELETE: 'bg-red-100 text-red-800',
  PATCH:  'bg-yellow-100 text-yellow-800',
};

function actionBadge(action: string) {
  const verb = action.split(':')[0] ?? action;
  const colorClass = ACTION_COLORS[verb] ?? 'bg-gray-100 text-gray-700';
  return (
    <span className={`inline-block px-2 py-0.5 rounded text-xs font-mono font-semibold ${colorClass}`}>
      {action}
    </span>
  );
}

export const AuditLogTab: React.FC = () => {
  const { toast } = useToast();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(false);
  const [actionFilter, setActionFilter] = useState('');

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        limit: String(PAGE_SIZE),
        offset: String(page * PAGE_SIZE),
      });
      if (actionFilter.trim()) params.set('action', actionFilter.trim());

      const res = await apiClient.get<{ success: boolean; data?: { logs: typeof logs; total: number }; message?: string }>(`/admin/audit-logs?${params}`);
      const data = res.data;
      if (data.success) {
        setLogs(data.data?.logs ?? []);
        setTotal(data.data?.total ?? 0);
      } else {
        toast({ title: 'Error', description: data.message ?? 'Failed to load audit logs', variant: 'destructive' });
      }
    } catch {
      toast({ title: 'Error', description: 'Network error loading audit logs', variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  }, [page, actionFilter, toast]);

  useEffect(() => { fetchLogs(); }, [fetchLogs]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <ClipboardList className="w-5 h-5" />
          Audit Log
        </CardTitle>
        <CardDescription>
          All admin mutations — who changed what and when ({total.toLocaleString()} entries)
        </CardDescription>
      </CardHeader>
      <CardContent>
        {/* Toolbar */}
        <div className="flex gap-2 mb-4">
          <Input
            placeholder="Filter by action (e.g. UPDATE:settings)…"
            value={actionFilter}
            onChange={e => { setActionFilter(e.target.value); setPage(0); }}
            className="max-w-xs"
          />
          <Button variant="outline" size="sm" onClick={fetchLogs} disabled={loading}>
            <RefreshCw className={`w-4 h-4 mr-1 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Timestamp</TableHead>
              <TableHead>Action</TableHead>
              <TableHead>Entity</TableHead>
              <TableHead>Entity ID</TableHead>
              <TableHead>Admin ID</TableHead>
              <TableHead>IP</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8 text-gray-400">Loading…</TableCell>
              </TableRow>
            ) : logs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8">
                  <ClipboardList className="w-10 h-10 text-gray-300 mx-auto mb-2" />
                  <p className="text-gray-400">No audit entries yet</p>
                </TableCell>
              </TableRow>
            ) : (
              logs.map(log => (
                <TableRow key={log.id}>
                  <TableCell className="text-xs text-gray-500 whitespace-nowrap">
                    {formatDate(log.created_at)}
                  </TableCell>
                  <TableCell>{actionBadge(log.action)}</TableCell>
                  <TableCell className="text-sm">{log.entity_type ?? '—'}</TableCell>
                  <TableCell>
                    <code className="text-xs bg-gray-100 px-1 rounded">
                      {log.entity_id ? log.entity_id.slice(0, 8) + '…' : '—'}
                    </code>
                  </TableCell>
                  <TableCell>
                    <code className="text-xs bg-gray-100 px-1 rounded">
                      {log.admin_user_id ? log.admin_user_id.slice(0, 8) + '…' : 'system'}
                    </code>
                  </TableCell>
                  <TableCell className="text-xs text-gray-500">{log.ip_address ?? '—'}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        {/* Pagination */}
        <div className="flex items-center justify-between mt-4">
          <p className="text-xs text-gray-400">
            Page {page + 1} of {totalPages} · {total.toLocaleString()} total entries
          </p>
          <div className="flex gap-1">
            <Button
              size="sm" variant="outline"
              onClick={() => setPage(p => Math.max(0, p - 1))}
              disabled={page === 0 || loading}
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <Button
              size="sm" variant="outline"
              onClick={() => setPage(p => Math.min(totalPages - 1, p + 1))}
              disabled={page >= totalPages - 1 || loading}
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default AuditLogTab;
