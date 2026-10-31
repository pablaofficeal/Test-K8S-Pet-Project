import { Link } from 'react-router-dom'
import DashboardText from './dashboardText'
import { ROUTES } from '../../lib/configs/site/routes'


function DashboardLinks() {
    return (
        <Link to={ROUTES.DASHBOARD}>
            <DashboardText />
        </Link>
    )
}
export default DashboardLinks