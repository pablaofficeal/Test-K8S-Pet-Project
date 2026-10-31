import enTranslations from '../meta/en.json'
import ruTranslations from '../meta/ru.json'


export type SupportedLangCode = 'en' | 'ru'

export const SUPPORTED_LANGUAGES: { code: SupportedLangCode; label: string; short: string }[] = [
  { code: 'en', label: 'English', short: 'EN' },
  { code: 'ru', label: 'Русский', short: 'RU' },
]

const translations: Record<string, Record<string, string>> = {
  en: enTranslations,
  ru: ruTranslations,
}

export function isSupportedLang(lang: string): lang is SupportedLangCode {
  return SUPPORTED_LANGUAGES.some((l) => l.code === lang)
}

const DEFAULT_LANG = 'en'
const LS_LANG_KEY = 'Telegram-Bot-Bilder-Lang'

function getLocalStorageLang(): string | null {
  if (typeof window !== 'undefined' && typeof localStorage !== 'undefined') {
    try {
      const v = localStorage.getItem(LS_LANG_KEY)
      if (v) return v
    } catch {
    }
  }
  return null
}

function setLocalStorageLang(lang: string): void {
  if (typeof window !== 'undefined' && typeof localStorage !== 'undefined') {
    try {
      localStorage.setItem(LS_LANG_KEY, lang)
    } catch {
    }
  }
}

export function setLanguageCookie(lang: string): void {
  if (typeof document !== 'undefined') {
    const maxAge = 60 * 60 * 24 * 365
    document.cookie = `lang=${encodeURIComponent(lang)}; max-age=${maxAge}; path=/; SameSite=Lax`;
  }
  setLocalStorageLang(lang)
}

export function getLanguageCookie(serverCookieValue?: string): string {
  if (typeof document === 'undefined') {
    return serverCookieValue || DEFAULT_LANG;
  }
  const matches = document.cookie.match(
    new RegExp('(?:^|; )' + 'lang'.replace(/([.$?*|{}()[\]\\/+^])/g, '\\$1') + '=([^;]*)')
  );
  return matches ? decodeURIComponent(matches[1]) : DEFAULT_LANG;
}

export function getCurrentLanguage(serverCookieValue?: string): string {
  const ls = getLocalStorageLang()
  if (ls && translations[ls]) return ls
  const cookieLang = getLanguageCookie(serverCookieValue);
  return translations[cookieLang] ? cookieLang : DEFAULT_LANG;
}

function normalizeLangTag(tag: string): string {
  return tag.toLowerCase().split('-')[0].split('_')[0]
}

export function detectBrowserLanguage(): SupportedLangCode {
  if (typeof navigator === 'undefined') return DEFAULT_LANG as SupportedLangCode
  const candidates: string[] = []
  if (Array.isArray(navigator.languages)) candidates.push(...navigator.languages)
  if (navigator.language) candidates.push(navigator.language)
  for (const tag of candidates) {
    const base = normalizeLangTag(tag)
    if (isSupportedLang(base)) return base
  }
  return DEFAULT_LANG as SupportedLangCode
}

export function ensureAutoLanguage(): string {
  const stored = getLocalStorageLang()
  if (stored && translations[stored]) return stored
  const cookie = getLanguageCookie()
  if (translations[cookie]) {
    setLocalStorageLang(cookie)
    return cookie
  }
  const detected = detectBrowserLanguage()
  setLanguageCookie(detected)
  return detected
}

export function t(key: string, serverCookieValue?: string): string {
  const lang = getCurrentLanguage(serverCookieValue);
  const currentDict = translations[lang] || translations[DEFAULT_LANG];
  const fallbackDict = translations[DEFAULT_LANG];

  return currentDict[key] || fallbackDict[key] || key;
}

export function changeLanguage(lang: string): void {
  if (translations[lang]) {
    setLanguageCookie(lang);
    if (typeof window !== 'undefined') {
      window.location.reload();
    }
  }
}