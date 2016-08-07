import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import Router, { Link } from "react-router"
import _ from "underscore"
import Item from "./Item.react"
import ModalButton from "../Common/ModalButton.react"
import SearchInput from "react-search-input"
import Loader from "halogen/ScaleLoader"
import MiniLoader from "halogen/PulseLoader"
import ModalUpdate from "./ModalUpdate.react"

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this)
    this.searchUpdated = this.searchUpdated.bind(this)
    this.openUpdateGroupModal = this.openUpdateGroupModal.bind(this)
    this.closeUpdateGroupModal = this.closeUpdateGroupModal.bind(this)

    this.state = {
      application: applicationsStore.getCachedApplication(this.props.appID),
      searchTerm: "",
      updateGroupModalVisible: false,
      updateGroupIDModal: null,
      updateAppIDModal: null
    }
  }

  static propTypes: {
    appID: React.PropTypes.string.isRequired
  }

  closeUpdateGroupModal() {
    this.setState({updateGroupModalVisible: false})
  }

  openUpdateGroupModal(appID, groupID) {
    this.setState({updateGroupModalVisible: true, updateGroupIDModal: groupID, updateAppIDModal: appID})
  }

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      application: applicationsStore.getCachedApplication(this.props.appID)
    })
  }

  searchUpdated(term) {
    this.setState({searchTerm: term})
  }

  render() {
    let application = this.state.application

    let channels = [],
        groups = [],
        packages = [],
        instances = 0,
        name = "",
        entries = ""

    const miniLoader = <div className="icon-loading-container"><MiniLoader color="#00AEEF" size="12px" /></div>

    if (application) {
      name = application.name
      groups = application.groups ? application.groups : []
      packages = application.packages ? application.packages : []
      instances = application.instances ? application.instances : []
      channels = application.channels ? application.channels : []

      if (this.refs.search) {
        var filters = ["name"]
        groups = groups.filter(this.refs.search.filter(filters))
      }

      if (_.isEmpty(groups)) {
        if (this.state.searchTerm) {
          entries = <div className="emptyBox">No results found.</div>
        } else {
          entries = <div className="emptyBox">There are no groups for this application yet.<br/><br/>Groups help you control how you want to distribute updates to a specific set of instances.</div>
        }
      } else {
        entries = _.map(groups, (group, i) => {
          return <Item key={"groupID_" + group.id} group={group} appName={name} channels={channels} handleUpdateGroup={this.openUpdateGroupModal} />
        })
      }

    } else {
      entries = <div className="icon-loading-container"><Loader color="#00AEEF" size="35px" margin="2px"/></div>
    }

    const groupToUpdate =  !_.isEmpty(groups) && this.state.updateGroupIDModal ? _.findWhere(groups, {id: this.state.updateGroupIDModal}) : null

		return (
      <div>
        <Col xs={8}>
          <Row>
            <Col xs={5}>
              <h1 className="displayInline mainTitle">Groups</h1>
              <ModalButton icon="plus" modalToOpen="AddGroupModal" data={{channels: channels, appID: this.props.appID}} />
            </Col>
            <Col xs={7} className="alignRight">
              <div className="searchblock">
                <SearchInput ref="search" onChange={this.searchUpdated} placeholder="Search..." />
                <label htmlFor="searchGroups"></label>
              </div>
            </Col>
          </Row>
          <Row>
            <Col xs={12} className="groups--container">
              {entries}
            </Col>
          </Row>
        </Col>
        {/* Update group modal */}
        {groupToUpdate &&
          <ModalUpdate
            data={{group: groupToUpdate, channels: channels}}
            modalVisible={this.state.updateGroupModalVisible}
            onHide={this.closeUpdateGroupModal} />
        }
      </div>
		)
  }

}

export default List
