import TextConfig from '../../lib/configs/site/text.configs'

function NavItems() {
    const items = [
        TextConfig.NavProduct,
        TextConfig.NavSecurity,
        TextConfig.NavTeams,
        TextConfig.NavPricing,
        TextConfig.NavDocs,
    ]

    return (
        <>
            {items.map((item) => (
                <span key={item}>{item}</span>
            ))}
        </>
    )
}
export default NavItems