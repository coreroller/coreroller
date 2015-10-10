import React, { PropTypes } from "react"
import StatusHistoryList from "./StatusHistoryList.react"
import _ from "underscore"

class StatusHistoryContainer extends React.Component {

  constructor(props) {
    super(props)
  }

  static PropTypes: {
    key: React.PropTypes.string.isRequired,
    active: React.PropTypes.array.isRequired,
    instance: React.PropTypes.object.isRequired
  }

  render() {
    let entries = "",
        additionalStyle = ""

    if (_.isEmpty(this.props.instance.statusHistory)) {
      entries = <div className="emptyBox">This instance hasn’t reported any events yet in the context of this application/group.</div>
      additionalStyle = " coreRollerTable-detail--empty"
    } else {
      entries = <StatusHistoryList entries={this.props.instance.statusHistory} />
    }

    return(
      <div className={"coreRollerTable-detail" + additionalStyle + this.props.active} id={"detail-" + this.props.key}>
        <div className="coreRollerTable-detailContent">
          {entries}
        </div>
      </div>
    )
  }

}

export default StatusHistoryContainer