import React, { PropTypes } from "react"
import Router, { RouteHandler } from "react-router"
import Header from "./Header.react"
import ProgressBar from "./ProgressBar.react"

class Main extends React.Component {

  constructor() {
    super()
  }

  render() {
    return (
      <div>
        <Header />
        <ProgressBar name="main_progress_bar" color="#ddd" width={0.2} /> 
        <RouteHandler />
      </div>
    )
  }

}

export default Main