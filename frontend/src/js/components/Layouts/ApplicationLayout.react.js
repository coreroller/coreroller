import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import _ from "underscore"
import { Link } from "react-router"
import ApplicationsList from "../Applications/List.react"
import GroupsList from "../Groups/List.react"
import ChannelsList from "../Channels/List.react"
import PackagesList from "../Packages/List.react"

class ApplicationLayout extends React.Component {

  constructor(props) {
    super(props);
    this.onChange = this.onChange.bind(this);

    let appID = props.params.appID
    this.state = {
      appID: appID,
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
    let appName = ""
    let applications = this.state.applications ? this.state.applications : [],
        application = _.findWhere(applications, {id: this.state.appID})

    if (application) {
      appName = application.name
    }

    return(
      <div className="container">
        <ol className="breadcrumb">
          <li><Link to="MainLayout">Applications</Link></li>
          <li className="active">{appName}</li>
        </ol>
        <Row>
          <GroupsList appID={this.state.appID} />
          <Col xs={4} className="group--info">
            <Row>
              <Col xs={1}></Col>
              <Col className={11}>
                <ChannelsList appID={this.state.appID} />
                <hr />
                <PackagesList appID={this.state.appID} />
              </Col>
            </Row>
          </Col>
        </Row>
      </div>
    )
  }

}

export default ApplicationLayout
