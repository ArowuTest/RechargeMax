import React, { useState } from 'react';
import { DrawsList } from '@/components/draws/DrawsList';
import { LoginModal } from '@/components/auth/LoginModal';

export const DrawsPage: React.FC = () => {
  const [showLoginModal, setShowLoginModal] = useState(false);

  const handleLoginRequired = () => {
    setShowLoginModal(true);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50 p-4">
      <div className="max-w-7xl mx-auto">
        <DrawsList onLoginRequired={handleLoginRequired} />
      </div>

      <LoginModal 
        isOpen={showLoginModal}
        onClose={() => setShowLoginModal(false)}
        onSuccess={() => {
          setShowLoginModal(false);
          // Refresh the page to show updated user entries
          window.location.reload();
        }}
      />
    </div>
  );
};

export default DrawsPage;
