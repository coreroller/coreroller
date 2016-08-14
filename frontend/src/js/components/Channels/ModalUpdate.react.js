import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert, ButtonInput } from "react-bootstrap"
import { Form, ValidatedInput } from "react-bootstrap-validation"
import ColorPicker from "react-color"
import moment from "moment"
import _ from "underscore"

class ModalUpdate extends React.Component {

  constructor(props) {
    super(props)
    this.handleFocus = this.handleFocus.bind(this)
    this.changeColor = this.changeColor.bind(this)
    this.handleColorPicker = this.handleColorPicker.bind(this)
    this.handleColorPickerClose = this.handleColorPickerClose.bind(this)
    this.updateChannel = this.updateChannel.bind(this)
    this.checkBlacklistChannels = this.checkBlacklistChannels.bind(this)
    this.handleValidSubmit = this.handleValidSubmit.bind(this)
    this.handleInvalidSubmit = this.handleInvalidSubmit.bind(this)
    this.exitedModal = this.exitedModal.bind(this)

    this.state = {
      channelColor: props.data.channel.color,
      displayColorPicker: false,
      isLoading: false,
      alertVisible: false
    }
  }

  static propTypes : {
    data: PropTypes.object.isRequired,
    modalVisible: PropTypes.bool.isRequired
  }

  updateChannel() {
    this.setState({isLoading: true})
    let data = {
      id: this.props.data.channel.id,
      name: this.refs.nameChannel.getValue(),
      color: this.state.channelColor,
      application_id: this.props.data.channel.application_id
    }

    let package_id = this.refs.packageChannel.getValue()
    if (package_id) {
      data["package_id"] = package_id
    }

    applicationsStore.updateChannel(data).
      done(() => {
        this.props.onHide()
        this.setState({isLoading: false})
      }).
      fail(() => {
        this.setState({
          channelColor: this.props.data.channel.color,
          displayColorPicker: false,
          isLoading: false,
          alertVisible: false
        })
      })
  }

  handleFocus() {
    this.setState({alertVisible: false})
  }

  handleColorPickerClose() {
    this.setState({ displayColorPicker: false })
  }

  handleColorPicker() {
    this.setState({ "displayColorPicker": !this.state.displayColorPicker })
  }

  changeColor(color) {
    this.setState({ channelColor: "#" + color.hex })
  }

  componentWillReceiveProps(nextProps) {
    this.setState({
      channelColor: nextProps.data.channel.color,
    })
  }

  checkBlacklistChannels() {
    let packagesWithChannelInBlacklist = _.map(this.props.data.packages, (packageItem) => {
      const channels_blacklist = _.isNull(packageItem.channels_blacklist) ? [] : packageItem.channels_blacklist
      if (_.indexOf(channels_blacklist, this.props.data.channel.id) > -1) {
        return packageItem.id
      }
    })
    return _.compact(packagesWithChannelInBlacklist)
  }

  handleValidSubmit() {
    this.updateChannel()
  }

  handleInvalidSubmit() {
    // this.setState({alertVisible: false})
  }

  exitedModal() {
    this.setState({
      channelColor: this.props.data.channel.color,
      displayColorPicker: false,
      isLoading: false,
      alertVisible: false
    })
  }

  render() {
    let packages = this.props.data.packages ? this.props.data.packages : [],
        selectedPackage = this.props.data.channel.package_id ? this.props.data.channel.package_id : "",
        popupPosition = {
          position: "absolute",
          top: "10px",
          left: "10px"
        },
        divColor = {
          backgroundColor: this.state.channelColor
        },
        btnStyle = this.state.isLoading ? " loading" : "",
        btnContent = this.state.isLoading ? "Please wait" : "Submit",
        channels_blacklist = this.checkBlacklistChannels()

    return (
      <Modal {...this.props} show={this.props.modalVisible} animation={true} onExited={this.onExited}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Update channel</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form" onFocus={this.handleFocus}>
            <Form onValidSubmit={this.handleValidSubmit} onInvalidSubmit={this.handleInvalidSubmit}>
              <ValidatedInput
                type="text"
                label="*Name:"
                name="nameChannel"
                ref="nameChannel"
                defaultValue={this.props.data.channel.name}
                required={true}
                validationEvent="onBlur"
                validate="required,isLength:1:25"
                errorHelp={{
                  required: "Please enter a name",
                  isLength: "Name must be less than 25 characters"
                }}
              />
              <div className="form-group">
                <label className="control-label">
                  <span>Color:</span>
                </label>
                <div className="swatch" >
                  <div className="color" style={divColor} onClick={ this.handleColorPicker } />
                </div>
                <ColorPicker
                  color={ this.state.channelColor }
                  position="below"
                  display={ this.state.displayColorPicker }
                  onChange={ this.changeColor }
                  onChangeComplete={ this.handleColorPickerClose }
                  type="compact" positionCSS={popupPosition} />
              </div>
              <Input type="select" label="Package:" defaultValue={selectedPackage} placeholder="" groupClassName="arrow-icon" ref="packageChannel">
                <option value="" />
                {packages.map((packageItem, i) =>
                  <option
                    value={packageItem.id}
                    disabled={ ( _.indexOf(channels_blacklist, packageItem.id) > -1) ? true : false }
                    key={"modalUpdateChannel_packageItem_" + i}>
                      {packageItem.version} &nbsp;&nbsp;(created: {moment.utc(packageItem.created_ts).local().format("DD/MM/YYYY")})
                  </option>
                )}
              </Input>
              <div className="form--legend minlegend marginBottom15">
                <b>NOTE:</b> updates only happen when a <b>higher</b> version is available. This means that if your instances are running version {"1.3.0"} and the channel is updated pointing it to a lower version (lets say {"1.2.0"}), they won’t execute a downgrade. Only after the channel is pointing to a version higher than {"1.3.0"} they will receive an update.
              </div>
              <div className="modal--footer">
                <Row>
                  <Col xs={8}>
                    <Alert bsStyle="danger" className={this.state.alertVisible ? "alert--visible" : ""}>
                      <strong>Error!</strong> The request failed, please check the form
                    </Alert>
                  </Col>
                  <Col xs={4}>
                    <ButtonInput
                      type="submit"
                      bsStyle="default"
                      className={"plainBtn" + btnStyle}
                      disabled={this.state.isLoading}
                      value={btnContent}
                    />
                  </Col>
                </Row>
              </div>
            </Form>
          </div>
        </Modal.Body>
      </Modal>
    )
  }

}

export default ModalUpdate
