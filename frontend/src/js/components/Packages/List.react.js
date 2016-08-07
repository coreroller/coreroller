import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import _ from "underscore"
import ModalButton from "../Common/ModalButton.react"
import Item from "./Item.react"
import ModalUpdate from "./ModalUpdate.react"

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this)
    this.closeUpdatePackageModal = this.closeUpdatePackageModal.bind(this)
    this.openUpdatePackageModal = this.openUpdatePackageModal.bind(this)

    this.state = {
      channels: applicationsStore.getCachedChannels(props.appID),
      packages: applicationsStore.getCachedPackages(props.appID),
      updatePackageModalVisible: false,
      updatePackageIDModal: null
    }
  }

  static propTypes: {
    appID: React.PropTypes.string.isRequired
  }

  closeUpdatePackageModal() {
    this.setState({updatePackageModalVisible: false})
  }

  openUpdatePackageModal(packageID) {
    this.setState({updatePackageModalVisible: true, updatePackageIDModal: packageID})
  }

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      channels: applicationsStore.getCachedChannels(this.props.appID),
      packages: applicationsStore.getCachedPackages(this.props.appID)
    })
  }

  render() {
    const channels = this.state.channels ? this.state.channels : [],
          packages = this.state.packages ? this.state.packages : []

    let entries = ""

    if (_.isEmpty(packages)) {
      entries = <div className="emptyBox">This application does not have any package yet</div>
    } else {
      entries = _.map(packages, (packageItem, i) => {
        return <Item key={"packageItemID_" + packageItem.id} packageItem={packageItem} channels={channels} handleUpdatePackage={this.openUpdatePackageModal} />
      })
    }

    const packageToUpdate =  !_.isEmpty(packages) && this.state.updatePackageIDModal ? _.findWhere(packages, {id: this.state.updatePackageIDModal}) : null

    return (
      <div>
        <Row>
          <Col xs={12}>
            <h1 className="displayInline mainTitle">Packages</h1>
            <ModalButton icon="plus" modalToOpen="AddPackageModal"
            data={{channels: channels, appID: this.props.appID}} />
          </Col>
        </Row>
        <div className="groups--packagesList">
          {entries}
        </div>
        {/* Update package modal */}
        {packageToUpdate &&
          <ModalUpdate
            data={{channels: this.props.channels, channel: packageToUpdate}}
            modalVisible={this.state.updatePackageModalVisible}
            onHide={this.closeUpdatePackageModal} />
        }
      </div>
    )
  }

}

export default List
