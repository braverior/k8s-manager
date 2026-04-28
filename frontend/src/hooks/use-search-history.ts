import { useCallback, useEffect, useState } from 'react';

const MAX_DEFAULT = 8;

function readStorage(key: string): string[] {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.filter((x): x is string => typeof x === 'string') : [];
  } catch {
    return [];
  }
}

export function useSearchHistory(storageKey: string, max: number = MAX_DEFAULT) {
  const [history, setHistory] = useState<string[]>(() => readStorage(storageKey));

  // 跨 Tab 同步
  useEffect(() => {
    const handler = (e: StorageEvent) => {
      if (e.key === storageKey) setHistory(readStorage(storageKey));
    };
    window.addEventListener('storage', handler);
    return () => window.removeEventListener('storage', handler);
  }, [storageKey]);

  const persist = useCallback((next: string[]) => {
    setHistory(next);
    try {
      localStorage.setItem(storageKey, JSON.stringify(next));
    } catch {
      // 忽略写入失败（如隐私模式）
    }
  }, [storageKey]);

  const add = useCallback((term: string) => {
    const trimmed = term.trim();
    if (!trimmed) return;
    setHistory((prev) => {
      const next = [trimmed, ...prev.filter((t) => t.toLowerCase() !== trimmed.toLowerCase())].slice(0, max);
      try {
        localStorage.setItem(storageKey, JSON.stringify(next));
      } catch { /* ignore */ }
      return next;
    });
  }, [storageKey, max]);

  const remove = useCallback((term: string) => {
    persist(history.filter((t) => t !== term));
  }, [history, persist]);

  const clear = useCallback(() => {
    persist([]);
  }, [persist]);

  return { history, add, remove, clear };
}
