import React, { useState, useEffect } from 'react';
import { adminApi } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { 
  Trophy, 
  Calendar, 
  Users, 
  DollarSign, 
  Play, 
  Pause, 
  Settings, 
  RefreshCw,
  Loader2,
  AlertCircle,
  CheckCircle,
  Clock,
  Target,
  Gift,
  Upload,
  XCircle,
  Pencil
} from 'lucide-react';

interface Draw {
  id: string;
  name: string;
  type: 'DAILY' | 'WEEKLY' | 'MONTHLY' | 'SPECIAL';
  status: 'UPCOMING' | 'ACTIVE' | 'DRAWING' | 'COMPLETED' | 'CANCELLED';
  start_time: string;
  end_time: string;
  draw_time?: string;
  prize_pool: number;
  total_entries: number;
  max_entries?: number;
  winners_count: number;
  runner_ups_count?: number;
  created_at: string;
}

interface DrawEntry {
  id: string;
  draw_id: string;
  user_phone: string;
  entries_count: number;
  source_type: 'RECHARGE' | 'SUBSCRIPTION' | 'BONUS';
  created_at: string;
}

interface DrawWinner {
  id: string;
  draw_id: string;
  user_phone: string;
  prize_amount: number;
  position: number;
  claimed: boolean;
  claimed_at?: string;
}

const DrawIntegrationDashboard: React.FC = () => {
  const [draws, setDraws] = useState<Draw[]>([]);
  const [entries, setEntries] = useState<DrawEntry[]>([]);
  const [winners, setWinners] = useState<DrawWinner[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string>('');

  // Prize Tier System state
  const [drawTypes, setDrawTypes] = useState<any[]>([]);
  const [prizeTemplates, setPrizeTemplates] = useState<any[]>([]);
  const [selectedDrawType, setSelectedDrawType] = useState<number | null>(null);
  const [selectedTemplate, setSelectedTemplate] = useState<number | null>(null);
  const [prizeCategories, setPrizeCategories] = useState<any[]>([]);
  const [totalPrizePool, setTotalPrizePool] = useState(0);

  // New draw form state
  const [newDraw, setNewDraw] = useState({
    name: '',
    type: 'DAILY' as const,
    draw_type_id: null as number | null,
    prize_template_id: null as number | null,
    duration_hours: 24
  });
  const [csvFile, setCsvFile] = useState<File | null>(null);

  // Edit draw dialog state
  const [editingDraw, setEditingDraw] = useState<Draw | null>(null);
  const [editDrawForm, setEditDrawForm] = useState({
    name: '',
    status: '' as Draw['status'],
    duration_hours: 24
  });
  const [editLoading, setEditLoading] = useState(false);

  useEffect(() => {
    fetchDrawData();
    fetchDrawTypes();
  }, []);

  const { toast } = useToast();

  const fetchDrawData = async () => {
    try {
      setLoading(true);
      
      // Fetch real data from backend API
      const drawsResponse = await adminApi.draws.getAll();
      setDraws(drawsResponse.data || []);
      
      // Fetch entries and winners for active draws
      // Note: These endpoints may need to be added to the backend
      // For now, initialize as empty arrays
      setEntries([]);
      setWinners([]);
    } catch (error: any) {
      console.error('Failed to fetch draw data:', error);
      toast({
        title: "Error Loading Draws",
        description: error.message || "Failed to load draw data from server",
        variant: "destructive",
      });
      // Set empty arrays on error
      setDraws([]);
      setEntries([]);
      setWinners([]);
    } finally {
      setLoading(false);
    }
  };

  const fetchDrawTypes = async () => {
    try {
      const response = await adminApi.get('/admin/draw-types');
      // adminApi.get returns the full response body: { success, data: [...] }
      setDrawTypes(response.data || []);
    } catch (error: any) {
      console.error('Failed to fetch draw types:', error);
      toast({
        title: "Error",
        description: "Failed to load draw types",
        variant: "destructive",
      });
    }
  };

  const fetchPrizeTemplates = async (drawTypeId: number) => {
    try {
      const response = await adminApi.get(`/admin/prize-templates?draw_type_id=${drawTypeId}`);
      // adminApi.get returns the full response body: { success, data: [...] }
      setPrizeTemplates(response.data || []);
    } catch (error: any) {
      console.error('Failed to fetch prize templates:', error);
      toast({
        title: "Error",
        description: "Failed to load prize templates",
        variant: "destructive",
      });
    }
  };

  const fetchPrizeCategories = async (templateId: number) => {
    try {
      const response = await adminApi.get(`/admin/prize-templates/${templateId}`);
      // adminApi.get returns the full response body: { success, data: {...} }
      const template = response.data;
      setPrizeCategories(template.prize_categories || []);
      
      // Calculate total prize pool
      const total = (template.prize_categories || []).reduce((sum: number, cat: any) => {
        return sum + (cat.prize_amount * cat.winner_count);
      }, 0);
      setTotalPrizePool(total);
    } catch (error: any) {
      console.error('Failed to fetch prize categories:', error);
      toast({
        title: "Error",
        description: "Failed to load prize categories",
        variant: "destructive",
      });
    }
  };

  const handleDrawTypeChange = async (value: string) => {
    const drawTypeId = parseInt(value);
    setSelectedDrawType(drawTypeId);
    setSelectedTemplate(null);
    setPrizeCategories([]);
    setTotalPrizePool(0);
    setNewDraw(prev => ({ ...prev, draw_type_id: drawTypeId, prize_template_id: null }));
    await fetchPrizeTemplates(drawTypeId);
  };

  const handleTemplateChange = async (value: string) => {
    const templateId = parseInt(value);
    setSelectedTemplate(templateId);
    setNewDraw(prev => ({ ...prev, prize_template_id: templateId }));
    await fetchPrizeCategories(templateId);
  };

  const handleCreateDraw = async () => {
    try {
      if (!selectedDrawType || !selectedTemplate) {
        toast({
          title: "Validation Error",
          description: "Please select both draw type and prize template",
          variant: "destructive",
        });
        return;
      }

      setActionLoading('create_draw');
      
      // Create draw via API
      // Only include prize_pool if no template is selected (backend requires either template OR prize_pool > 0, not both)
      const drawData: Record<string, any> = {
        name: newDraw.name,
        type: newDraw.type,
        draw_type_id: selectedDrawType,
        prize_template_id: selectedTemplate,
        duration_hours: newDraw.duration_hours,
        start_time: new Date().toISOString(),
        end_time: new Date(Date.now() + (newDraw.duration_hours * 3600000)).toISOString(),
      };
      // Only add prize_pool if template is not selected and pool is > 0
      if (!selectedTemplate && totalPrizePool > 0) {
        drawData.prize_pool = totalPrizePool;
      }

      await adminApi.draws.create(drawData);
      
      toast({
        title: "Draw Created",
        description: `${newDraw.name} has been created successfully with ${prizeCategories.length} prize categories`,
      });
      
      // Refresh draw list
      await fetchDrawData();
      
      // Reset form
      setNewDraw({
        name: '',
        type: 'DAILY',
        draw_type_id: null,
        prize_template_id: null,
        duration_hours: 24
      });
      setSelectedDrawType(null);
      setSelectedTemplate(null);
      setPrizeCategories([]);
      setTotalPrizePool(0);
      setCsvFile(null);
    } catch (error: any) {
      console.error('Failed to create draw:', error);
      toast({
        title: "Error Creating Draw",
        description: error.message || "Failed to create draw",
        variant: "destructive",
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleDrawAction = async (drawId: string, action: 'start' | 'pause' | 'complete' | 'cancel') => {
    try {
      setActionLoading(`${action}_${drawId}`);
      
      if (action === 'complete') {
        // Execute draw
        await adminApi.draws.execute(drawId);
        toast({
          title: "Draw Executed",
          description: "Draw has been executed and winners selected",
        });
      } else {
        // Update draw status
        const statusMap = {
          start: 'ACTIVE',
          pause: 'UPCOMING',
          cancel: 'CANCELLED'
        };
        await adminApi.draws.update(drawId, { status: statusMap[action as keyof typeof statusMap] });
        toast({
          title: "Draw Updated",
          description: `Draw has been ${action}ed successfully`,
        });
      }
      
      // Refresh draw list
      await fetchDrawData();
    } catch (error: any) {
      console.error(`Failed to ${action} draw:`, error);
      toast({
        title: "Error",
        description: error.message || `Failed to ${action} draw`,
        variant: "destructive",
      });
    } finally {
      setActionLoading('');
    }
  };

  const handleOpenEditDraw = (draw: Draw) => {
    setEditingDraw(draw);
    // Calculate remaining hours from now to end_time
    const remainingMs = new Date(draw.end_time).getTime() - Date.now();
    const remainingHours = Math.max(1, Math.round(remainingMs / 3600000));
    setEditDrawForm({
      name: draw.name,
      status: draw.status,
      duration_hours: remainingHours
    });
  };

  const handleUpdateDraw = async () => {
    if (!editingDraw) return;
    try {
      setEditLoading(true);
      const updatePayload: Record<string, any> = {
        name: editDrawForm.name,
        status: editDrawForm.status,
        end_time: new Date(Date.now() + editDrawForm.duration_hours * 3600000).toISOString()
      };
      await adminApi.draws.update(editingDraw.id, updatePayload);
      toast({
        title: 'Draw Updated',
        description: `${editDrawForm.name} has been updated successfully`,
      });
      setEditingDraw(null);
      await fetchDrawData();
    } catch (error: any) {
      console.error('Failed to update draw:', error);
      toast({
        title: 'Error Updating Draw',
        description: error.message || 'Failed to update draw',
        variant: 'destructive',
      });
    } finally {
      setEditLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ACTIVE':
        return 'bg-green-100 text-green-800';
      case 'UPCOMING':
        return 'bg-blue-100 text-blue-800';
      case 'DRAWING': // falls through
        return 'bg-purple-100 text-purple-800';
      case 'COMPLETED':
        return 'bg-gray-100 text-gray-800';
      case 'CANCELLED':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ACTIVE':
        return <CheckCircle className="w-4 h-4" />;
      case 'UPCOMING':
        return <Clock className="w-4 h-4" />;
      case 'DRAWING': // falls through
        return <Target className="w-4 h-4" />;
      case 'COMPLETED':
        return <Trophy className="w-4 h-4" />;
      case 'CANCELLED':
        return <AlertCircle className="w-4 h-4" />;
      default:
        return <Clock className="w-4 h-4" />;
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
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  if (loading) {
    return (
      <Card>
        <CardContent className="p-8 text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4" />
          <p>Loading draw integration dashboard...</p>
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
            <Trophy className="w-6 h-6" />
            Draw Engine Integration
          </h2>
          <p className="text-gray-600">
            Manage prize draws and lottery integrations
          </p>
        </div>
        <Button onClick={fetchDrawData} variant="outline">
          <RefreshCw className="w-4 h-4 mr-2" />
          Refresh
        </Button>
      </div>

      <Tabs defaultValue="active" className="space-y-4">
        <TabsList>
          <TabsTrigger value="active">Active Draws</TabsTrigger>
          <TabsTrigger value="create">Create Draw</TabsTrigger>
          <TabsTrigger value="entries">Entries</TabsTrigger>
          <TabsTrigger value="winners">Winners</TabsTrigger>
        </TabsList>

        {/* Active Draws Tab */}
        <TabsContent value="active">
          <Card>
            <CardHeader>
              <CardTitle>Current Draws</CardTitle>
              <CardDescription>
                Manage active and upcoming prize draws
              </CardDescription>
            </CardHeader>
            <CardContent>
              {draws.length > 0 ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Draw Name</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Prize Pool</TableHead>
                      <TableHead>Entries</TableHead>
                      <TableHead>End Time</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {draws.map((draw) => (
                      <TableRow key={draw.id}>
                        <TableCell className="font-medium">{draw.name}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{draw.type}</Badge>
                        </TableCell>
                        <TableCell>
                          <Badge className={getStatusColor(draw.status)}>
                            {getStatusIcon(draw.status)}
                            <span className="ml-1">{draw.status}</span>
                          </Badge>
                        </TableCell>
                        <TableCell>{formatCurrency(draw.prize_pool)}</TableCell>
                        <TableCell>
                          {draw.total_entries}
                          {draw.max_entries && ` / ${draw.max_entries}`}
                        </TableCell>
                        <TableCell>{formatDate(draw.end_time)}</TableCell>
                        <TableCell>
                          <div className="flex gap-1">
                            {draw.status === 'UPCOMING' && (
                              <Button
                                size="sm"
                                onClick={() => handleDrawAction(draw.id, 'start')}
                                disabled={actionLoading === `start_${draw.id}`}
                              >
                                {actionLoading === `start_${draw.id}` ? (
                                  <Loader2 className="w-3 h-3 animate-spin" />
                                ) : (
                                  <Play className="w-3 h-3" />
                                )}
                              </Button>
                            )}
                            {draw.status === 'ACTIVE' && (
                              <>
                                <Button
                                  size="sm"
                                  variant="outline"
                                  onClick={() => handleDrawAction(draw.id, 'pause')}
                                  disabled={actionLoading === `pause_${draw.id}`}
                                >
                                  <Pause className="w-3 h-3" />
                                </Button>
                                <Button
                                  size="sm"
                                  onClick={() => handleDrawAction(draw.id, 'complete')}
                                  disabled={actionLoading === `complete_${draw.id}`}
                                >
                                  <Trophy className="w-3 h-3" />
                                </Button>
                              </>
                            )}
                            <Button
                              size="sm"
                              variant="outline"
                              title="Edit Draw"
                              onClick={() => handleOpenEditDraw(draw)}
                            >
                              <Pencil className="w-3 h-3" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <div className="text-center py-8">
                  <Trophy className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-500">No draws found</p>
                  <p className="text-sm text-gray-400">Create a new draw to get started</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Create Draw Tab */}
        <TabsContent value="create">
          <Card>
            <CardHeader>
              <CardTitle>Create New Draw</CardTitle>
              <CardDescription>
                Set up a new prize draw with custom parameters
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <div>
                    <Label htmlFor="draw_name">Draw Name</Label>
                    <Input
                      id="draw_name"
                      value={newDraw.name}
                      onChange={(e) => setNewDraw(prev => ({ ...prev, name: e.target.value }))}
                      placeholder="e.g., Weekend Mega Draw"
                    />
                  </div>

                  <div>
                    <Label htmlFor="draw_type_selector">Draw Type</Label>
                    <Select 
                      value={selectedDrawType?.toString() || ''} 
                      onValueChange={handleDrawTypeChange}
                    >
                      <SelectTrigger id="draw_type_selector">
                        <SelectValue placeholder="Select draw type..." />
                      </SelectTrigger>
                      <SelectContent>
                        {drawTypes.map((type) => (
                          <SelectItem key={type.id} value={type.id.toString()}>
                            {type.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {selectedDrawType && (
                    <div>
                      <Label htmlFor="prize_template">Prize Template</Label>
                      <Select 
                        value={selectedTemplate?.toString() || ''} 
                        onValueChange={handleTemplateChange}
                      >
                        <SelectTrigger id="prize_template">
                          <SelectValue placeholder="Select prize template..." />
                        </SelectTrigger>
                        <SelectContent>
                          {prizeTemplates.map((template) => (
                            <SelectItem key={template.id} value={template.id.toString()}>
                              {template.name}
                              {template.is_default && (
                                <Badge variant="outline" className="ml-2">Default</Badge>
                              )}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </div>

                <div className="space-y-4">
                  {prizeCategories.length > 0 && (
                    <div>
                      <Label>Prize Categories</Label>
                      <div className="border rounded-lg p-4 space-y-2 max-h-64 overflow-y-auto">
                        {prizeCategories.map((category: any, index: number) => (
                          <div key={category.id} className="flex justify-between items-center p-2 bg-muted rounded">
                            <div>
                              <span className="font-medium">{category.category_name}</span>
                              <span className="text-sm text-muted-foreground ml-2">
                                ({category.winner_count} winner{category.winner_count > 1 ? 's' : ''})
                              </span>
                            </div>
                            <div className="text-right">
                              <div className="font-bold">₦{category.prize_amount.toLocaleString()}</div>
                              <div className="text-xs text-muted-foreground">
                                {category.runner_up_count} runner-up{category.runner_up_count > 1 ? 's' : ''}
                              </div>
                            </div>
                          </div>
                        ))}
                        
                        <div className="border-t pt-2 mt-2">
                          <div className="flex justify-between items-center font-bold">
                            <span>Total Prize Pool:</span>
                            <span className="text-lg">₦{totalPrizePool.toLocaleString()}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}

                  <div>
                    <Label htmlFor="duration">Duration (Hours)</Label>
                    <Input
                      id="duration"
                      type="number"
                      min="1"
                      max="168"
                      value={newDraw.duration_hours}
                      onChange={(e) => setNewDraw(prev => ({ ...prev, duration_hours: parseInt(e.target.value) || 24 }))}
                    />
                  </div>
                </div>
              </div>

              {/* CSV Upload Section */}
              <div className="mt-6 p-4 border rounded-lg bg-muted/50">
                <div className="flex items-center gap-2 mb-3">
                  <Upload className="h-4 w-4" />
                  <Label className="text-sm font-semibold">Manual Entry Upload (Optional)</Label>
                </div>
                <p className="text-xs text-muted-foreground mb-3">
                  Upload a CSV file with MSISDN and Points. Format: <code className="bg-background px-1 py-0.5 rounded">MSISDN,Points</code>
                </p>
                <div className="flex items-center gap-2">
                  <Input
                    type="file"
                    accept=".csv"
                    onChange={(e) => setCsvFile(e.target.files?.[0] || null)}
                    className="flex-1"
                  />
                  {csvFile && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setCsvFile(null)}
                    >
                      <XCircle className="h-4 w-4" />
                    </Button>
                  )}
                </div>
                {csvFile && (
                  <p className="text-xs text-green-600 mt-2 flex items-center gap-1">
                    <CheckCircle className="h-3 w-3" />
                    {csvFile.name} selected
                  </p>
                )}
              </div>

              {/* Preview */}
              {newDraw.name && totalPrizePool > 0 && prizeCategories.length > 0 && (
                <Alert className="mt-6">
                  <Gift className="h-4 w-4" />
                  <AlertDescription>
                    <strong>Preview:</strong> {newDraw.name} - {formatCurrency(totalPrizePool)} total prize pool 
                    across {prizeCategories.length} prize categor{prizeCategories.length > 1 ? 'ies' : 'y'}, 
                    running for {newDraw.duration_hours} hours
                  </AlertDescription>
                </Alert>
              )}

              <Button 
                onClick={handleCreateDraw}
                disabled={!newDraw.name || !selectedDrawType || !selectedTemplate || actionLoading === 'create_draw'}
                className="w-full mt-6"
              >
                {actionLoading === 'create_draw' ? (
                  <Loader2 className="w-4 h-4 animate-spin mr-2" />
                ) : (
                  <Trophy className="w-4 h-4 mr-2" />
                )}
                Create Draw
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Entries Tab */}
        <TabsContent value="entries">
          <Card>
            <CardHeader>
              <CardTitle>Draw Entries</CardTitle>
              <CardDescription>
                View all entries across active draws
              </CardDescription>
            </CardHeader>
            <CardContent>
              {entries.length > 0 ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>User</TableHead>
                      <TableHead>Draw</TableHead>
                      <TableHead>Entries</TableHead>
                      <TableHead>Source</TableHead>
                      <TableHead>Date</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {entries.map((entry) => {
                      const draw = draws.find(d => d.id === entry.draw_id);
                      return (
                        <TableRow key={entry.id}>
                          <TableCell>
                            <code className="text-sm bg-gray-100 px-2 py-1 rounded">
                              {entry.user_phone}
                            </code>
                          </TableCell>
                          <TableCell>{draw?.name || 'Unknown Draw'}</TableCell>
                          <TableCell>
                            <Badge variant="secondary">
                              {entry.entries_count} entries
                            </Badge>
                          </TableCell>
                          <TableCell>
                            <Badge variant="outline">
                              {entry.source_type}
                            </Badge>
                          </TableCell>
                          <TableCell>{formatDate(entry.created_at)}</TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              ) : (
                <div className="text-center py-8">
                  <Users className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-500">No entries found</p>
                  <p className="text-sm text-gray-400">Entries will appear when users participate in draws</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Winners Tab */}
        <TabsContent value="winners">
          <Card>
            <CardHeader>
              <CardTitle>Draw Winners</CardTitle>
              <CardDescription>
                View winners from completed draws
              </CardDescription>
            </CardHeader>
            <CardContent>
              {winners.length > 0 ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Winner</TableHead>
                      <TableHead>Draw</TableHead>
                      <TableHead>Position</TableHead>
                      <TableHead>Prize</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Date</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {winners.map((winner) => {
                      const draw = draws.find(d => d.id === winner.draw_id);
                      return (
                        <TableRow key={winner.id}>
                          <TableCell>
                            <code className="text-sm bg-gray-100 px-2 py-1 rounded">
                              {winner.user_phone}
                            </code>
                          </TableCell>
                          <TableCell>{draw?.name || 'Unknown Draw'}</TableCell>
                          <TableCell>
                            <Badge variant="outline">
                              #{winner.position}
                            </Badge>
                          </TableCell>
                          <TableCell className="font-medium">
                            {formatCurrency(winner.prize_amount)}
                          </TableCell>
                          <TableCell>
                            <Badge variant={winner.claimed ? 'default' : 'destructive'}>
                              {winner.claimed ? 'Claimed' : 'Pending'}
                            </Badge>
                          </TableCell>
                          <TableCell>
                            {winner.claimed_at ? formatDate(winner.claimed_at) : '-'}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              ) : (
                <div className="text-center py-8">
                  <Trophy className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-500">No winners yet</p>
                  <p className="text-sm text-gray-400">Winners will appear after draws are completed</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
      {/* Edit Draw Dialog */}
      {editingDraw && (
        <Dialog open={!!editingDraw} onOpenChange={(open) => { if (!open) setEditingDraw(null); }}>
          <DialogContent className="max-w-md">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <Settings className="w-5 h-5" />
                Edit Draw
              </DialogTitle>
            </DialogHeader>
            <div className="space-y-4 py-2">
              <div>
                <Label htmlFor="edit_draw_name">Draw Name</Label>
                <Input
                  id="edit_draw_name"
                  value={editDrawForm.name}
                  onChange={(e) => setEditDrawForm(prev => ({ ...prev, name: e.target.value }))}
                  placeholder="Draw name"
                />
              </div>
              <div>
                <Label htmlFor="edit_draw_status">Status</Label>
                <Select
                  value={editDrawForm.status}
                  onValueChange={(val) => setEditDrawForm(prev => ({ ...prev, status: val as Draw['status'] }))}
                >
                  <SelectTrigger id="edit_draw_status">
                    <SelectValue placeholder="Select status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="UPCOMING">UPCOMING</SelectItem>
                    <SelectItem value="ACTIVE">ACTIVE</SelectItem>
                    <SelectItem value="CANCELLED">CANCELLED</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <Label htmlFor="edit_draw_duration">Extend Duration (Hours from now)</Label>
                <Input
                  id="edit_draw_duration"
                  type="number"
                  min="1"
                  max="720"
                  value={editDrawForm.duration_hours}
                  onChange={(e) => setEditDrawForm(prev => ({ ...prev, duration_hours: parseInt(e.target.value) || 24 }))}
                />
                <p className="text-xs text-muted-foreground mt-1">
                  New end time: {new Date(Date.now() + editDrawForm.duration_hours * 3600000).toLocaleString()}
                </p>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setEditingDraw(null)} disabled={editLoading}>
                Cancel
              </Button>
              <Button onClick={handleUpdateDraw} disabled={!editDrawForm.name || editLoading}>
                {editLoading ? (
                  <Loader2 className="w-4 h-4 animate-spin mr-2" />
                ) : (
                  <Settings className="w-4 h-4 mr-2" />
                )}
                Update Draw
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}
    </div>
  );
};

export default DrawIntegrationDashboard;