import React from 'react';
import { Outlet } from 'react-router-dom';
import { Header } from '@/components/Header';

/**
 * AdminLayout Component
 * 
 * Enterprise-grade layout component for admin routes.
 * Provides consistent header and renders child routes via Outlet.
 * 
 * This fixes the routing issue where all admin routes were rendering
 * the same component. The Outlet component from React Router will
 * render the matched child route component.
 */
export const AdminLayout: React.FC = () => {
  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <main className="container mx-auto px-4 py-8">
        {/* Outlet renders the matched child route component */}
        <Outlet />
      </main>
    </div>
  );
};
