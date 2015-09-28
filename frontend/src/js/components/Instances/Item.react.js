import { instancesStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import moment from "moment"
import { Label } from "react-bootstrap"

class Item extends React.Component {

  constructor(props) {
    super(props)
    this.state = {status: {}}
    this.fetchStatusFromStore = this.fetchStatusFromStore.bind(this)
  }

  static PropTypes: {
    instance: React.PropTypes.object.isRequired,
    key: React.PropTypes.number.isRequired,
    selected: React.PropTypes.bool,
    versionNumbers: React.PropTypes.array,
    styles: React.PropTypes.array
  }

  fetchStatusFromStore() {
    let status = instancesStore.getInstanceStatus(this.props.instance.application.status)
    this.setState({status: status})
  }

  componentDidMount() {
    this.fetchStatusFromStore()
  }

  render() {
    let date = moment(this.props.instance.application.last_check_for_updates).format("DD/MM/YYYY, hh:mma"),
        active = this.props.selected ? " active" : "",
        index = this.props.versionNumbers.indexOf(this.props.instance.application.version),
        boxStyle = "default"
    
    if (index >= 0 && index < this.props.styles.length) {
      boxStyle = this.props.styles[index]
    }

    let downloadingIcon = this.state.status.spinning ? <img src="img/animated_dots.gif" /> : "",
        instanceLabel = this.state.status.className ? <Label>{downloadingIcon} {this.state.status.description}</Label> : <div>&nbsp;</div>

    return(
      <div className="instance">
        <div className="coreRollerTable-body">
          <div className="coreRollerTable-cell lightText">
            <p>
              {this.props.instance.ip} 
            </p>
          </div>
          <div className="coreRollerTable-cell coreRollerTable-cell--medium">
            <p>{this.props.instance.id}</p>
          </div>
          <div className="coreRollerTable-cell">
            {instanceLabel}
          </div>
          <div className="coreRollerTable-cell">
            <p className={"box--" + boxStyle}>{this.props.instance.application.version}</p>
          </div>
          <div className="coreRollerTable-cell">
            <p>{date}</p>
          </div>   
        </div>
      </div>
    )
  }

}

export default Item