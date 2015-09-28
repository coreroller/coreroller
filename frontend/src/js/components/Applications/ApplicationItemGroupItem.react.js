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
    return(
      <Link to="GroupLayout" params={{appID: this.props.group.application_id, groupID: this.props.group.id}}>
        <span className="activeLink lighter">
          {this.props.group.name} <i className="fa fa-caret-right"></i>
        </span> 
      </Link>
    )
  }

}

export default ApplicationItemGroupItem