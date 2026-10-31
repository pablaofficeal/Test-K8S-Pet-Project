import Logos from '../../../public/logo.webp'

function Logo() {
    return (
        <img src={Logos} alt="Telegram Bot Bilders" 
        width={64} 
        style={
            { 
                height: 'auto',
                margin: '0 auto',
                position: 'relative',
                top: 0,
                left: 0,
            }
        } 
        />
    )
}
export default Logo
