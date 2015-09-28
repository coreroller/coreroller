import React, { PropTypes } from "react"
import AddApplicationModal from "../Applications/ModalAdd.react"
import UpdateApplicationModal from "../Applications/ModalUpdate.react"
import AddGroupModal from "../Groups/ModalAdd.react"
import UpdateGroupModal from "../Groups/ModalUpdate.react"
import AddChannelModal from "../Channels/ModalAdd.react"
import UpdateChannelModal from "../Channels/ModalUpdate.react"
import AddPackageModal from "../Packages/ModalAdd.react"
import UpdatePackageModal from "../Packages/ModalUpdate.react"

class ModalButton extends React.Component {

  constructor(props) {
    super(props)
    this.state = {showModal: false}
    this.close = this.close.bind(this)
    this.open = this.open.bind(this)
  }

  static propTypes : {
    icon: PropTypes.string.isRequired,
    modalToOpen: PropTypes.string.isRequired,
    data: PropTypes.object
  }

  close() {
    this.setState({showModal: false})
  }

  open() {
    this.setState({showModal: true})
  }

  render() {
    var options = {
      show: this.state.showModal,
      data: this.props.data
    }

    switch (this.props.modalToOpen) {
      case "AddApplicationModal":
        var modal = <AddApplicationModal {...options} onHide={this.close} />
        break
      case "UpdateApplicationModal":
        var modal = <UpdateApplicationModal {...options} onHide={this.close} />
        break
      case "AddGroupModal":
        var modal = <AddGroupModal {...options} onHide={this.close} />
        break
      case "UpdateGroupModal":
        var modal = <UpdateGroupModal {...options} onHide={this.close} />
        break
      case "AddChannelModal":
        var modal = <AddChannelModal {...options} onHide={this.close} />
        break
      case "UpdateChannelModal":
        var modal = <UpdateChannelModal {...options} onHide={this.close} />
        break
      case "AddPackageModal":
        var modal = <AddPackageModal {...options} onHide={this.close} />
        break
      case "UpdatePackageModal":
        var modal = <UpdatePackageModal {...options} onHide={this.close} />
        break
    }
    
    return(
      <a className={"cr-button displayInline fa fa-" + this.props.icon} href="javascript:void(0)" onClick={this.open.bind()} id={"openModal-" + this.props.modalToOpen}>
        {modal}
      </a>
    )
  }

}

export default ModalButton