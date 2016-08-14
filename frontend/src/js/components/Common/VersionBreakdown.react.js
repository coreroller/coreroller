import React, { PropTypes } from "react"
import { Col, ProgressBar } from "react-bootstrap"
import _ from "underscore"
import semver from "semver"
import { cleanSemverVersion } from "../../constants/helpers"

class VersionBreakdown extends React.Component {

  constructor() {
    super()
  }

	static PropTypes: {
    version_breakdown: React.PropTypes.array.isRequired,
    channel: React.PropTypes.object.isRequired
  }

  render() {
  	let versions = this.props.version_breakdown ? this.props.version_breakdown : [],
        lastVersionChannel = "",
        entries = [],
        channel = this.props.channel ? this.props.channel : {},
        legendVersion = null

    let versionsValues = (_.map(versions, (version) => {
      return cleanSemverVersion(version.version)
    })).sort(semver.rcompare)

    if (!_.isEmpty(versionsValues)) {
      entries = _.map(versions, function (version, i) {
        let barStyle = "default",
            labelLegend = cleanSemverVersion(version.version)

        if (!_.isEmpty(channel)) {
          lastVersionChannel = channel.package ? cleanSemverVersion(channel.package.version) : ""

          let currentVersionIndex = _.indexOf(versionsValues, lastVersionChannel)

          if (lastVersionChannel) {
            if (cleanSemverVersion(version.version) == lastVersionChannel) {
              barStyle = "success"
              labelLegend = cleanSemverVersion(version.version) + "*"
              legendVersion = <span className="subtitle lowerCase pull-right">{"*Current channel version"}</span>
            } else if (semver.gt(cleanSemverVersion(version.version), lastVersionChannel)) {
              barStyle = "info"
            } else {
              let indexDiff = _.indexOf(versionsValues, cleanSemverVersion(version.version)) - currentVersionIndex
              switch (indexDiff) {
                case 1:
                  barStyle = "warning"
                  break
                case 2:
                  barStyle = "danger"
                  break
              }
            }
          } else {
            legendVersion = <span className="subtitle noTextTransform pull-right">{"No colors available as channel is not pointing to any package"}</span>
          }

        }
        return <ProgressBar striped key={"progressVersion_" + i} bsStyle={barStyle} now={version.percentage} label={labelLegend} />
      })
    }

    return (
      <Col xs={12}>
        <span className="subtitle">Version breakdown</span>
        {legendVersion}
        <ProgressBar>
          {entries}
        </ProgressBar>
      </Col>
    )
  }

}

export default VersionBreakdown
