import React, { PropTypes } from "react"
import Router, { Link } from "react-router"

class ApplicationItemGroupItem extends React.Component {

  constructor() {
    super()
  } 

  static PropTypes: {
    group: React.PropTypes.object.isRequired,
    appName: React.PropTypes.string.isRequired
  }

  render() {
    const instances_total = this.props.group.instances_stats.total ? "(" + this.props.group.instances_stats.total + ")" : ""

    return(
      <Link to="GroupLayout" params={{appID: this.props.group.application_id, groupID: this.props.group.id}}>
        <span className="activeLink lighter">
          {this.props.group.name} {instances_total} <i className="fa fa-caret-right"></i>
        </span> 
      </Link>
    )
  }

}

export default ApplicationItemGroupItem