import { Link } from 'react-router-dom'
import Devices from './devices'
import TextConfig from '../../../lib/configs/site/text.configs'

function Hero() {
    return (
        <section className="home-hero">
            <div className="home-hero-text">
                <h1>{TextConfig.HomeHeroTitle}</h1>
                <p className="home-hero-lead">{TextConfig.HomeHeroText}</p>
                <div className="home-hero-actions">
                    <Link to="/Start/create/my/account/register.php" className="home-button home-button-primary">
                        {TextConfig.CreateAccount}
                    </Link>
                    <button type="button" className="home-button home-button-secondary">
                        {TextConfig.HomeHeroSync}
                    </button>
                </div>
                <p className="home-hero-note">{TextConfig.HomeHeroNote}</p>
            </div>
            <Devices />
        </section>
    )
}
export default Hero