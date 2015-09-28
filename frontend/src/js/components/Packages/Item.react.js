import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, OverlayTrigger, Button, Popover } from "react-bootstrap"
import ConfirmationContent from "../Common/ConfirmationContent.react"
import moment from "moment"
import _ from "underscore"
import VersionBullet from "../Common/VersionBullet.react"
import ModalButton from "../Common/ModalButton.react"

class Item extends React.Component {

  constructor(props) {
    super(props)
    this.deletePackage = this.deletePackage.bind(this)
  }

  static propTypes: {
    packageItem: React.PropTypes.object.isRequired,
    channels: React.PropTypes.array
  }

  deletePackage() {
    let confirmationText = "Are you sure you want to delete this package?"
    if (confirm(confirmationText)) {
      applicationsStore.deletePackage(this.props.packageItem.application_id, this.props.packageItem.id)
    }
  }

  render() {
    let filename = this.props.packageItem.filename ? this.props.packageItem.filename : "",
        url = this.props.packageItem.url ? this.props.packageItem.url : "#",
        date = moment(this.props.packageItem.created_ts).format("hh:mma, DD/MM"),
        type = this.props.packageItem.type ? this.props.packageItem.type : 1,
        processedChannels = _.where(this.props.channels, {package_id: this.props.packageItem.id}),
        popoverContent = {
          type: "package",
          appID: this.props.packageItem.application_id,
          packageID: this.props.packageItem.id
        }

    return (
      <Row>
        <Col xs={7} className="noPadding">
          <div className="package--info">
            <div className={"containerIcon container-" + type}></div>
            <br />          
            <span className="subtitle">Version:</span>
            {processedChannels.map((channel, i) =>
              <VersionBullet channel={channel} key={i} />
            )}
            {this.props.packageItem.version}
            <br />
            <span className="subtitle">Released:</span> {date}
          </div>
        </Col>
        <Col xs={5} className="alignRight marginTop7">
          <ModalButton icon="edit" modalToOpen="UpdatePackageModal" data={{channel: this.props.packageItem}} />
          <button className="cr-button displayInline fa fa-trash-o" onClick={this.deletePackage}></button>
        </Col>
      </Row>
    )
  }

}

export default Item