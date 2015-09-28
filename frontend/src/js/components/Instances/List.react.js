import React, { PropTypes } from "react"
import Item from "./Item.react"
import _ from "underscore"

class List extends React.Component {

  constructor(props) {
    super(props)
    this.state = {selections: {}}
    this.onItemToggle = this.onItemToggle.bind(this)
  }

  static PropTypes: {
    instances: React.PropTypes.array.isRequired,
    version_breakdown: React.PropTypes.array,
    channel: React.PropTypes.object
  }

  onItemToggle(id, selected) {
    let selections = this.state.selections
    if (selected == true) {
      _.mapObject(selections, (val, key) => {
        if (val == true) {
          selections[key] = false;
        }
      })
    }
    selections[id] = selected;

    this.setState({
      selections: selections
    })
  }

  render() {
    let versions = this.props.version_breakdown ? this.props.version_breakdown : []
    
    let versionNumbers = _.map(versions, (version, i) => {
      return version.version
    })
    
    let lastVersionChannel = this.props.channel.package ? this.props.channel.package.version : "",
        lastVersionBD = versionNumbers[0] ? versionNumbers[0] : "",
        styles = ["success", "warning", "danger"]

    // Removed success style if no instances with last version
    if (lastVersionBD !== lastVersionChannel) {
      styles.shift()
    }    

    return(
      <div className="coreRollerTable">
        <div className="coreRollerTable-header">
          <div className="coreRollerTable-cell">IP</div>
          <div className="coreRollerTable-cell coreRollerTable-cell--medium">Instance ID</div>
          <div className="coreRollerTable-cell">Current status</div>
          <div className="coreRollerTable-cell">Version</div>
          <div className="coreRollerTable-cell">Last check</div>
        </div>
        {this.props.instances.map((instance, i) =>
          <Item key={i} instance={instance} styles={styles} versionNumbers={versionNumbers} selected={this.state.selections[instance.id]} onToggle={this.onItemToggle} />
        )}
      </div>
    )
  }

}

export default List