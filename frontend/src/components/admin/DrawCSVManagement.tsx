/**
 * Draw CSV Management Component
 * Enterprise-grade CSV export/import for draw entries and winners
 * 
 * Features:
 * - Export draw entries with points aggregation from all sources
 * - Import winners CSV with comprehensive validation
 * - Export history tracking with audit trail
 * - File validation and error reporting
 * - Progress indicators for large operations
 * - Retry mechanism for failed operations
 */

import React, { useState, useEffect, useRef } from 'react';
import { adminApi } from '@/lib/api-client';
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
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Checkbox } from '@/components/ui/checkbox';
import { useToast } from '@/hooks/use-toast';
import {
  Loader2,
  Download,
  Upload,
  FileText,
  AlertCircle,
  CheckCircle2,
  XCircle,
  Calendar,
  Users,
  TrendingUp,
  RefreshCw,
  Info,
} from 'lucide-react';
import {
  drawCSVApi,
  type DrawExportRequest,
  type DrawExportHistory,
} from '@/lib/api-client-extensions';

interface Draw {
  id: string;
  name: string;
  draw_date: string;
  status: string;
  total_prizes: number;
}

interface ExportProgress {
  status: 'idle' | 'preparing' | 'exporting' | 'complete' | 'error';
  progress: number;
  message: string;
  totalMSISDNs?: number;
  totalPoints?: number;
  fileUrl?: string;
}

interface ImportProgress {
  status: 'idle' | 'validating' | 'importing' | 'complete' | 'error';
  progress: number;
  message: string;
  totalWinners?: number;
  totalRunnersUp?: number;
  errors?: string[];
}

interface ValidationError {
  row: number;
  field: string;
  message: string;
}

export default function DrawCSVManagement() {
  const { toast } = useToast();
  const fileInputRef = useRef<HTMLInputElement>(null);

  // State
  const [draws, setDraws] = useState<Draw[]>([]);
  const [exportHistory, setExportHistory] = useState<DrawExportHistory[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedDraw, setSelectedDraw] = useState<string>('');
  const [showExportDialog, setShowExportDialog] = useState(false);
  const [showImportDialog, setShowImportDialog] = useState(false);
  const [showHistoryDialog, setShowHistoryDialog] = useState(false);

  // Export state
  const [exportConfig, setExportConfig] = useState<{
    start_date: string;
    end_date: string;
    include_subscription_points: boolean;
    include_ussd_points: boolean;
    include_wheel_points: boolean;
  }>({
    start_date: '',
    end_date: '',
    include_subscription_points: true,
    include_ussd_points: true,
    include_wheel_points: true,
  });
  const [exportProgress, setExportProgress] = useState<ExportProgress>({
    status: 'idle',
    progress: 0,
    message: '',
  });

  // Import state
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [importProgress, setImportProgress] = useState<ImportProgress>({
    status: 'idle',
    progress: 0,
    message: '',
  });
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([]);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      // Fetch draws from actual API
      const drawsResponse = await adminApi.draws.getAll();
      if (drawsResponse.success && drawsResponse.data) {
        // Transform backend draw format to component format
        const transformedDraws: Draw[] = drawsResponse.data.map((draw: any) => ({
          id: draw.id,
          name: draw.name,
          draw_date: draw.draw_time || draw.end_time,
          status: draw.status === 'COMPLETED' ? 'completed' : 'pending',
          total_prizes: draw.winners_count || 0,
        }));
        setDraws(transformedDraws);
      } else {
        setDraws([]);
      }

      // Fetch export history
      const historyResponse = await drawCSVApi.getExportHistory();
      if (historyResponse.success && historyResponse.data) {
        setExportHistory(historyResponse.data);
      }
    } catch (error) {
      console.error('Failed to fetch draw data:', error);
      toast({
        title: 'Error',
        description: 'Failed to load draw data',
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleOpenExportDialog = () => {
    if (!selectedDraw) {
      toast({
        title: 'No Draw Selected',
        description: 'Please select a draw to export',
        variant: 'destructive',
      });
      return;
    }

    const draw = draws.find((d) => d.id === selectedDraw);
    if (draw) {
      // Set default date range to draw date
      const drawDate = new Date(draw.draw_date);
      const endDate = new Date(drawDate);
      const startDate = new Date(drawDate);
      startDate.setDate(startDate.getDate() - 30); // Default to 30 days before draw

      setExportConfig({
        ...exportConfig,
        start_date: startDate.toISOString().split('T')[0] || '',
        end_date: endDate.toISOString().split('T')[0] || '',
      });
    }

    setExportProgress({
      status: 'idle',
      progress: 0,
      message: '',
    });
    setShowExportDialog(true);
  };

  const validateExportConfig = (): boolean => {
    if (!exportConfig.start_date || !exportConfig.end_date) {
      toast({
        title: 'Invalid Date Range',
        description: 'Please select both start and end dates',
        variant: 'destructive',
      });
      return false;
    }

    const start = new Date(exportConfig.start_date);
    const end = new Date(exportConfig.end_date);

    if (start > end) {
      toast({
        title: 'Invalid Date Range',
        description: 'Start date must be before end date',
        variant: 'destructive',
      });
      return false;
    }

    const daysDiff = (end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24);
    if (daysDiff > 365) {
      toast({
        title: 'Date Range Too Large',
        description: 'Please select a date range of 365 days or less',
        variant: 'destructive',
      });
      return false;
    }

    if (!exportConfig.include_subscription_points && 
        !exportConfig.include_ussd_points && 
        !exportConfig.include_wheel_points) {
      toast({
        title: 'No Points Sources Selected',
        description: 'Please select at least one points source to include',
        variant: 'destructive',
      });
      return false;
    }

    return true;
  };

  const handleExportCSV = async () => {
    if (!validateExportConfig()) return;

    setExportProgress({
      status: 'preparing',
      progress: 10,
      message: 'Preparing export...',
    });

    try {
      // Simulate preparation phase
      await new Promise((resolve) => setTimeout(resolve, 500));

      setExportProgress({
        status: 'exporting',
        progress: 30,
        message: 'Aggregating points from all sources...',
      });

      const exportData: DrawExportRequest = {
        draw_id: selectedDraw,
        start_date: exportConfig.start_date,
        end_date: exportConfig.end_date,
        include_subscription_points: exportConfig.include_subscription_points,
        include_ussd_points: exportConfig.include_ussd_points,
        include_wheel_points: exportConfig.include_wheel_points,
      };

      const response = await drawCSVApi.exportCSV(exportData);

      if (response.success && response.data) {
        setExportProgress({
          status: 'complete',
          progress: 100,
          message: 'Export completed successfully!',
          totalMSISDNs: response.data.total_msisdns,
          fileUrl: response.data.file_url,
        });

        toast({
          title: 'Export Successful',
          description: `Exported ${response.data.total_msisdns} MSISDNs`,
        });

        // Refresh export history
        await fetchData();
      }
    } catch (error: any) {
      console.error('Export failed:', error);
      setExportProgress({
        status: 'error',
        progress: 0,
        message: error.response?.data?.error || 'Export failed. Please try again.',
      });

      toast({
        title: 'Export Failed',
        description: error.response?.data?.error || 'Failed to export draw entries',
        variant: 'destructive',
      });
    }
  };

  const handleDownloadCSV = async (fileUrl: string) => {
    try {
      const blob = await drawCSVApi.downloadCSV(fileUrl);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `draw_entries_${selectedDraw}_${Date.now()}.csv`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      toast({
        title: 'Download Started',
        description: 'CSV file is being downloaded',
      });
    } catch (error) {
      toast({
        title: 'Download Failed',
        description: 'Failed to download CSV file',
        variant: 'destructive',
      });
    }
  };

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.name.endsWith('.csv')) {
      toast({
        title: 'Invalid File Type',
        description: 'Please select a CSV file',
        variant: 'destructive',
      });
      return;
    }

    // Validate file size (max 10MB)
    if (file.size > 10 * 1024 * 1024) {
      toast({
        title: 'File Too Large',
        description: 'Please select a file smaller than 10MB',
        variant: 'destructive',
      });
      return;
    }

    setSelectedFile(file);
    setValidationErrors([]);
    setImportProgress({
      status: 'idle',
      progress: 0,
      message: '',
    });
  };

  const validateCSVStructure = async (file: File): Promise<ValidationError[]> => {
    return new Promise((resolve) => {
      const reader = new FileReader();
      const errors: ValidationError[] = [];

      reader.onload = (e) => {
        const text = e.target?.result as string;
        const lines = text.split('\n').filter((line) => line.trim());

        if (lines.length < 2) {
          errors.push({
            row: 0,
            field: 'file',
            message: 'CSV file is empty or contains only headers',
          });
          resolve(errors);
          return;
        }

        // Validate header
        const header = lines[0]?.toLowerCase() ?? '';
        const requiredColumns = ['msisdn', 'prize_name', 'prize_type', 'prize_value'];
        const missingColumns = requiredColumns.filter(
          (col) => !header.includes(col)
        );

        if (missingColumns.length > 0) {
          errors.push({
            row: 1,
            field: 'header',
            message: `Missing required columns: ${missingColumns.join(', ')}`,
          });
        }

        // Validate data rows (sample first 100 rows for performance)
        const dataLines = lines.slice(1, Math.min(101, lines.length));
        dataLines.forEach((line, index) => {
          const row = index + 2; // +2 because index starts at 0 and we skip header
          const columns = line.split(',');

          if (columns.length < requiredColumns.length) {
            errors.push({
              row,
              field: 'row',
              message: `Insufficient columns (expected ${requiredColumns.length}, got ${columns.length})`,
            });
            return;
          }

          // Validate MSISDN format (Nigerian format: 234XXXXXXXXXX or 0XXXXXXXXXX)
          const msisdn = columns[0]?.trim() ?? '';
          if (!/^(234\d{10}|0\d{10})$/.test(msisdn)) {
            errors.push({
              row,
              field: 'msisdn',
              message: `Invalid MSISDN format: ${msisdn}`,
            });
          }

          // Validate prize_type
          const prizeType = columns[2]?.trim().toLowerCase() ?? '';
          const validPrizeTypes = ['airtime', 'data', 'points', 'cash', 'physical_goods'];
          if (!validPrizeTypes.includes(prizeType)) {
            errors.push({
              row,
              field: 'prize_type',
              message: `Invalid prize type: ${prizeType}. Must be one of: ${validPrizeTypes.join(', ')}`,
            });
          }

          // Validate prize_value (must be a positive number)
          const prizeValue = parseFloat(columns[3]?.trim() ?? '0');
          if (isNaN(prizeValue) || prizeValue <= 0) {
            errors.push({
              row,
              field: 'prize_value',
              message: `Invalid prize value: ${columns[3]}. Must be a positive number`,
            });
          }
        });

        resolve(errors);
      };

      reader.onerror = () => {
        errors.push({
          row: 0,
          field: 'file',
          message: 'Failed to read file',
        });
        resolve(errors);
      };

      reader.readAsText(file);
    });
  };

  const handleImportCSV = async () => {
    if (!selectedFile) {
      toast({
        title: 'No File Selected',
        description: 'Please select a CSV file to import',
        variant: 'destructive',
      });
      return;
    }

    if (!selectedDraw) {
      toast({
        title: 'No Draw Selected',
        description: 'Please select a draw to import winners for',
        variant: 'destructive',
      });
      return;
    }

    // Validation phase
    setImportProgress({
      status: 'validating',
      progress: 10,
      message: 'Validating CSV structure...',
    });

    const errors = await validateCSVStructure(selectedFile);

    if (errors.length > 0) {
      setValidationErrors(errors);
      setImportProgress({
        status: 'error',
        progress: 0,
        message: `Validation failed with ${errors.length} error(s)`,
        errors: errors.map((e) => `Row ${e.row}, ${e.field}: ${e.message}`),
      });

      toast({
        title: 'Validation Failed',
        description: `Found ${errors.length} validation error(s). Please fix and try again.`,
        variant: 'destructive',
      });
      return;
    }

    // Import phase
    setImportProgress({
      status: 'importing',
      progress: 50,
      message: 'Importing winners...',
    });

    try {
      const response = await drawCSVApi.importWinners(selectedDraw, selectedFile);

      if (response.success && response.data) {
        setImportProgress({
          status: 'complete',
          progress: 100,
          message: 'Import completed successfully!',
          totalWinners: response.data.total_winners,
          totalRunnersUp: response.data.total_runners_up,
        });

        toast({
          title: 'Import Successful',
          description: `Imported ${response.data.total_winners} winners and ${response.data.total_runners_up} runners-up`,
        });

        // Reset file input
        setSelectedFile(null);
        if (fileInputRef.current) {
          fileInputRef.current.value = '';
        }
      }
    } catch (error: any) {
      console.error('Import failed:', error);
      setImportProgress({
        status: 'error',
        progress: 0,
        message: error.response?.data?.error || 'Import failed. Please try again.',
      });

      toast({
        title: 'Import Failed',
        description: error.response?.data?.error || 'Failed to import winners',
        variant: 'destructive',
      });
    }
  };

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
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Draw CSV Management</h2>
          <p className="text-muted-foreground">
            Export draw entries and import winners for external draw processing
          </p>
        </div>
        <Button variant="outline" onClick={() => setShowHistoryDialog(true)}>
          <FileText className="h-4 w-4 mr-2" />
          Export History
        </Button>
      </div>

      {/* Draw Selection */}
      <Card>
        <CardHeader>
          <CardTitle>Select Draw</CardTitle>
          <CardDescription>
            Choose a draw to export entries or import winners
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-end gap-4">
            <div className="flex-1">
              <Label htmlFor="draw-select">Draw</Label>
              <Select value={selectedDraw} onValueChange={setSelectedDraw}>
                <SelectTrigger id="draw-select">
                  <SelectValue placeholder="Select a draw" />
                </SelectTrigger>
                <SelectContent>
                  {draws.map((draw) => (
                    <SelectItem key={draw.id} value={draw.id}>
                      {draw.name} - {draw.status}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Button
              onClick={handleOpenExportDialog}
              disabled={!selectedDraw}
            >
              <Download className="h-4 w-4 mr-2" />
              Export Entries
            </Button>
            <Button
              onClick={() => setShowImportDialog(true)}
              disabled={!selectedDraw}
              variant="secondary"
            >
              <Upload className="h-4 w-4 mr-2" />
              Import Winners
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Information Card */}
      <Card className="bg-blue-50 border-blue-200">
        <CardContent className="pt-6">
          <div className="flex gap-3">
            <Info className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
            <div className="space-y-2">
              <p className="text-sm font-medium text-blue-900">CSV Export/Import Process</p>
              <ol className="text-sm text-blue-800 space-y-1 list-decimal list-inside">
                <li>Export draw entries CSV with aggregated points from all sources (recharges, subscriptions, USSD, wheel spins)</li>
                <li>Process the CSV through your external draw engine</li>
                <li>Import the winners CSV back into the system</li>
                <li>System will automatically notify winners and manage prize claims</li>
              </ol>
              <p className="text-xs text-blue-700 mt-2">
                <strong>Note:</strong> Each MSISDN with N points will be entered N times in the draw, but can only win once per draw.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Export Dialog */}
      <Dialog open={showExportDialog} onOpenChange={setShowExportDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Export Draw Entries</DialogTitle>
            <DialogDescription>
              Configure export settings and download CSV file with aggregated points
            </DialogDescription>
          </DialogHeader>

          {exportProgress.status === 'idle' && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="start_date">Start Date</Label>
                  <input
                    id="start_date"
                    type="date"
                    value={exportConfig.start_date}
                    onChange={(e) =>
                      setExportConfig({ ...exportConfig, start_date: e.target.value })
                    }
                    className="w-full px-3 py-2 border rounded-md"
                  />
                </div>
                <div>
                  <Label htmlFor="end_date">End Date</Label>
                  <input
                    id="end_date"
                    type="date"
                    value={exportConfig.end_date}
                    onChange={(e) =>
                      setExportConfig({ ...exportConfig, end_date: e.target.value })
                    }
                    className="w-full px-3 py-2 border rounded-md"
                  />
                </div>
              </div>

              <div className="space-y-3">
                <Label>Points Sources to Include</Label>
                <div className="space-y-2">
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="include_subscription"
                      checked={exportConfig.include_subscription_points}
                      onCheckedChange={(checked) =>
                        setExportConfig({
                          ...exportConfig,
                          include_subscription_points: checked as boolean,
                        })
                      }
                    />
                    <label htmlFor="include_subscription" className="text-sm font-medium">
                      Daily Subscription Points
                    </label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="include_ussd"
                      checked={exportConfig.include_ussd_points}
                      onCheckedChange={(checked) =>
                        setExportConfig({
                          ...exportConfig,
                          include_ussd_points: checked as boolean,
                        })
                      }
                    />
                    <label htmlFor="include_ussd" className="text-sm font-medium">
                      USSD Recharge Points
                    </label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      id="include_wheel"
                      checked={exportConfig.include_wheel_points}
                      onCheckedChange={(checked) =>
                        setExportConfig({
                          ...exportConfig,
                          include_wheel_points: checked as boolean,
                        })
                      }
                    />
                    <label htmlFor="include_wheel" className="text-sm font-medium">
                      Wheel Spin Bonus Points
                    </label>
                  </div>
                </div>
              </div>

              <div className="bg-gray-50 rounded-lg p-4">
                <p className="text-sm font-medium text-gray-700 mb-2">CSV Format</p>
                <code className="text-xs text-gray-600 block">
                  MSISDN, Total_Points, Recharge_Points, Subscription_Points, USSD_Points, Wheel_Points
                </code>
                <p className="text-xs text-gray-500 mt-2">
                  Each row represents one MSISDN with aggregated points from the selected date range and sources.
                </p>
              </div>
            </div>
          )}

          {(exportProgress.status === 'preparing' ||
            exportProgress.status === 'exporting') && (
            <div className="space-y-4 py-6">
              <div className="flex items-center justify-center">
                <Loader2 className="h-12 w-12 animate-spin text-blue-600" />
              </div>
              <div className="space-y-2">
                <Progress value={exportProgress.progress} />
                <p className="text-center text-sm text-muted-foreground">
                  {exportProgress.message}
                </p>
              </div>
            </div>
          )}

          {exportProgress.status === 'complete' && (
            <div className="space-y-4 py-6">
              <div className="flex items-center justify-center">
                <CheckCircle2 className="h-16 w-16 text-green-600" />
              </div>
              <div className="text-center space-y-2">
                <p className="font-semibold text-lg">{exportProgress.message}</p>
                <div className="flex items-center justify-center gap-6 text-sm">
                  <div className="flex items-center gap-2">
                    <Users className="h-4 w-4 text-gray-500" />
                    <span>{exportProgress.totalMSISDNs} MSISDNs</span>
                  </div>
                </div>
              </div>
              <Button
                className="w-full"
                onClick={() => exportProgress.fileUrl && handleDownloadCSV(exportProgress.fileUrl)}
              >
                <Download className="h-4 w-4 mr-2" />
                Download CSV File
              </Button>
            </div>
          )}

          {exportProgress.status === 'error' && (
            <div className="space-y-4 py-6">
              <div className="flex items-center justify-center">
                <XCircle className="h-16 w-16 text-red-600" />
              </div>
              <div className="text-center space-y-2">
                <p className="font-semibold text-lg text-red-600">Export Failed</p>
                <p className="text-sm text-muted-foreground">{exportProgress.message}</p>
              </div>
              <Button
                className="w-full"
                variant="outline"
                onClick={() =>
                  setExportProgress({ status: 'idle', progress: 0, message: '' })
                }
              >
                <RefreshCw className="h-4 w-4 mr-2" />
                Try Again
              </Button>
            </div>
          )}

          {exportProgress.status === 'idle' && (
            <DialogFooter>
              <Button variant="outline" onClick={() => setShowExportDialog(false)}>
                Cancel
              </Button>
              <Button onClick={handleExportCSV}>
                <Download className="h-4 w-4 mr-2" />
                Export CSV
              </Button>
            </DialogFooter>
          )}
        </DialogContent>
      </Dialog>

      {/* Import Dialog */}
      <Dialog open={showImportDialog} onOpenChange={setShowImportDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Import Winners</DialogTitle>
            <DialogDescription>
              Upload CSV file with winners from external draw engine
            </DialogDescription>
          </DialogHeader>

          {importProgress.status === 'idle' && (
            <div className="space-y-4">
              <div>
                <Label htmlFor="file-upload">Winners CSV File</Label>
                <input
                  id="file-upload"
                  ref={fileInputRef}
                  type="file"
                  accept=".csv"
                  onChange={handleFileSelect}
                  className="w-full px-3 py-2 border rounded-md"
                />
                {selectedFile && (
                  <p className="text-sm text-green-600 mt-2">
                    Selected: {selectedFile.name} ({(selectedFile.size / 1024).toFixed(2)} KB)
                  </p>
                )}
              </div>

              <div className="bg-gray-50 rounded-lg p-4">
                <p className="text-sm font-medium text-gray-700 mb-2">Required CSV Format</p>
                <code className="text-xs text-gray-600 block mb-2">
                  MSISDN, Prize_Name, Prize_Type, Prize_Value, Is_Runner_Up
                </code>
                <div className="text-xs text-gray-600 space-y-1">
                  <p><strong>MSISDN:</strong> Nigerian format (234XXXXXXXXXX or 0XXXXXXXXXX)</p>
                  <p><strong>Prize_Type:</strong> airtime, data, points, cash, or physical_goods</p>
                  <p><strong>Prize_Value:</strong> Numeric value (amount or quantity)</p>
                  <p><strong>Is_Runner_Up:</strong> true or false</p>
                </div>
              </div>

              {validationErrors.length > 0 && (
                <div className="bg-red-50 border border-red-200 rounded-lg p-4 max-h-60 overflow-y-auto">
                  <p className="text-sm font-medium text-red-800 mb-2">
                    Validation Errors ({validationErrors.length})
                  </p>
                  <div className="space-y-1">
                    {validationErrors.slice(0, 10).map((error, index) => (
                      <p key={index} className="text-xs text-red-700">
                        Row {error.row}, {error.field}: {error.message}
                      </p>
                    ))}
                    {validationErrors.length > 10 && (
                      <p className="text-xs text-red-700 font-medium">
                        ... and {validationErrors.length - 10} more errors
                      </p>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}

          {(importProgress.status === 'validating' ||
            importProgress.status === 'importing') && (
            <div className="space-y-4 py-6">
              <div className="flex items-center justify-center">
                <Loader2 className="h-12 w-12 animate-spin text-blue-600" />
              </div>
              <div className="space-y-2">
                <Progress value={importProgress.progress} />
                <p className="text-center text-sm text-muted-foreground">
                  {importProgress.message}
                </p>
              </div>
            </div>
          )}

          {importProgress.status === 'complete' && (
            <div className="space-y-4 py-6">
              <div className="flex items-center justify-center">
                <CheckCircle2 className="h-16 w-16 text-green-600" />
              </div>
              <div className="text-center space-y-2">
                <p className="font-semibold text-lg">{importProgress.message}</p>
                <div className="flex items-center justify-center gap-6 text-sm">
                  <div>
                    <p className="text-2xl font-bold text-green-600">
                      {importProgress.totalWinners}
                    </p>
                    <p className="text-muted-foreground">Winners</p>
                  </div>
                  <div>
                    <p className="text-2xl font-bold text-blue-600">
                      {importProgress.totalRunnersUp}
                    </p>
                    <p className="text-muted-foreground">Runners-Up</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {importProgress.status === 'error' && (
            <div className="space-y-4 py-6">
              <div className="flex items-center justify-center">
                <XCircle className="h-16 w-16 text-red-600" />
              </div>
              <div className="text-center space-y-2">
                <p className="font-semibold text-lg text-red-600">Import Failed</p>
                <p className="text-sm text-muted-foreground">{importProgress.message}</p>
              </div>
              {importProgress.errors && importProgress.errors.length > 0 && (
                <div className="bg-red-50 border border-red-200 rounded-lg p-4 max-h-40 overflow-y-auto">
                  {importProgress.errors.slice(0, 5).map((error, index) => (
                    <p key={index} className="text-xs text-red-700">
                      {error}
                    </p>
                  ))}
                </div>
              )}
              <Button
                className="w-full"
                variant="outline"
                onClick={() => {
                  setImportProgress({ status: 'idle', progress: 0, message: '' });
                  setValidationErrors([]);
                }}
              >
                <RefreshCw className="h-4 w-4 mr-2" />
                Try Again
              </Button>
            </div>
          )}

          {importProgress.status === 'idle' && (
            <DialogFooter>
              <Button variant="outline" onClick={() => setShowImportDialog(false)}>
                Cancel
              </Button>
              <Button onClick={handleImportCSV} disabled={!selectedFile}>
                <Upload className="h-4 w-4 mr-2" />
                Import Winners
              </Button>
            </DialogFooter>
          )}
        </DialogContent>
      </Dialog>

      {/* Export History Dialog */}
      <Dialog open={showHistoryDialog} onOpenChange={setShowHistoryDialog}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>Export History</DialogTitle>
            <DialogDescription>
              View all previous CSV exports with download links
            </DialogDescription>
          </DialogHeader>

          <div className="max-h-96 overflow-y-auto">
            {exportHistory.length === 0 ? (
              <p className="text-center text-muted-foreground py-8">No export history available</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Draw</TableHead>
                    <TableHead>Exported By</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>MSISDNs</TableHead>
                    <TableHead>Total Points</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {exportHistory.map((history) => (
                    <TableRow key={history.id}>
                      <TableCell className="font-medium">
                        {draws.find((d) => d.id === history.draw_id)?.name || history.draw_id}
                      </TableCell>
                      <TableCell>{history.exported_by}</TableCell>
                      <TableCell>
                        {new Date(history.exported_at).toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary">{history.total_msisdns}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary">{history.total_points}</Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDownloadCSV(history.file_url)}
                        >
                          <Download className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </div>

          <DialogFooter>
            <Button onClick={() => setShowHistoryDialog(false)}>Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
