import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Col, Row } from "react-bootstrap"
import ModalButton from "../Common/ModalButton.react"
import Item from "./Item.react"
import _ from "underscore"

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this);
    this.state = {applications: applicationsStore.getCachedApplications()}
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
    let applications = this.state.applications ? this.state.applications : []
    let entries = ""

    if (_.isEmpty(applications)) {
      entries = <div className="emptyBox">Ops, it looks like you have not created any application yet..<br/><br/> Now is a great time to create your first one, just click on the plus symbol above.</div>
    } else {
      entries = _.map(applications, (application, i) => {
        return <Item key={application.id} application={application} />
      })
    }

    return(
      <Col xs={7}>
        <Row>
          <Col xs={5}>
            <h1 className="displayInline mainTitle">Applications</h1>
            <ModalButton icon="plus" modalToOpen="AddApplicationModal" data={{applications: applications}} />
          </Col>
          <Col xs={7} className="alignRight">
            <div className="searchblock">
              <input type="text" name="searchApps" id="searchApps" placeholder="Search..." autoComplete="off" />
              <label htmlFor="searchApps"></label>
            </div>
          </Col>            
        </Row>
        <Row>
          <Col xs={12} className="apps--container">
            {entries}
          </Col>
        </Row>
      </Col>
    )
  }

}

export default List