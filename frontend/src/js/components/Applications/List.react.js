import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Col, Row, Button } from "react-bootstrap"
import ModalButton from "../Common/ModalButton.react"
import Item from "./Item.react"
import _ from "underscore"
import Loader from "halogen/ScaleLoader"
import SearchInput from "react-search-input"

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this)
    this.searchUpdated = this.searchUpdated.bind(this)
    this.state = {applications: applicationsStore.getCachedApplications(), searchTerm: ""}
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

  searchUpdated(term) {
    this.setState({searchTerm: term})
  }

  render() {
    let applications = this.state.applications ? this.state.applications : [],
        entries = ""

    if (this.refs.search) {
      var filters = ["name"]
      applications = applications.filter(this.refs.search.filter(filters))
    }

    if (_.isNull(applications)) {
      entries = <div className="icon-loading-container"><Loader color="#00AEEF" size="35px" margin="2px"/></div>
    } else {
      if (_.isEmpty(applications)) {
        if (this.state.searchTerm) {
          entries = <div className="emptyBox">No results found.</div>
        } else {
          entries = <div className="emptyBox">Ops, it looks like you have not created any application yet..<br/><br/> Now is a great time to create your first one, just click on the plus symbol above.</div>
        }
      } else {
        entries = _.map(applications, (application, i) => {
          return <Item key={application.id} application={application} />
        })
      }
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
              <SearchInput ref="search" onChange={this.searchUpdated} placeholder="Search..." />
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