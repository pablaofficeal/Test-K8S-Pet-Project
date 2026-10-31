import { useState } from 'react'
import TextConfig from "../../../lib/configs/site/text.configs"

function Remember() {
    const [checked, setChecked] = useState(true)

    return (
        <div className="auth-remember">
            <label className="auth-check">
                <input type="checkbox" checked={checked} onChange={(event) => setChecked(event.target.checked)} />
                <span className="auth-check-box" aria-hidden="true" />
                <span>{TextConfig.LoginRemember}</span>
            </label>
            <button type="button" className="auth-link-button">{TextConfig.LoginForgot}</button>
        </div>
    )
}
export default Remember