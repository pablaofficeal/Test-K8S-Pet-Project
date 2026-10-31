import TextConfig from '../../lib/configs/site/text.configs'

function NavItems() {
    return (
        <>
            <a href="#product">{TextConfig.NavProduct}</a>
            <a href="#security">{TextConfig.NavSecurity}</a>
        </>
    )
}
export default NavItems