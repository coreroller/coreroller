import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, OverlayTrigger, Button, Popover, Tooltip } from "react-bootstrap"
import ConfirmationContent from "../Common/ConfirmationContent.react"
import ChannelLabel from "../Common/ChannelLabel.react"

class Item extends React.Component {

  constructor(props) {
    super(props)
    this.deleteChannel = this.deleteChannel.bind(this)
    this.updateChannel = this.updateChannel.bind(this)
  }

  static propTypes: {
    channel: React.PropTypes.object.isRequired,
    packages: React.PropTypes.array.isRequired,
    handleUpdateChannel: React.PropTypes.func.isRequired
  }

  deleteChannel() {
    let confirmationText = "Are you sure you want to delete this channel?"
    if (confirm(confirmationText)) {
      applicationsStore.deleteChannel(this.props.channel.application_id, this.props.channel.id)
    }
  }

  updateChannel() {
    this.props.handleUpdateChannel(this.props.channel.id)
  }

  render() {
    let popoverContent = {
      type: "channel",
      appID: this.props.channel.application_id,
      channelID: this.props.channel.id
    }

    const name = this.props.channel ? this.props.channel.name : "",
          version = (this.props.channel && this.props.channel.package) ? this.props.channel.package.version : "-"

    const tooltip =  <Tooltip id={"tooltip_" + version}><strong>{name}</strong> - {version}</Tooltip>

    return (
      <Row>
        <Col xs={8}>
          <OverlayTrigger placement="bottom" overlay={tooltip} trigger={["hover", "focus"]}>
            <div>
              <ChannelLabel channel={this.props.channel} channelLabelStyle="fixedWidth" />
            </div>
          </OverlayTrigger>
        </Col>
        <Col xs={4} className="alignRight">
          <div className="channelsList-buttons">
            <button className="cr-button displayInline fa fa-edit" onClick={this.updateChannel}></button>
            <button className="cr-button displayInline fa fa-trash-o" onClick={this.deleteChannel}></button>
          </div>
        </Col>
      </Row>
    )
  }

}

export default Item
