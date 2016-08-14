import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, OverlayTrigger, Button, Popover, Label } from "react-bootstrap"
import ConfirmationContent from "../Common/ConfirmationContent.react"
import moment from "moment"
import _ from "underscore"
import VersionBullet from "../Common/VersionBullet.react"
import ModalButton from "../Common/ModalButton.react"
import { cleanSemverVersion } from "../../constants/helpers"

class Item extends React.Component {

  constructor(props) {
    super(props)
    this.deletePackage = this.deletePackage.bind(this)
    this.updatePackage = this.updatePackage.bind(this)
  }

  static propTypes: {
    packageItem: React.PropTypes.object.isRequired,
    channels: React.PropTypes.array,
    handleUpdatePackage: React.PropTypes.func.isRequired
  }

  deletePackage() {
    let confirmationText = "Are you sure you want to delete this package?"
    if (confirm(confirmationText)) {
      applicationsStore.deletePackage(this.props.packageItem.application_id, this.props.packageItem.id)
    }
  }

  updatePackage() {
    this.props.handleUpdatePackage(this.props.packageItem.id)
  }

  render() {
    let filename = this.props.packageItem.filename ? this.props.packageItem.filename : "",
        url = this.props.packageItem.url ? this.props.packageItem.url : "#",
        date = moment.utc(this.props.packageItem.created_ts).local().format("hh:mma, DD/MM"),
        type = this.props.packageItem.type ? this.props.packageItem.type : 1,
        processedChannels = _.where(this.props.channels, {package_id: this.props.packageItem.id}),
        popoverContent = {
          type: "package",
          appID: this.props.packageItem.application_id,
          packageID: this.props.packageItem.id
        },
        blacklistInfo = null

    if (this.props.packageItem.channels_blacklist) {
      let channelsList = _.map(this.props.packageItem.channels_blacklist, (channel, index) => {
        return (_.findWhere(this.props.channels, {id: channel})).name
      })
      blacklistInfo = channelsList.join(" - ")
    }

    return (
      <Row>
        <Col xs={7} className="noPadding">
          <div className="package--info">
            <div className={"containerIcon container-" + type}></div>
            <br />
            <span className="subtitle">Version:</span>
            {processedChannels.map((channel, i) =>
              <VersionBullet channel={channel} key={"packageItemBullet_" + i} />
            )}
            {cleanSemverVersion(this.props.packageItem.version)}
            <br />
            <span className="subtitle">Released:</span> {date}
            { !_.isNull(this.props.packageItem.channels_blacklist) &&
              <div className="label-packageItem-container">
                <Label bsStyle="danger" className="label-packageItem"><i className="fa fa-ban"></i> { blacklistInfo }</Label>
              </div>
            }
          </div>
        </Col>
        <Col xs={5} className="alignRight marginTop7">
          <button className="cr-button displayInline fa fa-edit" onClick={this.updatePackage}></button>
          <button className="cr-button displayInline fa fa-trash-o" onClick={this.deletePackage}></button>
        </Col>
      </Row>
    )
  }

}

export default Item
