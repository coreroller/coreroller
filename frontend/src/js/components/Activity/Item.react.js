import { activityStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import moment from "moment"

class Item extends React.Component {

  constructor(props) {
    super(props)

    this.state = {
      entryClass: {},
      entrySeverity: {}
    }
  }

  static PropTypes: {
    entry: React.PropTypes.object.isRequired
  };

  fetchEntryClassFromStore() {
    let entryClass = activityStore.getActivityEntryClass(this.props.entry.class, this.props.entry)
    this.setState({
      entryClass: entryClass
    })
  }

  fetchEntrySeverityFromStore() {
    let entrySeverity = activityStore.getActivityEntrySeverity(this.props.entry.severity)
    this.setState({
      entrySeverity: entrySeverity
    })
  }

  componentDidMount() {
    this.fetchEntryClassFromStore()
    this.fetchEntrySeverityFromStore()
  }

  render() {
    let ampm = moment.utc(this.props.entry.created_ts).local().format("a"),
        time = moment.utc(this.props.entry.created_ts).local().format("HH:mm"),
        subtitle = "",
        name = ""

    if (this.state.entryClass.type !== "activityChannelPackageUpdated") {
      subtitle = "GROUP:"
      name = this.state.entryClass.groupName
    }

    return (
      <li className = {this.state.entrySeverity.className}>
        <div className="timeline--icon">
          <span className={"fa " + this.state.entrySeverity.icon}></span>
        </div>
        <div className="timeline--event">
          {time}
          <br />
          <span className="timeline--ampm">{ampm}</span>
        </div>
        <div className="timeline--eventLabel">
          <div className="row timeline--eventLabelTitle">
            <div className="col-xs-5 noPadding">{this.state.entryClass.appName}</div>
            <div className="col-xs-7 noPadding">
              <span className="subtitle">{subtitle} </span>
              {name}
            </div>
          </div>
          <p>{this.state.entryClass.description}</p>
        </div>
      </li>
    )
  }

}

export default Item
