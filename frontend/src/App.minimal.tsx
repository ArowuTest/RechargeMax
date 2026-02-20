import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <div style={{padding: '20px', fontFamily: 'Arial'}}>
        <h1>RechargeMax - Minimal Test</h1>
        <p>If you see this, React is working!</p>
      </div>
    </QueryClientProvider>
  );
}

export default App;
