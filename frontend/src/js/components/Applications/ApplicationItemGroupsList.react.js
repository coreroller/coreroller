import React, { PropTypes } from "react"
import Item from "./ApplicationItemGroupItem.react"

class ApplicationItemGroupsList extends React.Component {

  constructor() {
    super()
  }

  static PropTypes: {
    groups: React.PropTypes.array.isRequired,
    appID: React.PropTypes.string.isRequired,
    appName: React.PropTypes.string.isRequired
  };

  render() {
    return(
      <span className="apps--groupsList">
        {this.props.groups.map((group, i) =>
          <Item key={"group_" + i} group={group} appID={this.props.appID} appName={this.props.appName} />
        )}
      </span>
    )
  }

}

export default ApplicationItemGroupsList