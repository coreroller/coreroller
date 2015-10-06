import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert } from "react-bootstrap"

class ModalAdd extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      isLoading: false,
      alertVisible: false
    }
    this.handleFocus = this.handleFocus.bind(this)
    this.createApplication = this.createApplication.bind(this)
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

  render() {
    let btnStyle = this.state.isLoading ? " loading" : "",
        btnContent = this.state.isLoading ? "Please wait" : "Submit" 

    return (
      <Modal {...this.props} animation={true}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Add new application</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form">
            <form role="form" action="" onFocus={this.handleFocus}>
              <Input type="text" label="*Name:" ref="nameNewApp" required={true} min="1" maxLength={50} />
              <Input type="textarea" label="Description:" ref="descriptionNewApp" maxLength={250} />
              <Input type="select" label="Clone channels/groups from:" placeholder="" groupClassName="arrow-icon" ref="cloningNewApp">
                <option value="" />
                {this.props.data.applications.map((application, i) =>
                  <option value={application.id} key={i}>{application.name}</option>
                )}
              </Input>
              <div className="modal--footer">
                <Row>
                  <Col xs={8}>
                    <Alert bsStyle="danger" className={this.state.alertVisible ? "alert--visible" : ""}>
                      <strong>Error!</strong> Please check the form
                    </Alert>
                  </Col>
                  <Col xs={4}>
                    <Button bsStyle="default" className={"plainBtn" + btnStyle} disabled={this.state.isLoading} onClick={this.createApplication}>{btnContent}</Button>
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

export default ModalAdd
