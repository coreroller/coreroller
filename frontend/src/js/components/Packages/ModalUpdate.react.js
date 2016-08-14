import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert, ButtonInput } from "react-bootstrap"
import { Form, ValidatedInput } from "react-bootstrap-validation"
import Select from "react-select"
import _ from "underscore"
import {REGEX_SEMVER, REGEX_URL, REGEX_SIZE} from "../../constants/regex"
import $ from "jquery"

class ModalUpdate extends React.Component {

  constructor(props) {
    super(props)
    this.handleFocus = this.handleFocus.bind(this)
    this.updatePackage = this.updatePackage.bind(this)
    this.hanldeBlacklistChange = this.hanldeBlacklistChange.bind(this)
    this.handleChangeTypePackage = this.handleChangeTypePackage.bind(this)
    this.handleChangeCoreOSSha256 = this.handleChangeCoreOSSha256.bind(this)
    this.handleValidSubmit = this.handleValidSubmit.bind(this)
    this.handleInvalidSubmit = this.handleInvalidSubmit.bind(this)
    this.exitedModal = this.exitedModal.bind(this)

    this.state = {
      isLoading: false,
      alertVisible: false,
      channels_blacklist: props.data.channel.channels_blacklist ? props.data.channel.channels_blacklist.join(",") : "",
      typeCoreOS: props.data.channel.type == 1 ? true : false,
      disabledCoreOSSha256: props.data.channel.type == 1 ? false : true,
      coreOSSha256Package: props.data.channel.coreos_action ? props.data.channel.coreos_action.sha256 : ""
    }
  }

  static propTypes : {
    data: PropTypes.object.isRequired,
    modalVisible: PropTypes.bool.isRequired
  }

  updatePackage() {
    this.setState({isLoading: true})
    let data = {
      id: this.props.data.channel.id,
      filename: this.refs.filenamePackage.getValue(),
      description: this.refs.descriptionPackage.getValue(),
      url: this.refs.urlPackage.getValue(),
      version: this.refs.versionPackage.getValue(),
      type: parseInt(this.refs.typePackage.getValue()),
      size: (this.refs.sizePackage.getValue()).toString(),
      hash: this.refs.hashPackage.getValue(),
      application_id: this.props.data.channel.application_id,
      channels_blacklist: _.isEmpty(this.state.channels_blacklist) ? null : this.state.channels_blacklist.split(",")
    }

    if (this.state.typeCoreOS) {
      data.coreos_action = {sha256: this.state.coreOSSha256Package}
    }

    applicationsStore.updatePackage(data).
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
    this.setState({channels_blacklist: value})
  }

  handleChangeTypePackage() {
    const selectedTypePackage = this.refs.typePackage.getValue()
    if (selectedTypePackage == 1) {
      this.setState({disabledCoreOSSha256: false, typeCoreOS: true})
    } else {
      this.setState({disabledCoreOSSha256: true, typeCoreOS: false, coreOSSha256Package: ""})
      // Validated not required fields when it´s not CoreOs type
      const sizePackage = $(React.findDOMNode(this.refs.sizePackage)).find("input")[0],
            hashPackage = $(React.findDOMNode(this.refs.hashPackage)).find("input")[0],
            filenamePackage = $(React.findDOMNode(this.refs.filenamePackage)).find("input")[0]

      sizePackage.focus()
      setTimeout(() => {
        sizePackage.blur()
        hashPackage.focus()
      }, 5)
      setTimeout(() => {
        hashPackage.blur()
        filenamePackage.focus()
      }, 10)
      setTimeout(() => {
        filenamePackage.blur()
      }, 15)
    }
  }

  handleChangeCoreOSSha256(e) {
    this.setState({coreOSSha256Package: e.target.value})
  }

  handleValidSubmit() {
    this.updatePackage()
  }

  handleInvalidSubmit() {
    // this.setState({alertVisible: true})
  }

  exitedModal() {
    this.setState({
      alertVisible: false,
      isLoading: false,
      channels_blacklist: this.props.data.channel.channels_blacklist ? this.props.data.channel.channels_blacklist.join(",") : "",
      typeCoreOS: this.props.data.channel.type == 1 ? true : false,
      disabledCoreOSSha256: this.props.data.channel.type == 1 ? false : true,
      coreOSSha256Package: this.props.data.channel.coreos_action ? this.props.data.channel.coreos_action.sha256 : ""
    })
  }

  render() {
    let packages = this.props.data.channels ? this.props.data.channels : [],
        btnStyle = this.state.isLoading ? " loading" : "",
        btnContent = this.state.isLoading ? "Please wait" : "Submit",
        typeCoreOS = this.state.typeCoreOS

    let blacklistOptions = _.map(packages, (packageItem) => {
      let option = { value: packageItem.id, label: packageItem.name }
      if (packageItem.package) {
        if (this.props.data.channel.version === packageItem.package.version) {
          option.disabled = true
          option.label += " (channel pointing to this package)"
        }
      }
      return option
    })

    return (
      <Modal {...this.props} show={this.props.modalVisible} animation={true} onExited={this.exitedModal}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Update package</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form modal--form-with-captions" onFocus={this.handleFocus}>
            <Form onValidSubmit={this.handleValidSubmit} onInvalidSubmit={this.handleInvalidSubmit}>
              <Input type="select" label="*Type:" defaultValue={this.props.data.channel.type} placeholder="" groupClassName="arrow-icon" ref="typePackage" required={true} onChange={this.handleChangeTypePackage}>
                <option value={1}>Coreos</option>
                <option value={4}>Other</option>
              </Input>
              <ValidatedInput
                type="text"
                label="*Url:"
                name="urlPackage"
                ref="urlPackage"
                defaultValue={this.props.data.channel.url}
                required={true}
                validationEvent="onBlur"
                validate={(val) => {
                  return REGEX_URL.test(val)
                }}
                errorHelp="Please enter a valid url and no more than 256 characters"
              />
              <ValidatedInput
                type="text"
                label={(typeCoreOS ? "*" : "") + "Filename:"}
                name="filenamePackage"
                ref="filenamePackage"
                defaultValue={this.props.data.channel.filename}
                required={typeCoreOS}
                validationEvent="onBlur"
                validate={(val) => {
                  if (typeCoreOS) {
                    return val.length > 1 && val.length <= 100
                  } else {
                    return val.length <= 100
                  }
                }}
                errorHelp="Please enter a valid filename (less than 100 characters)"
              />
              <Input type="textarea" label="Description:" defaultValue={this.props.data.channel.description} ref="descriptionPackage" maxLength={250} className="smallHeight" />
              <Row>
                <Col xs={6}>
                  <ValidatedInput
                    type="text"
                    label="*Version:"
                    name="versionPackage"
                    ref="versionPackage"
                    defaultValue={this.props.data.channel.version}
                    required={true}
                    validationEvent="onBlur"
                    validate={(val) => {
                      return REGEX_SEMVER.test(val)
                    }}
                    errorHelp="Please enter a valid semver (1.0.1)"
                  />
                  <div className="form--legend minlegend minlegend--formGroup">{"Use SemVer format (1.0.1)"}</div>
                </Col>
                <Col xs={6}>
                  <ValidatedInput
                    type="text"
                    label={(typeCoreOS ? "*" : "") + "Size (bytes):"}
                    name="sizePackage"
                    ref="sizePackage"
                    defaultValue={this.props.data.channel.size ? parseInt(this.props.data.channel.size) : ""}
                    required={typeCoreOS}
                    validationEvent="onBlur"
                    validate={(val) => {
                      if (typeCoreOS) {
                        return REGEX_SIZE.test(val) && val.length > 0
                      } else {
                        return (REGEX_SIZE.test(val) || _.isEmpty(val))  && val.length >= 0
                      }
                    }}
                    errorHelp="Please enter a valid size (less than 20 digits)"
                  />
                </Col>
              </Row>
              <ValidatedInput
                type="text"
                label={(typeCoreOS ? "*" : "") + "Hash:"}
                name="hashPackage"
                ref="hashPackage"
                defaultValue={this.props.data.channel.hash}
                required={typeCoreOS}
                validationEvent="onBlur"
                validate={(val) => {
                  if (typeCoreOS) {
                    return val.length > 1 && val.length <= 64
                  } else {
                    return val.length <= 64
                  }
                }}
                errorHelp="Please enter a valid hash (less than 64 characters)"
              />
              {typeCoreOS &&
                <div>
                  <div className="form--legend minlegend minlegend--formGroup">{"Tip: cat update.gz | openssl dgst -sha1 -binary | base64"}</div>
                  <ValidatedInput
                    type="text"
                    label="*CoreOS action sha256:"
                    name="coreOSSha256Package"
                    ref="coreOSSha256Package"
                    value={this.state.coreOSSha256Package}
                    required={true}
                    className={this.state.disabledCoreOSSha256 ? "disabled" : ""}
                    disabled={this.state.disabledCoreOSSha256}
                    onChange={this.handleChangeCoreOSSha256}
                    validationEvent="onBlur"
                    validate="required"
                    errorHelp={{
                      required: "Please enter a valid value"
                    }}
                  />
                  <div className="form--legend minlegend minlegend--formGroup">{"Tip: cat update.gz | openssl dgst -sha256 -binary | base64"}</div>
                </div>
              }
              <Row>
                <Col xs={12}>
                  <div className="form-group">
                    <label className="control-label">Channels Blacklist</label>
                    <Select
                      name="channels_blacklist"
                      value={ this.state.channels_blacklist }
                      multi
                      placeholder=" "
                      options={ blacklistOptions }
                      onChange={ this.hanldeBlacklistChange }
                    />
                    <div className="form--legend minlegend minlegend--formGroup">Blacklisted channels cannot point to this package</div>
                  </div>
                </Col>
              </Row>
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
