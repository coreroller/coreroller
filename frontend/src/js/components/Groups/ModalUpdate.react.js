import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert } from "react-bootstrap"
import Switch from "rc-switch"
import _ from "underscore"

class ModalUpdate extends React.Component {

  constructor(props) {
    super(props);
    this.handleFocus = this.handleFocus.bind(this)
    this.updateGroup = this.updateGroup.bind(this)
    this.changeSafeMode = this.changeSafeMode.bind(this)
    this.changePolicyUpdates = this.changePolicyUpdates.bind(this)
    this.state = {
      safeMode: this.props.data.group.policy_safe_mode,
      policyUpdates: this.props.data.group.policy_updates_enabled,
      isLoading: false,
      alertVisible: false
    }
  }

  static propTypes : {
    data: PropTypes.object.isRequired
  }

  updateGroup() {
    this.setState({isLoading: true})
    let period_interval = this.refs.timingUpdatesPerPeriod.getValue() + " " + this.refs.timingUpdatesPerPeriodUnit.getValue(),
        update_timeout = this.refs.timingUpdatesTimeout.getValue() + " " + this.refs.timingUpdatesTimeoutUnit.getValue()

    let data = {
      name: this.refs.nameGroup.getValue(),
      description: this.refs.descriptionGroup.getValue(),
      policy_safe_mode: this.state.safeMode,
      policy_max_updates_per_period: parseInt(this.refs.maxUpdatesPerPeriodInterval.getValue()),
      policy_updates_enabled: this.state.updatesEnabled,
      policy_period_interval: period_interval,
      policy_update_timeout: update_timeout
    }

    let channel_id = this.refs.channelGroup.getValue()
    if (channel_id) {
      data["channel_id"] = channel_id;
    }

    let groupToUpdate = _.extend(this.props.data.group, data)

    applicationsStore.updateGroup(groupToUpdate).
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

  render() {
    let current_period_interval = this.props.data.group.policy_period_interval.split(" "),
        current_update_timeout = this.props.data.group.policy_update_timeout.split(" "),
        channels = this.props.data.channels ? this.props.data.channels : [],
        btnStyle = this.state.isLoading ? " loading" : "",
        btnContent = this.state.isLoading ? "Please wait" : "Submit" 

    return (
      <Modal {...this.props} animation={true}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Update group</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form">
            <form role="form" action="" onFocus={this.handleFocus}>
              <Input type="text" label="*Name:" defaultValue={this.props.data.group.name} ref="nameGroup" required={true} maxLength={50} />
              <Input type="textarea" label="Description:" defaultValue={this.props.data.group.description} ref="descriptionGroup" maxLength={100} />
              <Input type="select" label="Channel:" placeholder="" defaultValue={this.props.data.group.channel_id} groupClassName="arrow-icon" ref="channelGroup">
                <option value="" />
                {channels.map((channel, i) =>
                  <option value={channel.id} key={i}>{channel.name}</option>
                )}
              </Input>
              <h5><span>Rollout policy</span></h5>
              <Row>
                <Col xs={6}>
                  <div className="form-group noMargin">
                    <label className="normalText" htmlFor="policyUpdatesGroup">Updates enabled:</label>
                    <div className="displayInline">
                      <Switch checked={this.state.policyUpdates} onChange={this.changePolicyUpdates} checkedChildren={"✔"} unCheckedChildren={"✘"} />                      
                    </div>
                  </div>
                </Col>
                <Col xs={6}>
                  <div className="form-group noMargin">
                    <label className="normalText" htmlFor="safeModeGroup">Safe mode:</label>
                    <div className="displayInline">
                      <Switch checked={this.state.safeMode} onChange={this.changeSafeMode} checkedChildren={"✔"} unCheckedChildren={"✘"} />                      
                    </div>
                  </div>
                </Col>
              </Row>
              <div className="form--legend minlegend marginBottom15">
                {"If safe mode is enabled, when a new rollout starts only one instance will be granted an updated, and if it doesn’t succeed updates will be disabled in the group automatically."}
              </div>  
              <div className="form--limit">
                Max 
                <Input type="number" label="" standalone className="form-control" ref="maxUpdatesPerPeriodInterval" defaultValue={this.props.data.group.policy_max_updates_per_period} required={true} min="1" /> 
                updates per period interval 
                <Input type="number" label="" standalone className="form-control" ref="timingUpdatesPerPeriod" defaultValue={current_period_interval[0]} required={true} /> 
                <Input type="select" placeholder="" ref="timingUpdatesPerPeriodUnit" groupClassName="form-group group-inline arrow-icon" defaultValue={current_period_interval[1]} required={true}>
                  <option value="minutes">minutes</option>
                  <option value="hours">hours</option>
                  <option value="days">days</option>
                </Input>       
              </div>
              <div className="form--legend minlegend marginBottom15">Never update more than 10 instances per 15 minute time-window.</div>
              <div className="form-group">
                <div className="form--limit">
                  Updates timeout
                  <Input type="number" label="" standalone className="form-control" ref="timingUpdatesTimeout" defaultValue={current_update_timeout[0]} required={true} min="1" />
                  <div className="form-group group-inline">
                    <Input type="select" placeholder="" ref="timingUpdatesTimeoutUnit" groupClassName="form-group group-inline arrow-icon" defaultValue={current_update_timeout[1]} required={true}>
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
                      <strong>Error!</strong> Please check the form
                    </Alert>
                  </Col>
                  <Col xs={4}>
                    <Button bsStyle="default" className={"plainBtn" + btnStyle} disabled={this.state.isLoading} onClick={this.updateGroup}>{btnContent}</Button>
                  </Col>
                </Row>
              </div>
            </form>
          </div>
        </Modal.Body>
      </Modal>
    );
  }
};

export default ModalUpdate