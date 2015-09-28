import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import Router, { Link } from "react-router"
import _ from "underscore"
import Item from "./Item.react"
import ModalButton from "../Common/ModalButton.react"

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this);
    this.state = {applications: applicationsStore.getCachedApplications()}
  }

  static propTypes: {
    appID: React.PropTypes.string.isRequired
  }

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      applications: applicationsStore.getCachedApplications()
    })
  }

  render() {
    let application = _.findWhere(this.state.applications, {id: this.props.appID})

    let channels = [],
        groups = [],
        packages = [],
        instances = 0,
        name = ""

    if (application) {
      name = application.name
      groups = application.groups ? application.groups : []
      packages = application.packages ? application.packages : []
      instances = application.instances ? application.instances : []
      channels = application.channels ? application.channels : []
    }

    let entries = ""

    if (_.isEmpty(groups)) {
      entries = <div className="emptyBox">There are no groups for this application yet.<br/><br/>Groups help you control how you want to distribute updates to a specific set of instances.</div>
    } else {
      entries = _.map(groups, (group, i) => {
        return <Item key={group.id} group={group} appName={name} channels={channels} />
      })
    }

		return (
      <Col xs={8}>
        <Row>
          <Col xs={5}>
            <h1 className="displayInline mainTitle">Groups</h1>
            <ModalButton icon="plus" modalToOpen="AddGroupModal" data={{channels: channels, appID: this.props.appID}} />
          </Col>
          <Col xs={7} className="alignRight">
            <div className="searchblock">
              <input type="text" name="searchApps" id="searchApps" placeholder="Search..." />
              <label htmlFor="searchApps"></label>
            </div>
          </Col>            
        </Row>
        <Row>
          <Col xs={12} className="groups--container">
            {entries}
          </Col>
        </Row>
      </Col>
		)
  }

}

export default List