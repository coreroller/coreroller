import React, { PropTypes } from "react"
import ChannelLabel from "../Common/ChannelLabel.react"
import _ from "underscore"

class ApplicationItemChannelsList extends React.Component {

  constructor() {
    super()
  }

  static PropTypes: {
    channels: React.PropTypes.array.isRequired
  }

  render() {
    let channels = this.props.channels ? this.props.channels : [],
        entries = ""

    if (_.isEmpty(channels)) {
      entries = "-"
    } else {
      entries = _.map(channels, (channel, i) => {
        return <ChannelLabel channel={channel} key={"channel_" + i} />
      })
    }

    return(
      <ul className={_.isEmpty(channels) ? "apps--channelsList apps--channelsList--extraPadding" : "apps--channelsList"}>
        {entries}
      </ul>
    )
  }

}

export default ApplicationItemChannelsList
