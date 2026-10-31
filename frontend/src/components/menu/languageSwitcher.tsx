import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import TextConfig from '../../lib/configs/site/text.configs'
import { SUPPORTED_LANGUAGES, changeLanguage, getCurrentLanguage } from '../../lib/i18n/i18n'
import '../../assets/css/language.css'

function LanguageSwitcher() {
    const [open, setOpen] = useState(false)
    const current = getCurrentLanguage()
    const currentShort = SUPPORTED_LANGUAGES.find((lang) => lang.code === current)?.short

    useEffect(() => {
        if (!open) {
            return
        }

        function handleKeyDown(event: KeyboardEvent) {
            if (event.key === 'Escape') {
                setOpen(false)
            }
        }

        document.addEventListener('keydown', handleKeyDown)
        return () => document.removeEventListener('keydown', handleKeyDown)
    }, [open])

    return (
        <>
            <button type="button" className="language-button" onClick={() => setOpen(true)} title={TextConfig.LanguageTitle}>
                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="10" /><path d="M2 12h20M12 2a15 15 0 0 1 0 20M12 2a15 15 0 0 0 0 20" /></svg>
                <span>{currentShort}</span>
            </button>
            {open && createPortal(
                <div className="language-overlay" onClick={() => setOpen(false)}>
                    <div
                        className="language-modal"
                        role="dialog"
                        aria-modal="true"
                        aria-labelledby="language-title"
                        onClick={(event) => event.stopPropagation()}
                    >
                        <div className="language-modal-head">
                            <h2 id="language-title">{TextConfig.LanguageTitle}</h2>
                            <button type="button" className="language-close" onClick={() => setOpen(false)} aria-label={TextConfig.LanguageClose} title={TextConfig.LanguageClose}>
                                <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M18 6 6 18M6 6l12 12" /></svg>
                            </button>
                        </div>
                        <ul className="language-list">
                            {SUPPORTED_LANGUAGES.map((lang) => (
                                <li key={lang.code}>
                                    <button
                                        type="button"
                                        className={lang.code === current ? 'language-option active' : 'language-option'}
                                        onClick={() => changeLanguage(lang.code)}
                                    >
                                        <span>{lang.label}</span>
                                        <span>{lang.short}</span>
                                    </button>
                                </li>
                            ))}
                        </ul>
                    </div>
                </div>,
                document.body,
            )}
        </>
    )
}
export default LanguageSwitcher