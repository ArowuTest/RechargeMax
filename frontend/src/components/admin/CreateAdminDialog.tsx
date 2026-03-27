import React, { useState } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, Shield, UserPlus, Info } from 'lucide-react';

interface CreateAdminDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (adminData: AdminFormData) => Promise<void>;
  loading?: boolean;
}

interface AdminFormData {
  full_name: string;
  email: string;
  role: string;
  permissions: string[];
  is_active: boolean;
}

const ADMIN_ROLES = [
  { value: 'ADMIN', label: 'Admin', description: 'Standard admin with limited permissions' },
  { value: 'SUPER_ADMIN', label: 'Super Admin', description: 'Full system access and control' },
  { value: 'MODERATOR', label: 'Moderator', description: 'User and transaction management only' },
  { value: 'SUPPORT', label: 'Support', description: 'Customer support — view and manage users' },
  { value: 'VIEWER', label: 'Viewer', description: 'Read-only access to analytics and monitoring' },
];

const AVAILABLE_PERMISSIONS = [
  { key: 'view_analytics', label: 'View Analytics', description: 'Access to dashboard and reports' },
  { key: 'manage_users', label: 'Manage Users', description: 'View and manage user accounts' },
  { key: 'manage_transactions', label: 'Manage Transactions', description: 'View and process transactions' },
  { key: 'manage_networks', label: 'Manage Networks', description: 'Configure network providers and data plans' },
  { key: 'manage_prizes', label: 'Manage Prizes', description: 'Configure wheel prizes and draws' },
  { key: 'manage_affiliates', label: 'Manage Affiliates', description: 'Approve and manage affiliate accounts' },
  { key: 'manage_settings', label: 'Manage Settings', description: 'Configure platform settings' },
  { key: 'manage_admins', label: 'Manage Admins', description: 'Create and manage admin accounts' },
  { key: 'view_monitoring', label: 'View Monitoring', description: 'Access system monitoring dashboard' },
  { key: 'manage_draws', label: 'Manage Draws', description: 'Configure and run prize draws' }
];

export const CreateAdminDialog: React.FC<CreateAdminDialogProps> = ({
  open,
  onOpenChange,
  onSave,
  loading = false
}) => {
  const [formData, setFormData] = useState<AdminFormData>({
    full_name: '',
    email: '',
    role: 'ADMIN',
    permissions: [],
    is_active: true
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.full_name.trim()) {
      newErrors.full_name = 'Full name is required';
    }

    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email address';
    }

    if (!formData.role) {
      newErrors.role = 'Please select a role';
    }

    if (formData.role === 'ADMIN' && formData.permissions.length === 0) {
      newErrors.permissions = 'Please select at least one permission for Admin role';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    try {
      // Super Admin gets all permissions automatically
      const finalPermissions = formData.role === 'SUPER_ADMIN' 
        ? AVAILABLE_PERMISSIONS.map(p => p.key)
        : formData.permissions;

      await onSave({
        ...formData,
        permissions: finalPermissions
      });
      
      // Reset form
      setFormData({
        full_name: '',
        email: '',
        role: 'ADMIN',
        permissions: [],
        is_active: true
      });
      setErrors({});
    } catch (error) {
      console.error('Failed to create admin:', error);
    }
  };

  const handleInputChange = (field: keyof AdminFormData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Clear error for this field when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  const handlePermissionChange = (permission: string, checked: boolean) => {
    setFormData(prev => ({
      ...prev,
      permissions: checked
        ? [...prev.permissions, permission]
        : prev.permissions.filter(p => p !== permission)
    }));
    
    // Clear permissions error when user selects any permission
    if (errors.permissions) {
      setErrors(prev => ({ ...prev, permissions: '' }));
    }
  };

  const handleRoleChange = (role: string) => {
    handleInputChange('role', role);
    
    // Clear permissions when switching to Super Admin
    if (role === 'SUPER_ADMIN') {
      handleInputChange('permissions', []);
    }
  };

  const getSelectedRole = () => {
    return ADMIN_ROLES.find(role => role.value === formData.role);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <UserPlus className="w-5 h-5" />
            Create New Admin
          </DialogTitle>
          <DialogDescription>
            Create a new admin account with specific roles and permissions
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Full Name */}
          <div>
            <Label htmlFor="full_name">Full Name</Label>
            <Input
              id="full_name"
              value={formData.full_name}
              onChange={(e) => handleInputChange('full_name', e.target.value)}
              placeholder="Enter admin's full name"
              className={errors.full_name ? 'border-red-500' : ''}
            />
            {errors.full_name && (
              <p className="text-red-500 text-sm mt-1">{errors.full_name}</p>
            )}
          </div>

          {/* Email */}
          <div>
            <Label htmlFor="email">Email Address</Label>
            <Input
              id="email"
              type="email"
              value={formData.email}
              onChange={(e) => handleInputChange('email', e.target.value)}
              placeholder="admin@rechargemax.ng"
              className={errors.email ? 'border-red-500' : ''}
            />
            {errors.email && (
              <p className="text-red-500 text-sm mt-1">{errors.email}</p>
            )}
          </div>

          {/* Role Selection */}
          <div>
            <Label htmlFor="role">Admin Role</Label>
            <Select 
              value={formData.role} 
              onValueChange={handleRoleChange}
            >
              <SelectTrigger className={errors.role ? 'border-red-500' : ''}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {ADMIN_ROLES.map((role) => (
                  <SelectItem key={role.value} value={role.value}>
                    <div>
                      <div className="flex items-center gap-2">
                        <Shield className="w-4 h-4" />
                        {role.label}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        {role.description}
                      </div>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.role && (
              <p className="text-red-500 text-sm mt-1">{errors.role}</p>
            )}
            
            {getSelectedRole() && (
              <div className="mt-2 p-2 bg-blue-50 rounded text-sm text-blue-700">
                <Info className="w-4 h-4 inline mr-1" />
                {getSelectedRole()?.description}
              </div>
            )}
          </div>

          {/* Permissions (only for ADMIN role) */}
          {formData.role === 'ADMIN' && (
            <div>
              <Label>Permissions</Label>
              <div className="mt-2 space-y-3 max-h-48 overflow-y-auto border rounded-lg p-3">
                {AVAILABLE_PERMISSIONS.map((permission) => (
                  <div key={permission.key} className="flex items-start space-x-3">
                    <Checkbox
                      id={permission.key}
                      checked={formData.permissions.includes(permission.key)}
                      onCheckedChange={(checked) => 
                        handlePermissionChange(permission.key, checked as boolean)
                      }
                    />
                    <div className="flex-1">
                      <label 
                        htmlFor={permission.key}
                        className="text-sm font-medium cursor-pointer"
                      >
                        {permission.label}
                      </label>
                      <p className="text-xs text-gray-500 mt-1">
                        {permission.description}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
              {errors.permissions && (
                <p className="text-red-500 text-sm mt-1">{errors.permissions}</p>
              )}
              
              {formData.permissions.length > 0 && (
                <div className="mt-2">
                  <Label className="text-xs text-gray-600">Selected Permissions:</Label>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {formData.permissions.map(permission => {
                      const perm = AVAILABLE_PERMISSIONS.find(p => p.key === permission);
                      return (
                        <Badge key={permission} variant="secondary" className="text-xs">
                          {perm?.label}
                        </Badge>
                      );
                    })}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Super Admin Notice */}
          {formData.role === 'SUPER_ADMIN' && (
            <Alert className="border-yellow-200 bg-yellow-50">
              <Shield className="h-4 w-4 text-yellow-600" />
              <AlertDescription className="text-yellow-800">
                <strong>Super Admin</strong> role grants full access to all system features and permissions automatically.
              </AlertDescription>
            </Alert>
          )}

          {/* Account Status */}
          <div className="flex items-center justify-between p-3 border rounded-lg">
            <div>
              <div className="font-medium">Account Active</div>
              <div className="text-sm text-gray-500">Enable login for this admin account</div>
            </div>
            <Checkbox
              checked={formData.is_active}
              onCheckedChange={(checked) => handleInputChange('is_active', checked)}
            />
          </div>

          {/* Security Notice */}
          <Alert className="border-blue-200 bg-blue-50">
            <Info className="h-4 w-4 text-blue-600" />
            <AlertDescription className="text-blue-800">
              A temporary password will be generated and displayed after account creation. 
              The admin will be required to change it on first login.
            </AlertDescription>
          </Alert>

          {/* Action Buttons */}
          <div className="flex gap-2 pt-4">
            <Button
              type="submit"
              disabled={loading}
              className="flex-1"
            >
              {loading ? (
                <Loader2 className="w-4 h-4 animate-spin mr-2" />
              ) : (
                <UserPlus className="w-4 h-4 mr-2" />
              )}
              Create Admin Account
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={loading}
            >
              Cancel
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
};