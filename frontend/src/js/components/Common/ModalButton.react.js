import React, { PropTypes } from "react"
import AddApplicationModal from "../Applications/ModalAdd.react"
import AddGroupModal from "../Groups/ModalAdd.react"
import AddChannelModal from "../Channels/ModalAdd.react"
import AddPackageModal from "../Packages/ModalAdd.react"

class ModalButton extends React.Component {

  constructor(props) {
    super(props)
    this.close = this.close.bind(this)
    this.open = this.open.bind(this)

    this.state = {showModal: false}
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
      case "AddGroupModal":
        var modal = <AddGroupModal {...options} onHide={this.close} />
        break
      case "AddChannelModal":
        var modal = <AddChannelModal {...options} onHide={this.close} />
        break
      case "AddPackageModal":
        var modal = <AddPackageModal {...options} onHide={this.close} />
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
