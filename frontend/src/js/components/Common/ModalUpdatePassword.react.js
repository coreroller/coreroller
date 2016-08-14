import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert, ButtonInput } from "react-bootstrap"
import { Form, ValidatedInput } from "react-bootstrap-validation"
import API from "../../api/API"

class ModalUpdatePassword extends React.Component {

  constructor(props) {
    super(props)
    this.handleFocus = this.handleFocus.bind(this)
    this.updatePassword = this.updatePassword.bind(this)
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

  updatePassword() {
    this.setState({isLoading: true})
    let formData = {
      password: this.refs.password.getValue(),
      password2: this.refs.password2.getValue()
    }

    if (formData.password !== formData.password2) {
      this.setState({alertVisible: true, isLoading: false})

    } else {
      let data = {
        username: "admin",
        password: formData.password
      }

      API.updateUserPassword(data).
        done(() => {
          this.props.onHide()
          this.setState({isLoading: false})
        }).
        fail(() => {
          this.setState({alertVisible: true, isLoading: false})
        })
    }
  }

  handleFocus() {
    this.setState({alertVisible: false})
  }

  handleValidSubmit() {
    this.updatePassword()
  }

  handleInvalidSubmit() {
    this.setState({alertVisible: true})
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
          <Modal.Title id='contained-modal-title-lg'>Change password</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form modal--form-with-captions">
            <Form onValidSubmit={this.handleValidSubmit} onInvalidSubmit={this.handleInvalidSubmit} onFocus={this.handleFocus}>
              <ValidatedInput
                type="password"
                label="*New Password:"
                name="password"
                ref="password"
                required={true}
                validate="required"
                errorHelp={{
                  required: "Please enter a password"
                }}
              />
              <ValidatedInput
                type="password"
                label="*Confirm password:"
                name="password2"
                ref="password2"
                required={true}
                validate={(val, context) => val === context.password}
                errorHelp="Passwords do not match"
              />
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

export default ModalUpdatePassword
