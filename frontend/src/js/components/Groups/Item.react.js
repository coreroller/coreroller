import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react";
import Router, { Link } from "react-router";
import { Row, Col, OverlayTrigger, Button, Popover } from "react-bootstrap";
import Switch from "rc-switch"
import _ from "underscore"
import ChannelLabel from "../Common/ChannelLabel.react"
import VersionBreakdown from "../Common/VersionBreakdown.react"
import ConfirmationContent from "../Common/ConfirmationContent.react"

class Item extends React.Component {

  constructor(props) {
    super(props)
    this.deleteGroup = this.deleteGroup.bind(this)
    this.updateGroup = this.updateGroup.bind(this)
  }

	static PropTypes: {
    group: React.PropTypes.object.isRequired,
    appName: React.PropTypes.string.isRequired,
    channels: React.PropTypes.array.isRequired,
    handleUpdateGroup: React.PropTypes.func.isRequired
}

  deleteGroup() {
    let confirmationText = "Are you sure you want to delete this group?"
    if (confirm(confirmationText)) {
      applicationsStore.deleteGroup(this.props.group.application_id, this.props.group.id)
    }
  }

  updateGroup() {
    this.props.handleUpdateGroup(this.props.group.application_id, this.props.group.id)
  }

  render() {
    let version_breakdown = (this.props.group && this.props.group.version_breakdown) ? this.props.group.version_breakdown : [],
        instances_total = this.props.group.instances_stats ? this.props.group.instances_stats.total : 0,
        description = this.props.group.description ? this.props.group.description : "No description provided",
        channel = this.props.group.channel ? this.props.group.channel : {},
        styleDescription = this.props.group.description ? "" : " italicText",
        popoverContent = {
          type: "group",
          appID: this.props.group.application_id,
          groupID: this.props.group.id
        }

    let groupChannel = _.isEmpty(this.props.group.channel) ? "No channel provided" : <ChannelLabel channel={this.props.group.channel} />
    let styleGroupChannel = _.isEmpty(this.props.group.channel) ? "italicText" : ""

    return (
      <div className="groups--box">
        <Row className="groups--boxHeader">
          <Col xs={10}>
            <h3 className="groups--boxTitle">
              <Link to="GroupLayout" params={{groupID: this.props.group.id, appID: this.props.group.application_id}}>
                {this.props.group.name} <i className="fa fa-caret-right"></i>
              </Link>
              <span className="groups--id">(ID: {this.props.group.id})</span>
            </h3>
            <span className={"groups--description" + styleDescription}>{description}</span>
          </Col>
          <Col xs={2}>
            <div className="groups--buttons">
              <button className="cr-button displayInline fa fa-edit" onClick={this.updateGroup}></button>
              <button className="cr-button displayInline fa fa-trash-o" onClick={this.deleteGroup}></button>
            </div>
          </Col>
        </Row>
        <div className="groups--boxContent">
          <Row className="groups--resume">
            <Col xs={12}>
              <span className="subtitle">Instances:</span><Link to="GroupLayout" params={{groupID: this.props.group.id, appID: this.props.group.application_id}}><span className="activeLink"> {instances_total}<span className="fa fa-caret-right" /></span></Link>
              <div className="divider">|</div>
              <span className="subtitle">Channel:</span> <span className={styleGroupChannel}>{groupChannel}</span>
            </Col>
          </Row>
          <Row className="groups--resume noExtended">
            <Col xs={8}>
              <span className="subtitle">Rollout policy:</span> Max {this.props.group.policy_max_updates_per_period} updates per {this.props.group.policy_period_interval}
            </Col>
            <Col xs={4} className="alignRight">
              <span className="subtitle displayInline">Updates enabled:</span>
              <div className="displayInline">
                <Switch checked={this.props.group.policy_updates_enabled} disabled={true} checkedChildren={"✔"} unCheckedChildren={"✘"} />
              </div>
            </Col>
          </Row>
          <Row className="groups--resume">
            <VersionBreakdown version_breakdown={version_breakdown} channel={channel} />
          </Row>
        </div>
      </div>
    )
  }

}

export default Item
