import React, { useState, useEffect } from 'react';
import { Plus, Edit2, Trash2, Save, X, ChevronUp, ChevronDown } from 'lucide-react';

interface DrawType {
  id: number;
  name: string;
  description: string;
  is_active: boolean;
}

interface PrizeCategory {
  id?: number;
  category_name: string;
  prize_amount: number;
  winner_count: number;
  runner_up_count: number;
  display_order: number;
}

interface PrizeTemplate {
  id?: number;
  name: string;
  draw_type_id: number;
  description: string;
  is_default: boolean;
  is_active: boolean;
  categories?: PrizeCategory[];
}

const API_BASE = '/api/v1';

const PrizeTemplateManagement: React.FC = () => {
  const [drawTypes, setDrawTypes] = useState<DrawType[]>([]);
  const [templates, setTemplates] = useState<PrizeTemplate[]>([]);
  const [selectedTemplate, setSelectedTemplate] = useState<PrizeTemplate | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Form state
  const [formData, setFormData] = useState<PrizeTemplate>({
    name: '',
    draw_type_id: 0,
    description: '',
    is_default: false,
    is_active: true,
    categories: []
  });

  // Category form state
  const [newCategory, setNewCategory] = useState<PrizeCategory>({
    category_name: '',
    prize_amount: 0,
    winner_count: 1,
    runner_up_count: 0,
    display_order: 1
  });

  useEffect(() => {
    fetchDrawTypes();
    fetchTemplates();
  }, []);

  const fetchDrawTypes = async () => {
    try {
      const response = await fetch(`${API_BASE}/admin/draw-types`, { credentials: 'include' });
      const data = await response.json();
      if (data.success) {
        setDrawTypes(data.data || []);
      }
    } catch (err) {
      console.error('Failed to fetch draw types:', err);
    }
  };

  const fetchTemplates = async () => {
    setLoading(true);
    try {
      const response = await fetch(`${API_BASE}/admin/prize-templates`, { credentials: 'include' });
      const data = await response.json();
      if (data.success) {
        setTemplates(data.data || []);
      }
    } catch (err) {
      setError('Failed to fetch templates');
    } finally {
      setLoading(false);
    }
  };

  const fetchTemplateDetails = async (templateId: number) => {
    try {
      const response = await fetch(`${API_BASE}/admin/prize-templates/${templateId}`, { credentials: 'include' });
      const data = await response.json();
      if (data.success) {
        setSelectedTemplate(data.data);
        setFormData(data.data);
      }
    } catch (err) {
      setError('Failed to fetch template details');
    }
  };

  const handleCreateTemplate = async () => {
    if (!formData.name || !formData.draw_type_id || formData.categories!.length === 0) {
      setError('Please fill in all required fields and add at least one category');
      return;
    }

    setLoading(true);
    setError('');
    try {
      const response = await fetch(`${API_BASE}/admin/prize-templates`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      });

      const data = await response.json();
      if (data.success) {
        setSuccess('Template created successfully!');
        setIsCreating(false);
        resetForm();
        fetchTemplates();
        setTimeout(() => setSuccess(''), 3000);
      } else {
        setError(data.message || 'Failed to create template');
      }
    } catch (err) {
      setError('Failed to create template');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateTemplate = async () => {
    if (!selectedTemplate?.id) return;

    setLoading(true);
    setError('');
    try {
      const response = await fetch(`${API_BASE}/admin/prize-templates/${selectedTemplate.id}`, {
        method: 'PUT',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      });

      const data = await response.json();
      if (data.success) {
        setSuccess('Template updated successfully!');
        setIsEditing(false);
        setSelectedTemplate(null);
        resetForm();
        fetchTemplates();
        setTimeout(() => setSuccess(''), 3000);
      } else {
        setError(data.message || 'Failed to update template');
      }
    } catch (err) {
      setError('Failed to update template');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteTemplate = async (templateId: number) => {
    if (!confirm('Are you sure you want to delete this template? This will also delete all associated categories.')) {
      return;
    }

    setLoading(true);
    setError('');
    try {
      const response = await fetch(`${API_BASE}/admin/prize-templates/${templateId}`, {
        method: 'DELETE',
        credentials: 'include',
      });

      const data = await response.json();
      if (data.success) {
        setSuccess('Template deleted successfully!');
        fetchTemplates();
        setTimeout(() => setSuccess(''), 3000);
      } else {
        setError(data.message || 'Failed to delete template');
      }
    } catch (err) {
      setError('Failed to delete template');
    } finally {
      setLoading(false);
    }
  };

  const addCategory = () => {
    if (!newCategory.category_name || newCategory.prize_amount <= 0) {
      setError('Please fill in category name and prize amount');
      return;
    }

    const categories = formData.categories || [];
    setFormData({
      ...formData,
      categories: [
        ...categories,
        { ...newCategory, display_order: categories.length + 1 }
      ]
    });

    setNewCategory({
      category_name: '',
      prize_amount: 0,
      winner_count: 1,
      runner_up_count: 0,
      display_order: 1
    });
    setError('');
  };

  const removeCategory = (index: number) => {
    const categories = formData.categories || [];
    setFormData({
      ...formData,
      categories: categories.filter((_, i) => i !== index).map((cat, i) => ({
        ...cat,
        display_order: i + 1
      }))
    });
  };

  const moveCategoryUp = (index: number) => {
    if (index === 0) return;
    const categories = [...(formData.categories || [])];
    const temp = categories[index - 1];
    if (temp && categories[index]) {
      categories[index - 1] = categories[index]!;
      categories[index] = temp;
    }
    setFormData({
      ...formData,
      categories: categories.map((cat, i) => ({ ...cat, display_order: i + 1 }))
    });
  };

  const moveCategoryDown = (index: number) => {
    const categories = formData.categories || [];
    if (index === categories.length - 1) return;
    const newCategories = [...categories];
    const temp = newCategories[index];
    if (temp && newCategories[index + 1]) {
      newCategories[index] = newCategories[index + 1]!;
      newCategories[index + 1] = temp;
    }
    setFormData({
      ...formData,
      categories: newCategories.map((cat, i) => ({ ...cat, display_order: i + 1 }))
    });
  };

  const resetForm = () => {
    setFormData({
      name: '',
      draw_type_id: 0,
      description: '',
      is_default: false,
      is_active: true,
      categories: []
    });
    setNewCategory({
      category_name: '',
      prize_amount: 0,
      winner_count: 1,
      runner_up_count: 0,
      display_order: 1
    });
  };

  const startEdit = (template: PrizeTemplate) => {
    fetchTemplateDetails(template.id!);
    setIsEditing(true);
    setIsCreating(false);
  };

  const startCreate = () => {
    resetForm();
    setIsCreating(true);
    setIsEditing(false);
    setSelectedTemplate(null);
  };

  const cancelEdit = () => {
    setIsCreating(false);
    setIsEditing(false);
    setSelectedTemplate(null);
    resetForm();
    setError('');
  };

  const calculateTotalPool = (categories: PrizeCategory[]) => {
    return categories.reduce((sum, cat) => sum + (cat.prize_amount * cat.winner_count), 0);
  };

  const getDrawTypeName = (drawTypeId: number) => {
    const drawType = drawTypes.find(dt => dt.id === drawTypeId);
    return drawType?.name || 'Unknown';
  };

  return (
    <div className="prize-template-management p-6">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Prize Template Management</h2>
        {!isCreating && !isEditing && (
          <button
            onClick={startCreate}
            className="flex items-center gap-2 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
          >
            <Plus size={20} />
            Create Template
          </button>
        )}
      </div>

      {/* Success/Error Messages */}
      {success && (
        <div className="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded mb-4">
          {success}
        </div>
      )}
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      {/* Create/Edit Form */}
      {(isCreating || isEditing) && (
        <div className="bg-white shadow-lg rounded-lg p-6 mb-6">
          <h3 className="text-xl font-bold mb-4">
            {isCreating ? 'Create New Template' : 'Edit Template'}
          </h3>

          <div className="grid grid-cols-2 gap-4 mb-4">
            <div>
              <label className="block text-sm font-medium mb-2">Template Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="w-full border rounded px-3 py-2"
                placeholder="e.g., Daily Mega Template"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Draw Type *</label>
              <select
                value={formData.draw_type_id}
                onChange={(e) => setFormData({ ...formData, draw_type_id: parseInt(e.target.value) })}
                className="w-full border rounded px-3 py-2"
              >
                <option value={0}>Select Draw Type</option>
                {drawTypes.map(dt => (
                  <option key={dt.id} value={dt.id}>{dt.name}</option>
                ))}
              </select>
            </div>

            <div className="col-span-2">
              <label className="block text-sm font-medium mb-2">Description</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full border rounded px-3 py-2"
                rows={2}
                placeholder="Template description..."
              />
            </div>

            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={formData.is_default}
                  onChange={(e) => setFormData({ ...formData, is_default: e.target.checked })}
                />
                <span className="text-sm">Set as Default</span>
              </label>

              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={formData.is_active}
                  onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                />
                <span className="text-sm">Active</span>
              </label>
            </div>
          </div>

          {/* Category Builder */}
          <div className="border-t pt-4 mt-4">
            <h4 className="font-bold mb-3">Prize Categories *</h4>

            {/* Add Category Form */}
            <div className="bg-gray-50 p-4 rounded mb-4">
              <div className="grid grid-cols-5 gap-3 mb-3">
                <input
                  type="text"
                  value={newCategory.category_name}
                  onChange={(e) => setNewCategory({ ...newCategory, category_name: e.target.value })}
                  className="border rounded px-3 py-2"
                  placeholder="Category Name"
                />
                <input
                  type="number"
                  value={newCategory.prize_amount}
                  onChange={(e) => setNewCategory({ ...newCategory, prize_amount: parseFloat(e.target.value) })}
                  className="border rounded px-3 py-2"
                  placeholder="Prize Amount"
                />
                <input
                  type="number"
                  value={newCategory.winner_count}
                  onChange={(e) => setNewCategory({ ...newCategory, winner_count: parseInt(e.target.value) })}
                  className="border rounded px-3 py-2"
                  placeholder="Winners"
                  min="1"
                />
                <input
                  type="number"
                  value={newCategory.runner_up_count}
                  onChange={(e) => setNewCategory({ ...newCategory, runner_up_count: parseInt(e.target.value) })}
                  className="border rounded px-3 py-2"
                  placeholder="Runner-ups"
                  min="0"
                />
                <button
                  onClick={addCategory}
                  className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
                >
                  Add Category
                </button>
              </div>
            </div>

            {/* Categories List */}
            {formData.categories && formData.categories.length > 0 ? (
              <div className="space-y-2">
                {formData.categories.map((category, index) => (
                  <div key={index} className="flex items-center gap-3 bg-white border rounded p-3">
                    <div className="flex flex-col gap-1">
                      <button
                        onClick={() => moveCategoryUp(index)}
                        disabled={index === 0}
                        className="text-gray-600 hover:text-blue-600 disabled:opacity-30"
                      >
                        <ChevronUp size={16} />
                      </button>
                      <button
                        onClick={() => moveCategoryDown(index)}
                        disabled={index === formData.categories!.length - 1}
                        className="text-gray-600 hover:text-blue-600 disabled:opacity-30"
                      >
                        <ChevronDown size={16} />
                      </button>
                    </div>
                    <div className="flex-1 grid grid-cols-4 gap-3">
                      <div>
                        <div className="text-xs text-gray-500">Category</div>
                        <div className="font-medium">{category.category_name}</div>
                      </div>
                      <div>
                        <div className="text-xs text-gray-500">Prize Amount</div>
                        <div className="font-medium">₦{category.prize_amount.toLocaleString()}</div>
                      </div>
                      <div>
                        <div className="text-xs text-gray-500">Winners</div>
                        <div className="font-medium">{category.winner_count}</div>
                      </div>
                      <div>
                        <div className="text-xs text-gray-500">Runner-ups</div>
                        <div className="font-medium">{category.runner_up_count}</div>
                      </div>
                    </div>
                    <button
                      onClick={() => removeCategory(index)}
                      className="text-red-600 hover:text-red-800"
                    >
                      <Trash2 size={18} />
                    </button>
                  </div>
                ))}

                <div className="bg-blue-50 border border-blue-200 rounded p-3 mt-4">
                  <div className="flex justify-between items-center">
                    <span className="font-bold">Total Prize Pool:</span>
                    <span className="text-xl font-bold text-blue-600">
                      ₦{calculateTotalPool(formData.categories).toLocaleString()}
                    </span>
                  </div>
                  <div className="text-sm text-gray-600 mt-1">
                    {formData.categories.length} categories • {' '}
                    {formData.categories.reduce((sum, cat) => sum + cat.winner_count, 0)} total winners • {' '}
                    {formData.categories.reduce((sum, cat) => sum + cat.runner_up_count, 0)} total runner-ups
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-500 py-8 border-2 border-dashed rounded">
                No categories added yet. Add at least one category to create the template.
              </div>
            )}
          </div>

          {/* Form Actions */}
          <div className="flex gap-3 mt-6">
            <button
              onClick={isCreating ? handleCreateTemplate : handleUpdateTemplate}
              disabled={loading}
              className="flex items-center gap-2 bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
            >
              <Save size={18} />
              {loading ? 'Saving...' : (isCreating ? 'Create Template' : 'Update Template')}
            </button>
            <button
              onClick={cancelEdit}
              className="flex items-center gap-2 bg-gray-300 text-gray-700 px-6 py-2 rounded hover:bg-gray-400"
            >
              <X size={18} />
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Templates List */}
      {!isCreating && !isEditing && (
        <div className="bg-white shadow rounded-lg">
          <div className="p-4 border-b">
            <h3 className="font-bold">Existing Templates</h3>
          </div>
          {loading ? (
            <div className="p-8 text-center text-gray-500">Loading templates...</div>
          ) : templates.length === 0 ? (
            <div className="p-8 text-center text-gray-500">
              No templates found. Create your first template to get started.
            </div>
          ) : (
            <div className="divide-y">
              {templates.map(template => (
                <div key={template.id} className="p-4 hover:bg-gray-50">
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h4 className="font-bold text-lg">{template.name}</h4>
                        <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded">
                          {getDrawTypeName(template.draw_type_id)}
                        </span>
                        {template.is_default && (
                          <span className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded">
                            Default
                          </span>
                        )}
                        {!template.is_active && (
                          <span className="px-2 py-1 bg-gray-100 text-gray-800 text-xs rounded">
                            Inactive
                          </span>
                        )}
                      </div>
                      {template.description && (
                        <p className="text-sm text-gray-600 mb-2">{template.description}</p>
                      )}
                      {template.categories && template.categories.length > 0 && (
                        <div className="text-sm text-gray-600">
                          {template.categories.length} categories • 
                          Total Pool: ₦{calculateTotalPool(template.categories).toLocaleString()}
                        </div>
                      )}
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={() => startEdit(template)}
                        className="flex items-center gap-1 text-blue-600 hover:text-blue-800 px-3 py-1 border border-blue-600 rounded"
                      >
                        <Edit2 size={16} />
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteTemplate(template.id!)}
                        className="flex items-center gap-1 text-red-600 hover:text-red-800 px-3 py-1 border border-red-600 rounded"
                      >
                        <Trash2 size={16} />
                        Delete
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default PrizeTemplateManagement;
