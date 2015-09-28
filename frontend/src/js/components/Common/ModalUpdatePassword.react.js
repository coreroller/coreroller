import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert } from "react-bootstrap"
import API from "../../api/API"

class ModalUpdatePassword extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      isLoading: false,
      alertVisible: false
    }
    this.handleFocus = this.handleFocus.bind(this)
    this.updatePassword = this.updatePassword.bind(this)
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

  render() {
    let btnStyle = this.state.isLoading ? " loading" : "",
        btnContent = this.state.isLoading ? "Please wait" : "Submit" 

    return (
      <Modal {...this.props} animation={true}>
        <Modal.Header closeButton>
          <Modal.Title id='contained-modal-title-lg'>Change password</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form">
            <form role="form" action="" onFocus={this.handleFocus}>
              <Input type="password" label="*New Password:" ref="password" required={true} />
              <Input type="password" label="*Confirm password:" ref="password2" required={true} />
              <div className="modal--footer">
                <Row>
                  <Col xs={8}>
                    <Alert bsStyle="danger" className={this.state.alertVisible ? "alert--visible" : ""}>
                      <strong>Error!</strong> Please check the form
                    </Alert>
                  </Col>
                  <Col xs={4}>
                    <Button bsStyle="default" className={"plainBtn" + btnStyle} disabled={this.state.isLoading} onClick={this.updatePassword}>{btnContent}</Button>
                  </Col>
                </Row>
              </div>
            </form>
          </div>
        </Modal.Body>
      </Modal>
    )
  }

}

export default ModalUpdatePassword