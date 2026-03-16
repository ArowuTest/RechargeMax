import { useState, useCallback, useEffect } from 'react';

export interface Toast {
  id: string;
  title?: string;
  description?: string;
  variant?: 'default' | 'destructive';
  duration?: number;
  action?: React.ReactNode;
}

interface ToastState {
  toasts: Toast[];
}

let toastCount = 0;
let globalToastState: ToastState = { toasts: [] };

// Subscriber set — every mounted useToast() consumer registers its setState here.
// This fixes the race where the last-mounted component overwrote the single
// globalSetState reference, causing the <Toaster /> component to never re-render.
const subscribers = new Set<(state: ToastState) => void>();

function notifyAll(next: ToastState) {
  globalToastState = next;
  subscribers.forEach(fn => fn(next));
}

// Standalone toast function — safe to call from anywhere (callbacks, effects, etc.)
export const toast = ({
  title,
  description,
  variant = 'default',
  duration = 5000,
}: Omit<Toast, 'id'>) => {
  const id = (++toastCount).toString();
  const newToast: Toast = { id, title, description, variant, duration };

  notifyAll({ toasts: [...globalToastState.toasts, newToast] });

  if (duration > 0) {
    setTimeout(() => {
      notifyAll({ toasts: globalToastState.toasts.filter(t => t.id !== id) });
    }, duration);
  }

  return id;
};

export const useToast = () => {
  const [state, setState] = useState<ToastState>(globalToastState);

  // Register this component's setState as a subscriber on mount;
  // unregister on unmount. Multiple components (Toaster + page components)
  // can all subscribe simultaneously — the Set ensures each fires once.
  useEffect(() => {
    subscribers.add(setState);
    // Sync to latest state in case toasts were emitted before this component mounted.
    setState(globalToastState);
    return () => {
      subscribers.delete(setState);
    };
  }, []);

  const toastFn = useCallback(
    ({
      title,
      description,
      variant = 'default',
      duration = 5000,
    }: Omit<Toast, 'id'>) => toast({ title, description, variant, duration }),
    [],
  );

  const dismiss = useCallback((toastId?: string) => {
    notifyAll({
      toasts: toastId
        ? globalToastState.toasts.filter(t => t.id !== toastId)
        : [],
    });
  }, []);

  return { toast: toastFn, dismiss, toasts: state.toasts };
};
