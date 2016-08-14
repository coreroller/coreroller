import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert, OverlayTrigger, ButtonInput } from "react-bootstrap"
import { Form, ValidatedInput } from "react-bootstrap-validation"
import Switch from "rc-switch"
import _ from "underscore"
import moment from "moment-timezone"
import {tooltipSafeMode, tooltipOfficeHours} from "../Common/Tooltips.react"

class ModalAdd extends React.Component {

  constructor(props) {
    super(props);
    this.handleFocus = this.handleFocus.bind(this)
    this.createGroup = this.createGroup.bind(this)
    this.validateTimezone = this.validateTimezone.bind(this)
    this.changeSafeMode = this.changeSafeMode.bind(this)
    this.changePolicyUpdates = this.changePolicyUpdates.bind(this)
    this.changePolicyOfficeHours = this.changePolicyOfficeHours.bind(this)
    this.handleValidSubmit = this.handleValidSubmit.bind(this)
    this.handleInvalidSubmit = this.handleInvalidSubmit.bind(this)
    this.exitedModal = this.exitedModal.bind(this)

    this.state = {
      safeMode: true,
      policyUpdates: true,
      policyOfficeHours: false,
      isLoading: false,
      alertVisible: false,
      timezoneError: false
    }
  }

  static propTypes : {
    data: PropTypes.object
  }

  createGroup() {
    this.setState({isLoading: true})

    let isValidTimezone = this.validateTimezone()

    if (isValidTimezone) {
      let period_interval = this.refs.timingUpdatesPerPeriod.getValue() + " " + this.refs.timingUpdatesPerPeriodUnit.getValue(),
          update_timeout = this.refs.timingUpdatesTimeout.getValue() + " " + this.refs.timingUpdatesTimeoutUnit.getValue()

      var data = {
        name: this.refs.nameNewGroup.getValue(),
        description: this.refs.descriptionNewGroup.getValue(),
        policy_safe_mode: this.state.safeMode,
        policy_max_updates_per_period: parseInt(this.refs.maxUpdatesPerPeriodInterval.getValue()),
        policy_updates_enabled: this.state.policyUpdates,
        policy_period_interval: period_interval,
        policy_update_timeout: update_timeout,
        application_id: this.props.data.appID,
        policy_office_hours: this.state.policyOfficeHours
      }

      let channel_id = this.refs.channelGroup.getValue()
      if (channel_id) {
        data["channel_id"] = channel_id
      }

      let timezone = this.refs.policyTimezone.getValue()
      if (timezone) {
        data["policy_timezone"] = timezone
      }

      applicationsStore.createGroup(data).
        done(() => {
          this.props.onHide()
          this.setState({isLoading: false})
        }).
        fail(() => {
          this.setState({alertVisible: true, isLoading: false})
        })
    } else {
      this.setState({isLoading: false, timezoneError: true})
    }
  }

  validateTimezone() {
    let timezone = this.refs.policyTimezone.getValue()

    if (this.state.policyOfficeHours && _.isEmpty(timezone)) {
      return false
    } else {
      return true
    }
  }

  handleFocus() {
    this.setState({alertVisible: false, timezoneError: false})
  }

  changeSafeMode() {
    this.setState({
      safeMode: !this.state.safeMode
    })
  }

  changePolicyUpdates() {
    this.setState({
      policyUpdates: !this.state.policyUpdates
    })
  }

  changePolicyOfficeHours() {
    this.setState({
      policyOfficeHours: !this.state.policyOfficeHours
    })
  }

  handleValidSubmit() {
    this.createGroup()
  }

  handleInvalidSubmit() {
    // this.setState({alertVisible: true})
  }

  exitedModal() {
    this.setState({
      safeMode: true,
      policyUpdates: true,
      policyOfficeHours: false,
      isLoading: false,
      alertVisible: false,
      timezoneError: false
    })
  }

  render() {
    let channels = this.props.data.channels ? this.props.data.channels : [],
        btnStyle = this.state.isLoading ? " loading" : "",
        btnContent = this.state.isLoading ? "Please wait" : "Submit",
        timezones = moment.tz.names(),
        timezoneError = this.state.timezoneError ? {bsStyle: "error"} : ""

    return (
      <Modal {...this.props} animation={true} onExited={this.exitedModal}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Add new group</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form" onFocus={this.handleFocus}>
            <Form onValidSubmit={this.handleValidSubmit} onInvalidSubmit={this.handleInvalidSubmit}>
              <ValidatedInput
                type="text"
                label="*Name:"
                name="nameNewGroup"
                ref="nameNewGroup"
                required={true}
                validationEvent="onBlur"
                validate="required,isLength:1:50"
                errorHelp={{
                  required: "Please enter a name",
                  isLength: "Name must be less than 50 characters"
                }}
              />
              <Input type="text" label="Description:" ref="descriptionNewGroup" maxLength={250} />
              <Input type="select" label="Channel:" placeholder="" defaultValue="" groupClassName="arrow-icon" ref="channelGroup">
                <option value="" />
                {channels.map((channel, i) =>
                  <option value={channel.id} key={"modalAddGroup_channel_" + i}>{channel.name}</option>
                )}
              </Input>
              <h5><span>Rollout policy</span></h5>
              <Row>
                <Col xs={12} className="form--limit marginBottom15">
                  <div className="form-group noMargin">
                    <label className="normalText" htmlFor="policyUpdatesNewGroup">Updates enabled:</label>
                    <div className="displayInline">
                      <Switch defaultChecked onChange={this.changePolicyUpdates} checkedChildren={"✔"} unCheckedChildren={"✘"} />
                    </div>
                  </div>
                  <div className="form-group noMargin">
                    <OverlayTrigger trigger={["hover", "focus"]} container={this} placement="bottom" overlay={tooltipOfficeHours}>
                      <label className="normalText" htmlFor="policyOfficeHours"><i className="fa fa-question-circle"></i> Only office hours:</label>
                    </OverlayTrigger>
                    <div className="displayInline">
                      <Switch onChange={this.changePolicyOfficeHours} checkedChildren={"✔"} unCheckedChildren={"✘"} />
                    </div>
                  </div>
                  <div className="form-group noMargin">
                    <OverlayTrigger trigger={["hover", "focus"]} container={this} placement="bottom" overlay={tooltipSafeMode}>
                      <label className="normalText" htmlFor="safeModeNewGroup"><i className="fa fa-question-circle"></i> Safe mode:</label>
                    </OverlayTrigger>
                    <div className="displayInline lastCheck">
                      <Switch defaultChecked onChange={this.changeSafeMode} checkedChildren={"✔"} unCheckedChildren={"✘"} />
                    </div>
                  </div>
                </Col>
              </Row>
              <Row>
                <Col xs={12}>
                  <Input type="select" label="Timezone:" {...timezoneError} placeholder="" groupClassName="arrow-icon" ref="policyTimezone">
                    <option value="" />
                    {timezones.map((timezone, i) =>
                      <option value={timezone} key={"modalAddGroup_timezone_" + i}>{timezone}</option>
                    )}
                  </Input>
                </Col>
              </Row>
              <div className="form--limit form--limit-inputs">
                Max
                <ValidatedInput
                  type="number"
                  label=""
                  standalone
                  name="maxUpdatesPerPeriodInterval"
                  ref="maxUpdatesPerPeriodInterval"
                  defaultValue={2}
                  min={1}
                  required={true}
                  validationEvent="onBlur"
                  validate="required,isInt"
                  errorHelp={{
                    required: "",
                    isInt: ""
                  }}
                />
                updates per period interval
                <ValidatedInput
                  type="number"
                  label=""
                  standalone
                  name="timingUpdatesPerPeriod"
                  ref="timingUpdatesPerPeriod"
                  defaultValue={15}
                  min={1}
                  required={true}
                  validationEvent="onBlur"
                  validate="required,isInt"
                  errorHelp={{
                    required: "",
                    isint: ""
                  }}
                />
                <Input type="select" placeholder="" ref="timingUpdatesPerPeriodUnit" groupClassName="form-group group-inline arrow-icon" defaultValue="minutes" required={true}>
                  <option value="minutes">minutes</option>
                  <option value="hours">hours</option>
                  <option value="days">days</option>
                </Input>
              </div>
              <div className="form--legend minlegend marginBottom15">Never update more than MAX instances per interval defined.</div>
              <div className="form-group">
                <div className="form--limit form--limit-inputs">
                  Updates timeout
                  <ValidatedInput
                    type="number"
                    label=""
                    standalone
                    name="timingUpdatesTimeout"
                    ref="timingUpdatesTimeout"
                    defaultValue={60}
                    min={1}
                    required={true}
                    validationEvent="onBlur"
                    validate="required,isInt"
                    errorHelp={{
                      required: "",
                      isInt: ""
                    }}
                  />
                  <div className="form-group group-inline">
                    <Input type="select" placeholder="" ref="timingUpdatesTimeoutUnit" groupClassName="form-group group-inline arrow-icon" defaultValue="minutes">
                      <option value="minutes">minutes</option>
                      <option value="hours">hours</option>
                      <option value="days">days</option>
                    </Input>
                  </div>
                </div>
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

export default ModalAdd
