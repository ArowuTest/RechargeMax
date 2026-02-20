import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

console.log('=== MAIN-DEBUG.TSX STARTING ===');

// Test 1: Check if root element exists
const rootElement = document.getElementById('root');
console.log('Root element:', rootElement);

if (!rootElement) {
  console.error('ERROR: Root element not found!');
  document.body.innerHTML = '<div style="color: red; padding: 20px;">ERROR: Root element not found!</div>';
} else {
  console.log('Root element found, attempting to render...');
  
  try {
    // Test 2: Try minimal render
    const root = createRoot(rootElement);
    console.log('Root created successfully');
    
    root.render(
      <StrictMode>
        <div style={{ padding: '20px', fontFamily: 'Arial' }}>
          <h1 style={{ color: 'green' }}>✅ React is Working!</h1>
          <p>If you see this, React is rendering correctly.</p>
          <p>Timestamp: {new Date().toISOString()}</p>
        </div>
      </StrictMode>
    );
    console.log('Render called successfully');
  } catch (error) {
    console.error('ERROR during render:', error);
    document.body.innerHTML = `<div style="color: red; padding: 20px;">
      <h1>ERROR during render</h1>
      <pre>${error}</pre>
    </div>`;
  }
}

console.log('=== MAIN-DEBUG.TSX COMPLETED ===');
