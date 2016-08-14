import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Modal, Input, Button, Col, Row, Alert, ButtonInput } from "react-bootstrap"
import { Form, ValidatedInput } from "react-bootstrap-validation"

class ModalUpdate extends React.Component {

  constructor(props) {
    super(props)
    this.handleFocus = this.handleFocus.bind(this)
    this.updateApplication = this.updateApplication.bind(this)
    this.handleValidSubmit = this.handleValidSubmit.bind(this)
    this.handleInvalidSubmit = this.handleInvalidSubmit.bind(this)
    this.exitedModal = this.exitedModal.bind(this)

    this.state = {
      isLoading: false,
      alertVisible: false
    }
}

  static propTypes : {
    data: PropTypes.object.isRequired,
    modalVisible: PropTypes.bool.isRequired
  }

  updateApplication() {
    this.setState({isLoading: true})
    var data = {
      name: this.refs.nameApp.getValue(),
      description: this.refs.descriptionApp.getValue()
    }

    applicationsStore.updateApplication(this.props.data.id, data).
      done(() => {
        this.props.onHide()
        this.setState({isLoading: false})
      }).
      fail(() => {
        this.setState({alertVisible: true, isLoading: false})
      })
  }

  handleFocus() {
    this.setState({alertVisible: false})
  }

  handleValidSubmit() {
    this.updateApplication()
  }

  handleInvalidSubmit() {
    // this.setState({alertVisible: true})
  }

  exitedModal() {
    this.setState({alertVisible: false, isLoading: false})
  }

  render() {
    let btnStyle = this.state.isLoading ? " loading" : "",
        btnContent = this.state.isLoading ? "Please wait" : "Submit"
    return (
      <Modal {...this.props} show={this.props.modalVisible} animation={true} onExited={this.exitedModal}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Update application</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form" onFocus={this.handleFocus}>
            <Form onValidSubmit={this.handleValidSubmit} onInvalidSubmit={this.handleInvalidSubmit}>
              <ValidatedInput
                type="text"
                label="*Name:"
                name="nameApp"
                ref="nameApp"
                defaultValue={this.props.data.name}
                required={true}
                validationEvent="onBlur"
                validate="required,isLength:1:50"
                errorHelp={{
                  required: "Please enter a name",
                  isLength: "Name must be less than 50 characters"
                }}
              />
              <Input type="textarea" label="Description:" ref="descriptionApp" defaultValue={this.props.data.description} maxLength={250} />
              <div className="modal--footer">
                <Row>
                  <Col xs={8}>
                    <Alert bsStyle="danger" className={this.state.alertVisible ? "alert--visible" : ""}>
                      <strong>Error!</strong> The request failed, please check the form
                    </Alert>
                  </Col>
                  <Col xs={4}>
                    <ButtonInput
                      type="submit"
                      bsStyle="default"
                      className={"plainBtn" + btnStyle}
                      disabled={this.state.isLoading}
                      value={btnContent}
                    />
                  </Col>
                </Row>
              </div>
            </Form>
          </div>
        </Modal.Body>
      </Modal>
    )
  }

}

export default ModalUpdate
