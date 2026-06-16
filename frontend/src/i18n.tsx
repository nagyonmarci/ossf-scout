import { createContext, useContext, useState, useCallback, type ReactNode } from 'react';
import en from './locales/en';
import hu from './locales/hu';

export type Lang = 'en' | 'hu';
export type Dict = typeof en;

const dictionaries: Record<Lang, Dict> = { en, hu };

const STORAGE_KEY = 'ossf-scout-lang';

function detectInitialLang(): Lang {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored === 'en' || stored === 'hu') return stored;
  return navigator.language.toLowerCase().startsWith('hu') ? 'hu' : 'en';
}

type TranslateFn = (key: keyof Dict, params?: Record<string, string | number>) => string;

const LangContext = createContext<{ lang: Lang; setLang: (l: Lang) => void; t: TranslateFn } | null>(null);

function interpolate(template: string, params?: Record<string, string | number>): string {
  if (!params) return template;
  return template.replace(/\{(\w+)\}/g, (match, key) => (key in params ? String(params[key]) : match));
}

export function LangProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(detectInitialLang);

  const setLang = useCallback((l: Lang) => {
    localStorage.setItem(STORAGE_KEY, l);
    setLangState(l);
  }, []);

  const t = useCallback<TranslateFn>(
    (key, params) => interpolate(dictionaries[lang][key] ?? dictionaries.en[key] ?? key, params),
    [lang],
  );

  return <LangContext.Provider value={{ lang, setLang, t }}>{children}</LangContext.Provider>;
}

export function useLang() {
  const ctx = useContext(LangContext);
  if (!ctx) throw new Error('useLang must be used within LangProvider');
  return ctx;
}
