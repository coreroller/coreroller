import React, { PropTypes } from "react"
import Router, { Route, DefaultRoute, Redirect } from "react-router"
import Main from "./components/Main.react"
import MainLayout from "./components/Layouts/MainLayout.react"
import ApplicationLayout from "./components/Layouts/ApplicationLayout.react"
import GroupLayout from "./components/Layouts/GroupLayout.react"

var routes = (
  <Route handler={Main}>
    <Route name="MainLayout" path="/apps" handler={MainLayout} />
    <Route name="ApplicationLayout" path="/apps/:appID" handler={ApplicationLayout} />
    <Route name="GroupLayout" path="/apps/:appID/groups/:groupID" handler={GroupLayout}/>
    <DefaultRoute handler={MainLayout}/>
  </Route>
)

Router.run(routes, Router.HashLocation, function(Root, state) {
  React.render(<Root pathParams={state.params}/>, document.body)
})