import { Link } from 'react-router-dom'
import TextConfig from "../../../lib/configs/site/text.configs"
import { ROUTES } from "../../../lib/configs/site/routes"

function VerifyBanner() {
    return (
        <div className="verify-banner" role="status">
            <span>{TextConfig.VerifyBannerText}</span>
            <Link to={ROUTES.VERIFY}>{TextConfig.VerifyBannerAction}</Link>
        </div>
    )
}
export default VerifyBanner