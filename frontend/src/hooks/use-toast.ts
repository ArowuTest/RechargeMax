import { useState, useCallback } from 'react';

export interface Toast {
  id: string;
  title?: string;
  description?: string;
  variant?: 'default' | 'destructive';
  duration?: number;
}

interface ToastState {
  toasts: Toast[];
}

let toastCount = 0;
let globalToastState: ToastState = { toasts: [] };
let globalSetState: ((state: ToastState) => void) | null = null;

// Standalone toast function that can be called outside of React components
export const toast = ({ title, description, variant = 'default', duration = 5000 }: Omit<Toast, 'id'>) => {
  const id = (++toastCount).toString();
  const newToast: Toast = { id, title, description, variant, duration };

  globalToastState = {
    toasts: [...globalToastState.toasts, newToast]
  };

  if (globalSetState) {
    globalSetState(globalToastState);
  }

  // Auto remove toast after duration
  if (duration > 0) {
    setTimeout(() => {
      globalToastState = {
        toasts: globalToastState.toasts.filter(t => t.id !== id)
      };
      if (globalSetState) {
        globalSetState(globalToastState);
      }
    }, duration);
  }

  return id;
};

export const useToast = () => {
  const [state, setState] = useState<ToastState>(globalToastState);

  // Register this component's setState as the global one
  useState(() => {
    globalSetState = setState;
    return () => {
      globalSetState = null;
    };
  });

  const toastFn = useCallback(({ title, description, variant = 'default', duration = 5000 }: Omit<Toast, 'id'>) => {
    return toast({ title, description, variant, duration });
  }, []);

  const dismiss = useCallback((toastId?: string) => {
    globalToastState = {
      toasts: toastId 
        ? globalToastState.toasts.filter(t => t.id !== toastId)
        : []
    };
    setState(globalToastState);
  }, []);

  return {
    toast: toastFn,
    dismiss,
    toasts: state.toasts
  };
};
