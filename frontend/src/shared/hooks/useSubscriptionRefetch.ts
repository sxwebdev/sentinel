import { Code, ConnectError } from "@connectrpc/connect";
import { useEffect, useRef } from "react";

export type SubscribeFn<T> = (
  signal: AbortSignal,
) => AsyncIterable<T> | Promise<AsyncIterable<T>>;

export interface UseSubscriptionRefetchOptions<T> {
  // function to call to refetch data; kept in ref outside deps
  refetch: () => Promise<unknown> | unknown;
  // function that returns an async iterable subscription; receives AbortSignal
  subscribe: SubscribeFn<T>;
  // debounce window for trailing refetches
  debounceTimeout?: number;
  // optional error handler
  onErrorFn?: (err: unknown) => void;
  // optional data handler called for every event received
  onDataFn?: (event: T) => void;
  // external ref to control start/stop if needed (optional)
  ref?: React.RefObject<{ abort: () => void } | null>;
}

/**
 * Subscribes to an async-iterable stream (e.g., server-sent updates).
 * Performs immediate refetch on first event (leading) and at most one trailing refetch
 * after a cooldown of debounceTimeout ms, coalescing bursts.
 */
export function useSubscriptionRefetch<T = unknown>({
  refetch,
  subscribe,
  debounceTimeout = 300,
  onErrorFn,
  onDataFn,
  ref,
}: UseSubscriptionRefetchOptions<T>) {
  const refetchRef = useRef(refetch);
  refetchRef.current = refetch;

  useEffect(() => {
    const abort = new AbortController();
    let inCooldown = false;
    let scheduleTimer: number | undefined;
    let trailingNeeded = false;

    const scheduleRefetch = () => {
      if (!inCooldown) {
        // leading
        void refetchRef.current?.();
        inCooldown = true;
        scheduleTimer = window.setTimeout(() => {
          const shouldRunTrailing = trailingNeeded;
          trailingNeeded = false;
          inCooldown = false;
          if (shouldRunTrailing) {
            void refetchRef.current?.();
          }
        }, debounceTimeout);
      } else {
        // request single trailing
        trailingNeeded = true;
      }
    };

    const run = async () => {
      try {
        const iterable = await subscribe(abort.signal);
        for await (const event of iterable) {
          // invoke data handler if provided (non-blocking)
          try {
            onDataFn?.(event as T);
          } catch (handlerErr) {
            onErrorFn?.(handlerErr);
          }
          scheduleRefetch();
        }
      } catch (e) {
        if (e instanceof ConnectError) {
          if (e.code === Code.Aborted || e.code === Code.Canceled) return;
        }
        onErrorFn?.(e);
      }
    };

    // expose abort via external ref if provided
    if (ref) {
      ref.current = { abort: () => abort.abort() };
    }

    run();

    return () => {
      if (scheduleTimer) window.clearTimeout(scheduleTimer);
      abort.abort();
      if (ref) ref.current = null;
    };
    // subscribe and debounceTimeout are stable inputs from caller context
    // refetch is read from ref to avoid deps churn
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [subscribe, debounceTimeout]);
}
