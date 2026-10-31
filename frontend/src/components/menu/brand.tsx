import { Link } from 'react-router-dom'
import TextConfig from '../../lib/configs/site/text.configs'

function Brand() {
    return (
        <Link to="/" className="brand" aria-label={TextConfig.title}>
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><rect x="3" y="3" width="11" height="11" rx="3" /><rect x="10" y="10" width="11" height="11" rx="3" /></svg>
            <span>{TextConfig.title}</span>
        </Link>
    )
}
export default Brand