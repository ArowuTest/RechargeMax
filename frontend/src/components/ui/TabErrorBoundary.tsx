import React from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';

interface Props {
  tabName: string;
  children: React.ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

/**
 * TabErrorBoundary — wraps each admin tab.
 * If a tab throws during render, it shows an isolated error card instead of
 * crashing the entire portal. The user can click "Try Again" to reset.
 */
export class TabErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  override componentDidCatch(error: Error, info: React.ErrorInfo) {
    // Log to console so it's visible in Render logs
    console.error(`[TabErrorBoundary] Tab "${this.props.tabName}" crashed:`, error, info.componentStack);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
  };

  override render() {
    if (this.state.hasError) {
      return (
        <div className="flex flex-col items-center justify-center py-16 px-6 text-center">
          <div className="w-14 h-14 bg-red-100 rounded-full flex items-center justify-center mb-4">
            <AlertTriangle className="w-7 h-7 text-red-500" />
          </div>
          <h3 className="text-lg font-semibold text-gray-800 mb-1">
            Something went wrong in <span className="text-red-600">{this.props.tabName}</span>
          </h3>
          <p className="text-sm text-gray-500 mb-2 max-w-md">
            This tab encountered an unexpected error. The rest of the portal is still working.
          </p>
          {this.state.error && (
            <pre className="text-xs text-red-400 bg-red-50 border border-red-100 rounded-md px-4 py-2 mb-5 max-w-lg overflow-auto text-left">
              {this.state.error.message}
            </pre>
          )}
          <button
            onClick={this.handleReset}
            className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors"
          >
            <RefreshCw className="w-4 h-4" />
            Try Again
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}
