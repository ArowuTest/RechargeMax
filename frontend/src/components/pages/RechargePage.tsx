import React from 'react';
import { PremiumRechargeForm } from '@/components/recharge/PremiumRechargeForm';

export const RechargePage: React.FC = () => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-50">
      <PremiumRechargeForm />
    </div>
  );
};

export default RechargePage;