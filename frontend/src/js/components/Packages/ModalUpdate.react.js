import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col, Modal, Input, Button, Alert } from "react-bootstrap"

class ModalUpdate extends React.Component {

  constructor(props) {
    super(props)
    this.state = {
      isLoading: false,
      alertVisible: false
    }
    this.handleFocus = this.handleFocus.bind(this)
    this.updatePackage = this.updatePackage.bind(this)
  }

  static propTypes : {
    data: PropTypes.object
  }

  updatePackage() {
    this.setState({isLoading: true})
    let data = {
      id: this.props.data.channel.id,
      filename: this.refs.filenamePackage.getValue(),
      description: this.refs.descriptionPackage.getValue(),
      url: this.refs.urlPackage.getValue(),
      version: this.refs.versionPackage.getValue(),
      type: parseInt(this.refs.typePackage.getValue()),
      size: parseInt(this.refs.sizePackage.getValue()),
      hash: this.refs.hashPackage.getValue(),
      application_id: this.props.data.channel.application_id
    }

    applicationsStore.updatePackage(data).
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
          <Modal.Title id="contained-modal-title-lg">Update package</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <div className="modal--form">
            <form role="form" action="" onFocus={this.handleFocus}>
              <Input type="select" label="*Type:" defaultValue={this.props.data.channel.type} placeholder="" groupClassName="arrow-icon" ref="typePackage" required={true}>
                <option value={1}>Coreos</option>
                <option value={2}>Docker image</option>
                <option value={3}>Rocket image</option>
                <option value={4}>Other</option>
              </Input>      
              <Input type="url" label="*Url:" defaultValue={this.props.data.channel.url} ref="urlPackage" required={true} macLength={256} />
              <Input type="text" label="Filename:" defaultValue={this.props.data.channel.filename} ref="filenamePackage" maxLength={100} />
              <Input type="textarea" label="Description:" defaultValue={this.props.data.channel.description} ref="descriptionPackage" maxLength={250} />
              <Row>
                <Col xs={6}>
                  <Input type="text" label="*Version:" defaultValue={this.props.data.channel.version} ref="versionPackage" required={true} />        
                  <div className="form--legend minlegend minlegend--formGroup">Use SemVer format (1.0.1)</div>                  
                </Col>
                <Col xs={6}>
                  <Input type="number" label="Size:" defaultValue={this.props.data.channel.size} ref="sizePackage" maxLength={20} />        
                </Col>
              </Row>
              <Input type="text" label="Hash:" defaultValue={this.props.data.channel.hash} ref="hashPackage" maxLength={64} />  
              <div className="modal--footer">
                <Row>
                  <Col xs={8}>
                    <Alert bsStyle="danger" className={this.state.alertVisible ? "alert--visible" : ""}>
                      <strong>Error!</strong> Please check the form
                    </Alert>
                  </Col>
                  <Col xs={4}>
                    <Button bsStyle="default" className={"plainBtn" + btnStyle} disabled={this.state.isLoading} onClick={this.updatePackage}>{btnContent}</Button>
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
