import TextConfig from '../../../lib/configs/site/text.configs'

function Features() {
    return (
        <section id="security" className="home-features">
            <article>
                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><rect x="4" y="10" width="16" height="11" rx="3" /><path d="M8 10V7a4 4 0 0 1 8 0v3" /></svg>
                <h2>{TextConfig.HomeFeatureEncTitle}</h2>
                <p>{TextConfig.HomeFeatureEncText}</p>
            </article>
            <article>
                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><rect x="3" y="3" width="18" height="8" rx="3" /><rect x="3" y="13" width="18" height="8" rx="3" /><path d="M7 7h.01M7 17h.01" /></svg>
                <h2>{TextConfig.HomeFeatureEuTitle}</h2>
                <p>{TextConfig.HomeFeatureEuText}</p>
            </article>
            <article>
                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d="M4 6h12M4 12h8M4 18h8M15 16l2 2 4-4" /></svg>
                <h2>{TextConfig.HomeFeatureAuditTitle}</h2>
                <p>{TextConfig.HomeFeatureAuditText}</p>
            </article>
        </section>
    )
}
export default Features