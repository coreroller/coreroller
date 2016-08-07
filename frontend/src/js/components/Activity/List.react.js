import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import Item from "./Item.react"

class List extends React.Component {

  constructor() {
    super()
  }

  static PropTypes: {
    day: React.PropTypes.string.isRequired,
    entries: React.PropTypes.array.isRequired,
  }

  render() {
    let entries = this.props.entries ? this.props.entries : []

    return(
      <div>
        <h5 className="timeline--contentTitle">
          {this.props.day}
        </h5>
        <Row>
          <ul className="timeline--content">
            {entries.map((entry, i) =>
              <Item key={i} entry={entry} />
            )}
          </ul>
        </Row>
      </div>      
    )
  }

}

export default List