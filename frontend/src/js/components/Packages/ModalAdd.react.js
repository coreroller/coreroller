import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert, ButtonInput } from "react-bootstrap"
import { Form, ValidatedInput } from "react-bootstrap-validation"
import Select from "react-select"
import _ from "underscore"
import {REGEX_SEMVER, REGEX_URL, REGEX_SIZE} from "../../constants/regex"
import $ from "jquery"

class ModalAdd extends React.Component {

  constructor(props) {
    super(props)
    this.handleFocus = this.handleFocus.bind(this)
    this.createPackage = this.createPackage.bind(this)
    this.hanldeBlacklistChange = this.hanldeBlacklistChange.bind(this)
    this.handleChangeTypeNewPackage = this.handleChangeTypeNewPackage.bind(this)
    this.handleChangeCoreOSSha256 = this.handleChangeCoreOSSha256.bind(this)
    this.handleValidSubmit = this.handleValidSubmit.bind(this)
    this.handleInvalidSubmit = this.handleInvalidSubmit.bind(this)
    this.exitedModal = this.exitedModal.bind(this)

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
      size: (this.refs.sizeNewPackage.getValue()).toString(),
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
      // Validated not required fields when it´s not CoreOs type
      const sizeNewPackage = $(React.findDOMNode(this.refs.sizeNewPackage)).find("input")[0],
            hashNewPackage = $(React.findDOMNode(this.refs.hashNewPackage)).find("input")[0],
            filenameNewPackage = $(React.findDOMNode(this.refs.filenameNewPackage)).find("input")[0]

      sizeNewPackage.focus()
      setTimeout(() => {
        sizeNewPackage.blur()
        hashNewPackage.focus()
      }, 5)
      setTimeout(() => {
        hashNewPackage.blur()
        filenameNewPackage.focus()
      }, 10)
      setTimeout(() => {
        filenameNewPackage.blur()
      }, 15)
    }
  }

  handleChangeCoreOSSha256(e) {
    this.setState({coreOSSha256NewPackage: e.target.value})
  }

  handleValidSubmit() {
    this.createPackage()
  }

  handleInvalidSubmit() {
    // this.setState({alertVisible: true})
  }

  exitedModal() {
    this.setState({
      alertVisible: false,
      isLoading: false,
      blacklist_channels: "",
      typeCoreOS: false,
      disabledCoreOSSha256: true,
      coreOSSha256NewPackage: ""
    })
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
      <Modal {...this.props} animation={true} onExited={this.exitedModal}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Add new package</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form modal--form-with-captions" onFocus={this.handleFocus}>
            <Form onValidSubmit={this.handleValidSubmit} onInvalidSubmit={this.handleInvalidSubmit}>
              <Input type="select" label="*Type:" defaultValue={4} placeholder="" groupClassName="arrow-icon" ref="typeNewPackage" required={true} onChange={this.handleChangeTypeNewPackage}>
                <option value={1}>Coreos</option>
                <option value={4}>Other</option>
              </Input>
              <ValidatedInput
                type="text"
                label="*Url:"
                name="urlNewPackage"
                ref="urlNewPackage"
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
                name="filenameNewPackage"
                ref="filenameNewPackage"
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
              <Input type="textarea" label="Description:" ref="descriptionNewPackage" maxLength={250} className="smallHeight" />
              <Row>
                <Col xs={6}>
                  <ValidatedInput
                    type="text"
                    label="*Version:"
                    name="versionNewPackage"
                    ref="versionNewPackage"
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
                    name="sizeNewPackage"
                    ref="sizeNewPackage"
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
                name="hashNewPackage"
                ref="hashNewPackage"
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
                    name="coreOSSha256NewPackage"
                    ref="coreOSSha256NewPackage"
                    value={this.state.coreOSSha256NewPackage}
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

export default ModalAdd
