import TextConfig from '../../../lib/configs/site/text.configs'

type CheckProps = {
    done?: boolean
}

function Check({ done = false }: CheckProps) {
    return (
        <span className={done ? 'home-check done' : 'home-check'} aria-hidden="true">
            {done && <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" strokeWidth="3.2" strokeLinecap="round" strokeLinejoin="round"><path d="m5 12 5 5 9-10" /></svg>}
        </span>
    )
}

function Devices() {
    return (
        <div className="home-devices">
            <section className="home-device home-device-laptop">
                <header>
                    <span>{TextConfig.HomeDeviceLaptop}</span>
                    <span className="home-status"><i />{TextConfig.HomeStatusUpdated}</span>
                </header>
                <ul>
                    <li>
                        <Check />
                        <div>
                            <p>{TextConfig.HomeTaskEstimate}</p>
                            <small>{TextConfig.HomeTaskEstimateWhen}</small>
                        </div>
                        <b>{TextConfig.HomeAssigneeAk}</b>
                    </li>
                    <li>
                        <Check />
                        <div>
                            <p>{TextConfig.HomeTaskAccess}</p>
                            <small>{TextConfig.HomeTaskAccessWhen}</small>
                        </div>
                        <b>{TextConfig.HomeAssigneeMd}</b>
                    </li>
                    <li className="done">
                        <Check done />
                        <div>
                            <p>{TextConfig.HomeTaskReport}</p>
                        </div>
                        <b>{TextConfig.HomeAssigneeOp}</b>
                    </li>
                    <li>
                        <Check />
                        <div>
                            <p>{TextConfig.HomeTaskActs}</p>
                            <small>{TextConfig.HomeTaskActsWhen}</small>
                        </div>
                        <b>{TextConfig.HomeAssigneeAk}</b>
                    </li>
                </ul>
            </section>
            <section className="home-device home-device-tablet">
                <header>
                    <span>{TextConfig.HomeDeviceTablet}</span>
                    <span className="home-status"><i />{TextConfig.HomeStatusJustNow}</span>
                </header>
                <ul>
                    <li>
                        <Check />
                        <div>
                            <p>{TextConfig.HomeTaskEstimate}</p>
                        </div>
                    </li>
                    <li>
                        <Check />
                        <div>
                            <p>{TextConfig.HomeTaskAccess}</p>
                        </div>
                    </li>
                </ul>
            </section>
            <section className="home-device home-device-phone">
                <header>
                    <span>{TextConfig.HomeDevicePhone}</span>
                    <span className="home-status"><i /></span>
                </header>
                <ul>
                    <li>
                        <Check />
                        <div>
                            <p>{TextConfig.HomeTaskEstimateShort}</p>
                        </div>
                    </li>
                </ul>
                <footer>{TextConfig.HomeMoreTasks}</footer>
            </section>
        </div>
    )
}
export default Devices