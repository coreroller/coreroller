import React, { PropTypes } from "react"
import { Col, ProgressBar, OverlayTrigger, Tooltip } from "react-bootstrap"
import _ from "underscore"

class VersionBreakdown extends React.Component {

  constructor() {
    super()
  }

	static PropTypes: {
    version_breakdown: React.PropTypes.array.isRequired,
    channel: React.PropTypes.object
  }

  render() {
  	let styles = ["success", "warning", "danger"],
        versions = this.props.version_breakdown ? this.props.version_breakdown : [],
        lastVersionChannel = "",
        lastVersionBD = "",
        entries = []

    if (!_.isNull(this.props.channel)) {
      lastVersionChannel = this.props.channel.package ? this.props.channel.package.version : ""
      lastVersionBD = versions[0] ? versions[0].version : ""

      // Removed success style if no instances with last version
      if (lastVersionBD !== lastVersionChannel) {
        styles.shift()
      }

      entries = _.map(versions, function (version, i) {
        let barStyle = "default"

        if (i < styles.length) {
          barStyle = styles[i]
        }
        return <ProgressBar striped key={i} bsStyle={barStyle} now={version.percentage} label={version.version} />
      })
    }

    return (
      <Col xs={12}>
        <span className="subtitle">Version breakdown</span>
        <ProgressBar>
          {entries}
        </ProgressBar>
      </Col>
    )
  }

}

export default VersionBreakdown