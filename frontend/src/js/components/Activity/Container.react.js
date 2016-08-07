import { activityStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import List from "./List.react"
import _ from "underscore"
import Loader from "halogen/ScaleLoader"

class Container extends React.Component {

  constructor() {
    super()
    this.onChange = this.onChange.bind(this);

    this.state = {entries: activityStore.getCachedActivity()}
  }

  componentDidMount() {
    activityStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    activityStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      entries: activityStore.getCachedActivity()
    })
  }

  render() {
    let entries = ""

    if (_.isNull(this.state.entries)) {
      entries = <div className="icon-loading-container"><Loader color="#00AEEF" size="35px" margin="2px"/></div>
    } else {
      if (_.isEmpty(this.state.entries)) {
        entries = <div className="emptyBox">No activity found for the last week.<br/><br/>You will see here important events related to the rollout of your updates. Stay tuned!</div>
      } else {
        entries = _.mapObject(this.state.entries, (entry, key) => {
          return <List day={key} entries={entry} key={key} />
        })
      }      
    }

    return(
      <Col xs={5} className="timeline--container">
        <Row>
          <Col xs={12}>
            <h1 className="displayInline mainTitle padBottom25">Activity</h1>
          </Col>
        </Row>
        {entries}
      </Col>
    )
  }

}

export default Container