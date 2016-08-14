import React, { PropTypes } from "react"

class VersionBullet extends React.Component {

  constructor() {
    super()
  }

  static PropTypes: {
    channel: React.PropTypes.object.isRequired
  }

  render() {
    var divColor = {
      backgroundColor: this.props.channel.color
    }

    return(
      <div className="versionBullet" style={divColor}></div>
    )
  }

}

export default VersionBullet
