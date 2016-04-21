import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert } from "react-bootstrap"
import ColorPicker from "react-color"
import moment from "moment"

class ModalAdd extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      channelColor: "#777777",
      displayColorPicker: false,
      isLoading: false,
      alertVisible: false
    }
    this.handleFocus = this.handleFocus.bind(this)
    this.createChannel = this.createChannel.bind(this)
    this.changeColor = this.changeColor.bind(this)
    this.handleColorPicker = this.handleColorPicker.bind(this)
    this.handleColorPickerClose = this.handleColorPickerClose.bind(this)
  }

  static propTypes : {
    data: PropTypes.object
  }

  createChannel() {
    this.setState({isLoading: true})
    let data = {
      name: this.refs.nameNewChannel.getValue(),
      color: this.state.channelColor,
      application_id: this.props.data.applicationID
    }

    let package_id = this.refs.packageChannel.getValue()
    if (package_id) {
      data["package_id"] = package_id
    }

    applicationsStore.createChannel(data).
      done(() => {
        this.props.onHide()
        this.setState({isLoading: false})
      }).
      fail(() => {
        this.setState({alertVisible: true, isLoading: false})
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

  render() {
    let packages = this.props.data.packages ? this.props.data.packages : [],
        popupPosition = {
          position: "absolute",
          top: "10px",
          left: "10px"
        },
        divColor = {
          backgroundColor: this.state.channelColor
        },
        btnStyle = this.state.isLoading ? " loading" : "",
        btnContent = this.state.isLoading ? "Please wait" : "Submit"

    return (
      <Modal {...this.props} animation={true}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Add new channel</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form">
            <form role="form" action="" onFocus={this.handleFocus}>
              <Input type="text" label="*Name:" ref="nameNewChannel" required={true} maxLength={25} />
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
              <Input type="select" label="Package:" placeholder="" groupClassName="arrow-icon" ref="packageChannel">
                <option value="" />
                {packages.map((packageItem, i) =>
                  <option value={packageItem.id} key={i}>
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
                      <strong>Error!</strong> Please check the form
                    </Alert>
                  </Col>
                  <Col xs={4}>
                    <Button bsStyle="default" className={"plainBtn" + btnStyle} disabled={this.state.isLoading} onClick={this.createChannel}>{btnContent}</Button>
                  </Col>
                </Row>
              </div>
            </form>
          </div>
        </Modal.Body>
      </Modal>
    )
  }

}

export default ModalAdd
