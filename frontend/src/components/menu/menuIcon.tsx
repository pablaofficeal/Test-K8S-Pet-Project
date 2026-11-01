type MenuIconProps = {
    paths: string[]
}

function MenuIcon({ paths }: MenuIconProps) {
    return (
        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            {paths.map((path) => <path key={path} d={path} />)}
        </svg>
    )
}
export default MenuIcon