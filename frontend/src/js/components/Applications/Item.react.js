import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import Router, { Link } from "react-router"
import { Row, Col, OverlayTrigger, Button, Popover } from "react-bootstrap"
import GroupsList from "./ApplicationItemGroupsList.react"
import ChannelsList from "./ApplicationItemChannelsList.react"
import ConfirmationContent from "../Common/ConfirmationContent.react"

class Item extends React.Component {

  constructor(props) {
    super(props)
    this.updateApplication = this.updateApplication.bind(this)
    this.deleteApplication = this.deleteApplication.bind(this)
  }

  static propTypes: {
    application: React.PropTypes.object.isRequired,
    handleUpdateApplication: React.PropTypes.func.isRequired
  }

  updateApplication() {
    this.props.handleUpdateApplication(this.props.application.id)
  }

  deleteApplication() {
    let confirmationText = "Are you sure you want to delete this application?"
    if (confirm(confirmationText)) {
      applicationsStore.deleteApplication(this.props.application.id)
    }
  }

  render() {
    let application = this.props.application ? this.props.application : {},
        description = this.props.application.description ? this.props.application.description : "No description provided",
        styleDescription = this.props.application.description ? "" : " italicText",
        channels = this.props.application.channels ? this.props.application.channels : [],
        groups = this.props.application.groups ? this.props.application.groups : [],
        instances = this.props.application.instances ? this.props.application.instances : 0,
        appID = this.props.application ? this.props.application.id : "",
        popoverContent = {
          type: "application",
          appID: appID
        }

    return(
      <div className="apps--box">
        <Row className="apps--boxHeader">
          <Col xs={10}>
            <h3 className="apps--boxTitle">
              <Link to="ApplicationLayout" params={{appID}}>
                {this.props.application.name} <i className="fa fa-caret-right"></i>
              </Link>
              <span className="apps--id">(ID: {appID})</span>
            </h3>
            <span className={"apps--description" + styleDescription}>{description}</span>
          </Col>
          <Col xs={2}>
            <div className="apps--buttons">
              <button className="cr-button displayInline fa fa-edit" onClick={this.updateApplication}></button>
              <button className="cr-button displayInline fa fa-trash-o" onClick={this.deleteApplication}></button>
            </div>
          </Col>
        </Row>
        <div className="apps--boxContent">
          <Row className="apps--resume">
            <Col xs={2}>
              <span className="subtitle">
                Instances:
              </span>
              <span>
                {instances}
              </span>
            </Col>
            <Col xs={10}>
              <div className="divider abs-divider">|</div>
              <span className="subtitle">
                Groups:
              </span>
              <span>
                {groups.length}
              </span>
              <GroupsList
                groups={groups}
                appID={this.props.application.id}
                appName={this.props.application.name} />
            </Col>
          </Row>
          <Row className="apps--resume">
            <Col xs={1}>
              <span className="subtitle">Channels:</span>
            </Col>
            <Col xs={11}>
              <ChannelsList channels={channels} />
            </Col>
          </Row>
        </div>
      </div>
    )
  }

}

export default Item