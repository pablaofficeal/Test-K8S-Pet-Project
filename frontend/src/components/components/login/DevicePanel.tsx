import TextConfig from "../../../lib/configs/site/text.configs"

function DevicePanel() {
    const hashes = [TextConfig.LoginHash1, TextConfig.LoginHash2, TextConfig.LoginHash3, TextConfig.LoginHash4]

    return (
        <aside className="auth-device">
            <section>
                <header>
                    <span>{TextConfig.LoginPanelTitle}</span>
                    <span className="auth-device-lock">
                        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><rect x="4" y="11" width="16" height="10" rx="2" /><path d="M8 11V7a4 4 0 0 1 8 0v4" /></svg>
                        {TextConfig.LoginPanelEncrypted}
                    </span>
                </header>
                <ul>
                    {hashes.map((hash) => (
                        <li key={hash}>
                            <span className="auth-device-box" aria-hidden="true" />
                            <code>{hash}</code>
                            <i aria-hidden="true" />
                        </li>
                    ))}
                </ul>
            </section>
            <p>{TextConfig.LoginPanelNote}</p>
        </aside>
    )
}
export default DevicePanel