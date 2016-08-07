import React, { PropTypes } from "react"
import StatusHistoryItem from "./StatusHistoryItem.react"

class StatusHistoryList extends React.Component {

  constructor(props) {
    super(props)
  }

  static PropTypes: {
    entries: React.PropTypes.array.isRequired
  }

  render() {
    let entries = this.props.entries ? this.props.entries : []

    return(
      <ul className="timeline--events">
        <li className="timeline--eventsTitle">
          <div>Timestamp</div>
          <div>Status</div>
          <div>Message</div>
        </li>
        {entries.map((entry, i) =>
          <StatusHistoryItem key={"statusHistory_" + i} entry={entry} />
        )}
      </ul>
    )
  }

}

export default StatusHistoryList
