import { StrictMode, Component, ErrorInfo, ReactNode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';

console.log('=== MAIN-STEP2: Starting ===');

// Error Boundary Component
class ErrorBoundary extends Component<
  { children: ReactNode },
  { hasError: boolean; error: Error | null; errorInfo: ErrorInfo | null }
> {
  constructor(props: { children: ReactNode }) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error) {
    console.error('ErrorBoundary caught error:', error);
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary componentDidCatch:', error, errorInfo);
    this.setState({ error, errorInfo });
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{ padding: '40px', fontFamily: 'Arial, sans-serif', color: 'red' }}>
          <h1>❌ Error Loading App</h1>
          <h2>Error:</h2>
          <pre style={{ background: '#f5f5f5', padding: '20px', overflow: 'auto' }}>
            {this.state.error?.toString()}
          </pre>
          <h2>Component Stack:</h2>
          <pre style={{ background: '#f5f5f5', padding: '20px', overflow: 'auto' }}>
            {this.state.errorInfo?.componentStack}
          </pre>
        </div>
      );
    }

    return this.props.children;
  }
}

const rootElement = document.getElementById('root');
if (!rootElement) {
  console.error('Root element not found!');
  throw new Error('Root element not found');
}

console.log('Root element found, attempting to load App.tsx...');

// Try to import App
let App;
try {
  console.log('Importing App...');
  const AppModule = await import('./App.tsx');
  App = AppModule.default;
  console.log('App imported successfully:', App);
} catch (error) {
  console.error('ERROR importing App.tsx:', error);
  rootElement.innerHTML = `
    <div style="color: red; padding: 40px; font-family: Arial, sans-serif;">
      <h1>❌ Error Importing App.tsx</h1>
      <pre style="background: #f5f5f5; padding: 20px; overflow: auto;">${error}</pre>
    </div>
  `;
  throw error;
}

try {
  console.log('Creating root and rendering App...');
  const root = createRoot(rootElement);
  
  root.render(
    <StrictMode>
      <ErrorBoundary>
        <App />
      </ErrorBoundary>
    </StrictMode>
  );
  
  console.log('=== MAIN-STEP2: Render complete ===');
} catch (error) {
  console.error('ERROR rendering App:', error);
  rootElement.innerHTML = `
    <div style="color: red; padding: 40px; font-family: Arial, sans-serif;">
      <h1>❌ Error Rendering App</h1>
      <pre style="background: #f5f5f5; padding: 20px; overflow: auto;">${error}</pre>
    </div>
  `;
}
