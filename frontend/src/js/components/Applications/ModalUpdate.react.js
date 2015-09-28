import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Modal, Input, Button, Col, Row, Alert } from "react-bootstrap"

class ModalUpdate extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      isLoading: false,
      alertVisible: false
    }
    this.handleFocus = this.handleFocus.bind(this)
    this.updateApplication = this.updateApplication.bind(this)
  }

  static propTypes : {
    data: PropTypes.object
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

  render() { 
    let btnStyle = this.state.isLoading ? " loading" : "",
        btnContent = this.state.isLoading ? "Please wait" : "Submit" 
    return (
      <Modal {...this.props} animation={true}>
        <Modal.Header closeButton>
          <Modal.Title id="contained-modal-title-lg">Update application</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form">
            <form role="form" action="" onFocus={this.handleFocus}>
              <Input type="text" label="*Name:" ref="nameApp" required defaultValue={this.props.data.name} min={1} maxLength={50} />
              <Input type="textarea" label="Description:" ref="descriptionApp" defaultValue={this.props.data.description} maxLength={100} />
              <div className="modal--footer">
                <Row>
                  <Col xs={8}>
                    <Alert bsStyle="danger" className={this.state.alertVisible ? "alert--visible" : ""}>
                      <strong>Error!</strong> Please check the form
                    </Alert>
                  </Col>
                  <Col xs={4}>
                    <Button bsStyle="default" className={"plainBtn" + btnStyle} disabled={this.state.isLoading} onClick={this.updateApplication}>{btnContent}</Button>
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

export default ModalUpdate