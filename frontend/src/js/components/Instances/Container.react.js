import { instancesStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import List from "./List.react"
import _ from "underscore"

class Container extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this);
    this.state = {instances: instancesStore.getInstances(props.appID, props.groupID)}
  }

  static PropTypes: {
    appID: React.PropTypes.string.isRequired,
    groupID: React.PropTypes.string.isRequired,
    version_breakdown: React.PropTypes.array.isRequired,
    channel: React.PropTypes.object.isRequired
  }

  componentDidMount() {
    instancesStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    instancesStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      instances: instancesStore.getAll()
    })
  }

  render() {
    let groupInstances = this.state.instances ? this.state.instances[this.props.appID][this.props.groupID] : [] 

    let entries = ""

    if (_.isEmpty(groupInstances)) {
      entries = <div className="emptyBox">No instances have registered yet in this group.<br/><br/>Registration will happen automatically the first time the instance requests an update.</div>
    } else {
      entries = <List 
              instances={groupInstances} 
              version_breakdown={this.props.version_breakdown} 
              channel={this.props.channel} />
    }

    return(
      <div>
        <Row className="noMargin" id="instances">
          <h4 className="instancesList--title">Instances list</h4>
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
