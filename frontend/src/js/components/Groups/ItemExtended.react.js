import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import Router, { Link } from "react-router"
import Switch from "rc-switch"
import _ from "underscore"
import ChannelLabel from "../Common/ChannelLabel.react"
import InstancesContainer from "../Instances/Container.react"
import VersionBreakdown from "../Common/VersionBreakdown.react"

class ItemExtended extends React.Component {

  constructor() {
    super()
    this.onChange = this.onChange.bind(this)

    this.state = {applications: applicationsStore.getCachedApplications()}
  }

  static propTypes: {
    appID: React.PropTypes.string.isRequired,
    groupID: React.PropTypes.string.isRequired
  }

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      applications: applicationsStore.getCachedApplications()
    })
  }

  render() {
    let application = _.findWhere(this.state.applications, {id: this.props.appID})
    let group = application ? _.findWhere(application.groups, {id: this.props.groupID}) : null

    let name = "",
        groupId = "",
        description = "",
        instancesNum = 0,
        policyMaxUpdatesPerDay = 0,
        policyPeriodInterval = 0,
        channel = {},
        version_breakdown = [],
        policyUpdates,
        policyUpdatesTimeout,
        safeMode,
        officeHours,
        groupChannel,
        styleGroupChannel

    if (group) {
      name = group.name
      groupId = group.id
      description = group.description ? group.description : ""
      channel = group.channel ? group.channel : {}
      instancesNum = group.instances_stats ? group.instances_stats.total : 0
      policyMaxUpdatesPerDay = group.policy_max_updates_per_period ? group.policy_max_updates_per_period : 0
      policyPeriodInterval = group.policy_period_interval ? group.policy_period_interval : 0
      policyUpdates = group.policy_updates_enabled ? group.policy_updates_enabled : null
      policyUpdatesTimeout = group.policy_update_timeout ? group.policy_update_timeout : null
      safeMode = group.policy_safe_mode ? group.policy_safe_mode : null
      officeHours = group.policy_office_hours ? group.policy_office_hours : null
      version_breakdown = group.version_breakdown ? group.version_breakdown : []
      groupChannel = _.isEmpty(group.channel) ? "No channel provided" : <ChannelLabel channel={group.channel} />
      styleGroupChannel = _.isEmpty(group.channel) ? "italicText" : ""
    }

		return (
      <Row>
        <Col xs={12}>
          <Row>
            <Col xs={12} className="groups--container">
              <div className="groups--box">
                <Row className="groups--boxHeader">
                  <Col xs={12}>
                    <h3 className="groups--boxTitle">
                      {name}
                      <span className="groups--id">(ID: {groupId})</span>
                    </h3>
                    <span className="groups--description">{description}</span>
                  </Col>
                </Row>
                <div className="groups--boxContent">
                  <Row className="groups--resume">
                    <Col xs={12}>
                      <span className="subtitle">Instances:</span><span> {instancesNum}</span>
                      <div className="divider">|</div>
                      <span className="subtitle">Channel:</span> <span className={styleGroupChannel}>{groupChannel}</span>
                    </Col>
                  </Row>
                  <Row className="groups--resume">
                    <Col xs={12}>
                      <span className="subtitle">Rollout policy:</span> Max {policyMaxUpdatesPerDay} updates per {policyPeriodInterval} &nbsp;|&nbsp; Updates timeout { policyUpdatesTimeout }
                    </Col>
                  </Row>
                  <Row className="groups--resume">
                    <Col xs={12}>
                      <span className="subtitle displayInline">Updates enabled:</span>
                      <div className="displayInline">
                        <Switch checked={policyUpdates} disabled={true} checkedChildren={"✔"} unCheckedChildren={"✘"} />
                      </div>
                      <span className="subtitle displayInline">Only office hours:</span>
                      <div className="displayInline">
                        <Switch checked={officeHours} disabled={true} checkedChildren={"✔"} unCheckedChildren={"✘"} />
                      </div>
                      <span className="subtitle displayInline">Safe mode:</span>
                      <div className="displayInline">
                        <Switch checked={safeMode} disabled={true} checkedChildren={"✔"} unCheckedChildren={"✘"} />
                      </div>
                    </Col>
                  </Row>
                  <Row className="groups--resume">
                    <VersionBreakdown version_breakdown={version_breakdown} channel={channel} />
                  </Row>
                  {/* Instances */}
                  <InstancesContainer
                    appID={this.props.appID}
                    groupID={this.props.groupID}
                    version_breakdown={version_breakdown}
                    channel={channel} />
                </div>
              </div>
            </Col>
          </Row>
        </Col>
      </Row>
		)
  }

}

export default ItemExtended
