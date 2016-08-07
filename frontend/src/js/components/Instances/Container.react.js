import { instancesStore, applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import List from "./List.react"
import _ from "underscore"
import Loader from "halogen/ScaleLoader"
import MiniLoader from "halogen/PulseLoader"

class Container extends React.Component {

  constructor(props) {
    super(props)
    this.onChangeApplications = this.onChangeApplications.bind(this)
    this.onChangeInstances = this.onChangeInstances.bind(this)
    this.onChangeSelectedInstance = this.onChangeSelectedInstance.bind(this)

    this.state = {
      instances: instancesStore.getCachedInstances(props.appID, props.groupID),
      updating: false,
      selectedInstance: ""
    }
  }

  static PropTypes: {
    appID: React.PropTypes.string.isRequired,
    groupID: React.PropTypes.string.isRequired,
    version_breakdown: React.PropTypes.array.isRequired,
    channel: React.PropTypes.object.isRequired
  }

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChangeApplications)
    instancesStore.addChangeListener(this.onChangeInstances)
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChangeApplications)
    instancesStore.removeChangeListener(this.onChangeInstances)
  }

  onChangeSelectedInstance(selectedInstance) {
    this.setState({
      selectedInstance: selectedInstance
    })
  }

  onChangeApplications() {
    instancesStore.getInstances(this.props.appID, this.props.groupID, this.state.selectedInstance)

    this.setState({
      updating: true
    })
  }

  onChangeInstances() {
    this.setState({
      updating: false,
      instances: instancesStore.getCachedInstances(this.props.appID, this.props.groupID)
    })
  }

  render() {
    let groupInstances = this.state.instances,
        miniLoader = this.state.updating ? <MiniLoader color="#00AEEF" size="8px" margin="2px" /> : ""

    let entries = ""

    if (_.isNull(groupInstances)) {
      entries = <div className="icon-loading-container"><Loader color="#00AEEF" size="35px" margin="2px"/></div>
    } else {
      if (_.isEmpty(groupInstances)) {
        entries = <div className="emptyBox">No instances have registered yet in this group.<br/><br/>Registration will happen automatically the first time the instance requests an update.</div>
      } else {
        entries = <List
                instances={groupInstances}
                version_breakdown={this.props.version_breakdown}
                channel={this.props.channel}
                onChangeSelectedInstance={this.onChangeSelectedInstance} />
      }
    }

    return(
      <div>
        <Row className="noMargin" id="instances">
          <h4 className="instancesList--title">Instances list {miniLoader}</h4>
        </Row>
        <Row>
          <Col xs={12}>
            {entries}
          </Col>
        </Row>
      </div>
    )
  }

}

export default Container
