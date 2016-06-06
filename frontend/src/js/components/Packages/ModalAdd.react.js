import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert } from "react-bootstrap"
import Select from "react-select"
import _ from "underscore"

class ModalAdd extends React.Component {

  constructor(props) {
    super(props)
    this.handleFocus = this.handleFocus.bind(this)
    this.createPackage = this.createPackage.bind(this)
    this.hanldeBlacklistChange = this.hanldeBlacklistChange.bind(this)
    this.handleChangeTypeNewPackage = this.handleChangeTypeNewPackage.bind(this)
    this.handleChangeCoreOSSha256 = this.handleChangeCoreOSSha256.bind(this)
    this.state = {
      isLoading: false,
      alertVisible: false,
      blacklist_channels: "",
      typeCoreOS: false,
      disabledCoreOSSha256: true,
      coreOSSha256NewPackage: ""
    }
  }

  static propTypes : {
    data: PropTypes.object
  }

  createPackage() {
    this.setState({isLoading: true})
    let data = {
      filename: this.refs.filenameNewPackage.getValue(),
      description: this.refs.descriptionNewPackage.getValue(),
      url: this.refs.urlNewPackage.getValue(),
      version: this.refs.versionNewPackage.getValue(),
      type: parseInt(this.refs.typeNewPackage.getValue()),
      size: parseInt(this.refs.sizeNewPackage.getValue()),
      hash: this.refs.hashNewPackage.getValue(),
      application_id: this.props.data.appID,
      channels_blacklist: _.isEmpty(this.state.channels_blacklist) ? null : this.state.channels_blacklist.split(",")
    }

    if (this.state.typeCoreOS) {
      data.coreos_action = {sha256: this.state.coreOSSha256NewPackage}
    }

    applicationsStore.createPackage(data).
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

  hanldeBlacklistChange(value) {
    this.setState({blacklist_channels: value})
  }

  handleChangeTypeNewPackage() {
    const selectedTypeNewPackage = this.refs.typeNewPackage.getValue()
    if (selectedTypeNewPackage == 1) {
      this.setState({disabledCoreOSSha256: false, typeCoreOS: true})
    } else {
      this.setState({disabledCoreOSSha256: true, typeCoreOS: false, coreOSSha256NewPackage: ""})
    }
  }

  handleChangeCoreOSSha256(e) {
    this.setState({coreOSSha256NewPackage: e.target.value})
  }

  render() {
    let packages = this.props.data.channels ? this.props.data.channels : [],
        btnStyle = this.state.isLoading ? " loading" : "",
        btnContent = this.state.isLoading ? "Please wait" : "Submit",
        typeCoreOS = this.state.typeCoreOS

    let blacklistOptions = _.map(packages, (packageItem) => {
      return { value: packageItem.id, label: packageItem.name }
    })

    return (
      <Modal {...this.props} animation={true}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Add new package</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form">
            <form role="form" action="" onFocus={this.handleFocus}>
              <Input type="select" label="*Type:" defaultValue={4} placeholder="" groupClassName="arrow-icon" ref="typeNewPackage" required={true} onChange={this.handleChangeTypeNewPackage}>
                <option value={1}>Coreos</option>
                <option value={2}>Docker image</option>
                <option value={3}>Rocket image</option>
                <option value={4}>Other</option>
              </Input>
              <Input type="url" label="*Url:" ref="urlNewPackage" required={true} maxLength={256} />
              <Input type="text" label={(typeCoreOS ? "*" : "") + "Filename:"} ref="filenameNewPackage" maxLength={100} required={typeCoreOS} />
              <Input type="textarea" label="Description:" ref="descriptionNewPackage" maxLength={250} className="smallHeight" />
              <Row>
                <Col xs={6}>
                  <Input type="text" label="*Version:" ref="versionNewPackage" required={true} />
                  <div className="form--legend minlegend minlegend--formGroup">Use SemVer format (1.0.1)</div>
                </Col>
                <Col xs={6}>
                  <Input type="number" label={(typeCoreOS ? "*" : "") + "Size (bytes):"} ref="sizeNewPackage" maxLength={20} required={typeCoreOS} />
                </Col>
              </Row>
              <Input type="text" label={(typeCoreOS ? "*" : "") + "Hash:"} ref="hashNewPackage" maxLength={64} required={typeCoreOS} />
              <div className="form--legend minlegend minlegend--formGroup">Tip: cat update.gz | openssl dgst -sha1 -binary | base64</div>
              <Input type="text" label={(typeCoreOS ? "*" : "") + "CoreOS action sha256:"} value={this.state.coreOSSha256NewPackage} ref="coreOSSha256NewPackage" className={this.state.disabledCoreOSSha256 ? "disabled" : ""} disabled={this.state.disabledCoreOSSha256} onChange={this.handleChangeCoreOSSha256} required={typeCoreOS} />
              <div className="form--legend minlegend minlegend--formGroup">Tip: cat update.gz | openssl dgst -sha256 -binary | base64</div>
              <Row>
                <Col xs={12}>
                  <div className="form-group">
                    <label className="control-label">Channels Backlist</label>
                    <Select
                      name="channels_blacklist"
                      value={ this.state.blacklist_channels }
                      multi
                      placeholder=" "
                      options={ blacklistOptions }
                      onChange={ this.hanldeBlacklistChange }
                    />
                  </div>
                </Col>
              </Row>
              <div className="modal--footer">
                <Row>
                  <Col xs={8}>
                    <Alert bsStyle="danger" className={this.state.alertVisible ? "alert--visible" : ""}>
                      <strong>Error!</strong> Please check the form
                    </Alert>
                  </Col>
                  <Col xs={4}>
                    <Button bsStyle="default" className={"plainBtn" + btnStyle} disabled={this.state.isLoading} onClick={this.createPackage}>{btnContent}</Button>
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
