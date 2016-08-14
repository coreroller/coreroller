import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import { Link } from "react-router"
import _ from "underscore"
import GroupExtended from "../Groups/ItemExtended.react"

class GroupLayout extends React.Component {

 constructor(props) {
    super(props);
    this.onChange = this.onChange.bind(this);

    let appID = props.params.appID,
        groupID = props.params.groupID
    this.state = {
      appID: appID,
      groupID: groupID,
      applications: applicationsStore.getCachedApplications()
    }
  }

  componentWillMount() {
    applicationsStore.getApplication(this.props.params.appID)
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
    let appName = "",
        groupName = ""

    let applications = this.state.applications ? this.state.applications : [],
        application = _.findWhere(applications, {id: this.state.appID})

    if (application) {
      appName = application.name

      let group = _.findWhere(application.groups, {id: this.state.groupID})
      if (group) {
        groupName = group.name
      }
    }

    return(
      <div className="container">
        <ol className="breadcrumb">
          <li><Link to="/">Applications</Link></li>
          <li><Link to="ApplicationLayout" params={{appID: this.state.appID}}>
            {appName}
          </Link></li>
          <li className="active">{groupName}</li>
        </ol>

        <GroupExtended appID={this.state.appID} groupID={this.state.groupID} />
     </div>
    )
  }

}

export default GroupLayout
