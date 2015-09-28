import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import ApplicationsList from "../Applications/List.react"
import ActivityContainer from "../Activity/Container.react"

class MainLayout extends React.Component {

  constructor() {
    super()
  }

  static PropTypes: {
    stores: React.PropTypes.object.isRequired
  }

  render() {
    return(
      <div className="container">
        <Row>
          <ApplicationsList />
          <ActivityContainer />
        </Row>
      </div>
    )
  }

}

export default MainLayout