-- File storage setup for RechargeMax platform
-- Created: 2026-01-30 14:00 UTC
-- ============================================================================

-- Create storage buckets
INSERT INTO storage.buckets (id, name, public, file_size_limit, allowed_mime_types)
VALUES 
    (
        'user-avatars',
        'user-avatars',
        true,
        5242880, -- 5MB limit
        ARRAY['image/jpeg', 'image/png', 'image/webp', 'image/gif']
    ),
    (
        'user-documents',
        'user-documents',
        false,
        10485760, -- 10MB limit
        ARRAY['image/jpeg', 'image/png', 'application/pdf', 'image/webp']
    ),
    (
        'promotional-materials',
        'promotional-materials',
        true,
        20971520, -- 20MB limit
        ARRAY['image/jpeg', 'image/png', 'image/webp', 'image/gif', 'video/mp4', 'video/webm']
    ),
    (
        'transaction-receipts',
        'transaction-receipts',
        false,
        5242880, -- 5MB limit
        ARRAY['image/jpeg', 'image/png', 'application/pdf']
    )
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- STORAGE POLICIES
-- ============================================================================

-- User Avatars Bucket Policies
CREATE POLICY "Users can upload their own avatar"
ON storage.objects FOR INSERT
WITH CHECK (
    bucket_id = 'user-avatars' 
    AND auth.uid()::text = (storage.foldername(name))[1]
);

CREATE POLICY "Users can update their own avatar"
ON storage.objects FOR UPDATE
USING (
    bucket_id = 'user-avatars' 
    AND auth.uid()::text = (storage.foldername(name))[1]
);

CREATE POLICY "Users can delete their own avatar"
ON storage.objects FOR DELETE
USING (
    bucket_id = 'user-avatars' 
    AND auth.uid()::text = (storage.foldername(name))[1]
);

CREATE POLICY "Anyone can view avatars"
ON storage.objects FOR SELECT
USING (bucket_id = 'user-avatars');

-- User Documents Bucket Policies
CREATE POLICY "Users can upload their own documents"
ON storage.objects FOR INSERT
WITH CHECK (
    bucket_id = 'user-documents' 
    AND auth.uid()::text = (storage.foldername(name))[1]
);

CREATE POLICY "Users can view their own documents"
ON storage.objects FOR SELECT
USING (
    bucket_id = 'user-documents' 
    AND auth.uid()::text = (storage.foldername(name))[1]
);

CREATE POLICY "Users can delete their own documents"
ON storage.objects FOR DELETE
USING (
    bucket_id = 'user-documents' 
    AND auth.uid()::text = (storage.foldername(name))[1]
);

-- Promotional Materials Bucket Policies
CREATE POLICY "Anyone can view promotional materials"
ON storage.objects FOR SELECT
USING (bucket_id = 'promotional-materials');

CREATE POLICY "Service role can manage promotional materials"
ON storage.objects FOR ALL
USING (
    bucket_id = 'promotional-materials' 
    AND auth.role() = 'service_role'
);

-- Transaction Receipts Bucket Policies
CREATE POLICY "Users can upload transaction receipts"
ON storage.objects FOR INSERT
WITH CHECK (
    bucket_id = 'transaction-receipts' 
    AND auth.uid()::text = (storage.foldername(name))[1]
);

CREATE POLICY "Users can view their own receipts"
ON storage.objects FOR SELECT
USING (
    bucket_id = 'transaction-receipts' 
    AND auth.uid()::text = (storage.foldername(name))[1]
);

CREATE POLICY "Service role can access all receipts"
ON storage.objects FOR SELECT
USING (
    bucket_id = 'transaction-receipts' 
    AND auth.role() = 'service_role'
);

-- ============================================================================
-- FILE MANAGEMENT TABLE
-- ============================================================================

-- Table to track uploaded files and their metadata
CREATE TABLE public.file_uploads_2026_01_30_14_00 (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES public.users_2026_01_30_14_00(id) ON DELETE CASCADE,
    
    -- File details
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    bucket_name TEXT NOT NULL,
    
    -- File purpose
    file_type TEXT NOT NULL, -- 'avatar', 'document', 'receipt', 'promotional'
    description TEXT,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    
    -- Metadata
    upload_ip INET,
    user_agent TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_file_size CHECK (file_size > 0),
    CONSTRAINT valid_file_type CHECK (file_type IN ('avatar', 'document', 'receipt', 'promotional'))
);

-- Enable RLS on file uploads table
ALTER TABLE public.file_uploads_2026_01_30_14_00 ENABLE ROW LEVEL SECURITY;

-- RLS policies for file uploads
CREATE POLICY "users_select_own_files" ON public.file_uploads_2026_01_30_14_00
    FOR SELECT USING (auth.uid() = (SELECT auth_user_id FROM public.users_2026_01_30_14_00 WHERE id = user_id));

CREATE POLICY "users_insert_own_files" ON public.file_uploads_2026_01_30_14_00
    FOR INSERT WITH CHECK (auth.uid() = (SELECT auth_user_id FROM public.users_2026_01_30_14_00 WHERE id = user_id));

CREATE POLICY "users_update_own_files" ON public.file_uploads_2026_01_30_14_00
    FOR UPDATE USING (auth.uid() = (SELECT auth_user_id FROM public.users_2026_01_30_14_00 WHERE id = user_id));

CREATE POLICY "users_delete_own_files" ON public.file_uploads_2026_01_30_14_00
    FOR DELETE USING (auth.uid() = (SELECT auth_user_id FROM public.users_2026_01_30_14_00 WHERE id = user_id));

CREATE POLICY "service_role_manage_all_files" ON public.file_uploads_2026_01_30_14_00
    FOR ALL USING (auth.role() = 'service_role');

-- Indexes for file uploads
CREATE INDEX idx_file_uploads_user_id ON public.file_uploads_2026_01_30_14_00(user_id);
CREATE INDEX idx_file_uploads_file_type ON public.file_uploads_2026_01_30_14_00(file_type);
CREATE INDEX idx_file_uploads_bucket_name ON public.file_uploads_2026_01_30_14_00(bucket_name);
CREATE INDEX idx_file_uploads_created_at ON public.file_uploads_2026_01_30_14_00(created_at DESC);

-- Trigger for updated_at
CREATE TRIGGER update_file_uploads_updated_at 
    BEFORE UPDATE ON public.file_uploads_2026_01_30_14_00
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

