import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';

console.log('=== MAIN-STEP1: Starting ===');

const rootElement = document.getElementById('root');
if (!rootElement) {
  console.error('Root element not found!');
  throw new Error('Root element not found');
}

console.log('Root element found');

// Step 1: Test with minimal App component (no imports from App.tsx yet)
function MinimalApp() {
  console.log('MinimalApp rendering');
  return (
    <div style={{ padding: '40px', fontFamily: 'Arial, sans-serif' }}>
      <h1 style={{ color: '#10b981' }}>✅ RechargeMax - Step 1 Working!</h1>
      <p>React is rendering successfully.</p>
      <p>Next: Load actual App.tsx</p>
    </div>
  );
}

try {
  console.log('Creating root...');
  const root = createRoot(rootElement);
  console.log('Root created, rendering...');
  
  root.render(
    <StrictMode>
      <MinimalApp />
    </StrictMode>
  );
  
  console.log('=== MAIN-STEP1: Render complete ===');
} catch (error) {
  console.error('ERROR in main-step1:', error);
  rootElement.innerHTML = `<div style="color: red; padding: 20px;"><h1>Error</h1><pre>${error}</pre></div>`;
}
