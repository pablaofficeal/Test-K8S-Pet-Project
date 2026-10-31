import { Link } from 'react-router-dom'
import DashboardText from './dashboardText'


function DashboardLinks() {
    return (
        <Link to="/users/dashboard/index.php">
            <DashboardText />
        </Link>
    )
}
export default DashboardLinks