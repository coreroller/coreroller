import { applicationsStore } from "../../stores/Stores"
import React, { PropTypes } from "react"
import { Row, Col } from "react-bootstrap"
import _ from "underscore"
import ModalButton from "../Common/ModalButton.react"
import Item from "./Item.react"

class List extends React.Component {

  constructor(props) {
    super(props)
    this.onChange = this.onChange.bind(this);
    this.state = {applications: applicationsStore.getCachedApplications()}
  }

  static propTypes: {
    appID: React.PropTypes.string.isRequired
  };

  componentDidMount() {
    applicationsStore.addChangeListener(this.onChange)
  }

  componentWillUnmount() {
    applicationsStore.removeChangeListener(this.onChange)
  }

  onChange() {
    this.setState({
      applications: applicationsStore.getCachedApplications()
    })
  }

  render() {
    let application = _.findWhere(this.state.applications, {id: this.props.appID}),
        channels = [],
        packages = []

    if (application) {
      channels = application.channels ? application.channels : []
      packages = application.packages ? application.packages : []
    }

    let entries = "";

    if (_.isEmpty(channels)) {
      entries = <div className="emptyBox">This application does not have any channel yet</div>;
    } else {
      entries = _.map(channels, (channel, i) => {
        return <Item key={channel.id} channel={channel} packages={packages} />
      })
    }

    return (
      <div>
        <Row>
          <Col xs={12}>
            <h1 className="displayInline mainTitle">Channels</h1>
            <ModalButton 
              icon="plus" 
              modalToOpen="AddChannelModal" 
              data={{packages: packages, applicationID: this.props.appID}} />
          </Col>
        </Row>
        <div className="groups--channelsList">
          {entries}
        </div>
      </div>
    );

  }
};

export default List