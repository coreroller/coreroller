import React, { PropTypes } from "react"
import PubSub from "pubsub-js"
import ProgressBarJS from "progressbar.js"

class ProgressBar extends React.Component {

  constructor(props) {
    super(props)

    this.line = null
    this.inProgress = false
    this.inProgressCount = 0
  }

  static PropTypes: {
    name: React.PropTypes.string.isRequired,
    color: React.PropTypes.string.isRequired,
    width: React.PropTypes.number.isRequired
  }

  componentDidMount() {
    let lineContainer = React.findDOMNode(this.refs.progressBar)
    let lineOptions = {
      color: this.props.color,
      strokeWidth: this.props.width,
      easing: "easeInOut"
    }
    this.line = new ProgressBarJS.Line(React.findDOMNode(lineContainer), lineOptions)
  }

  componentWillMount() {
    PubSub.subscribe(this.props.name, (t, m) => { return this.handleMsg(m) })
  }

  componentWillUnmount() {
    PubSub.unsubscribe(this.props.name)
  }

  render() {
    return React.createElement("div", {
      className: this.props.containerClassName, 
      ref: "progressBar"
    })
  }

  handleMsg(msg) {
    if (this.inProgress) {
      switch (msg) {
        case "add":
          this.inProgressCount++
          break
        case "done":
          if (this.inProgressCount > 0) {
            this.inProgressCount--
          }
          if (this.inProgressCount == 0) {
            this.line.animate(1.0, {duration: 200}, () => {
              this.line.set(0)
              this.inProgress = false
            })
          }
          break
      }
    } else {
      switch (msg) {
        case "add":
          this.inProgressCount++
          this.inProgress = true
          this.line.animate(0.25, {duration: 5000}, () => {
            this.line.animate(0.75, {duration: 10000})
          })
          break
      }      
    }
  }

};

export default ProgressBar