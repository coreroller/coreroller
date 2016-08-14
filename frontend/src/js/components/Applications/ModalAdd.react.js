import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert, ButtonInput } from "react-bootstrap"
import { Form, ValidatedInput } from "react-bootstrap-validation"

class ModalAdd extends React.Component {

  constructor(props) {
    super(props)
    this.handleFocus = this.handleFocus.bind(this)
    this.createApplication = this.createApplication.bind(this)
    this.handleValidSubmit = this.handleValidSubmit.bind(this)
    this.handleInvalidSubmit = this.handleInvalidSubmit.bind(this)
    this.exitedModal = this.exitedModal.bind(this)

    this.state = {
      isLoading: false,
      alertVisible: false
    }
  }

  static propTypes : {
    data: PropTypes.object
  }

  createApplication() {
    this.setState({isLoading: true})
    let data = {
      name: this.refs.nameNewApp.getValue(),
      description: this.refs.descriptionNewApp.getValue()
    }

    let clonedApplication = this.refs.cloningNewApp.getValue()

    applicationsStore.createApplication(data, clonedApplication).
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
    this.createApplication()
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
      <Modal {...this.props} animation={true} onExited={this.exitedModal}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Add new application</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form" onFocus={this.handleFocus}>
            <Form onValidSubmit={this.handleValidSubmit} onInvalidSubmit={this.handleInvalidSubmit}>
              <ValidatedInput
                type="text"
                label="*Name:"
                name="nameNewApp"
                ref="nameNewApp"
                required={true}
                validationEvent="onBlur"
                validate="required,isLength:1:50"
                errorHelp={{
                  required: "Please enter a name",
                  isLength: "Name must be less than 50 characters"
                }}
              />
              <Input type="textarea" label="Description:" ref="descriptionNewApp" maxLength={250} />
              <Input type="select" label="Clone channels/groups from:" placeholder="" groupClassName="arrow-icon" ref="cloningNewApp">
                <option value="" />
                { this.props.data.applications &&
                  this.props.data.applications.map((application, i) =>
                    <option value={application.id} key={"app_" + i}>{application.name}</option>
                  )
                }
              </Input>
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

export default ModalAdd
