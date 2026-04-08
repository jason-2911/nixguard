import { useEffect, useRef } from 'react';

/**
 * Polls a callback at a fixed interval while the component is mounted
 * and the browser tab is visible. Skips ticks when the tab is hidden
 * to avoid wasting bandwidth.
 */
export function useAutoRefresh(callback: () => void, intervalMs = 5000) {
  const savedCallback = useRef(callback);

  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  useEffect(() => {
    let id: ReturnType<typeof setInterval>;

    const start = () => {
      id = setInterval(() => savedCallback.current(), intervalMs);
    };

    const handleVisibility = () => {
      clearInterval(id);
      if (!document.hidden) {
        savedCallback.current(); // fetch immediately when tab becomes visible
        start();
      }
    };

    start();
    document.addEventListener('visibilitychange', handleVisibility);

    return () => {
      clearInterval(id);
      document.removeEventListener('visibilitychange', handleVisibility);
    };
  }, [intervalMs]);
}
