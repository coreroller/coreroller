import API from "../api/API"
import React, { PropTypes } from "react"
import Router, { Link } from "react-router"
import { Navbar, Nav, NavItem, DropdownButton, MenuItem, Button } from "react-bootstrap"
import { NavItemLink } from "react-router-bootstrap"
import ModalUpdatePassword from "./Common/ModalUpdatePassword.react"

class Header extends React.Component {

  constructor() {
    super()
    this.close = this.close.bind(this)
    this.open = this.open.bind(this)

    this.state = {showModal: false}
  }

  logout() {
    API.logout();
  }

  close() {
    this.setState({showModal: false})
  }

  open() {
    this.setState({showModal: true})
  }

  render() {
  	var brand = <Link to="MainLayout">Core<span className="blueStyle">Roller</span></Link>
    var options = {
      show: this.state.showModal
    }

    return (
    	<Navbar brand={brand} fixedTop={true} toggleNavKey={0}>
    		<Nav right eventKey={1}>
			    <NavItemLink eventKey={1} to="MainLayout"><span className="fa fa-server"></span> Applications</NavItemLink>
          <li>
            <DropdownButton bsStyle="link" title="My account" key={2}>
              <MenuItem eventKey="2">
                <Button bsStyle="link" onClick={this.open.bind()} id="openModal-updatePassword">
                  <span className="fa fa-pencil-square-o"></span> Change password
                  <ModalUpdatePassword {...options} onHide={this.close} />
                </Button>
              </MenuItem>
              <MenuItem divider />
              <MenuItem eventKey="3"><Button bsStyle="link" onClick={this.logout}><span className="fa fa-sign-out"></span> Log out</Button></MenuItem>
            </DropdownButton>
          </li>
			  </Nav>
	   	</Navbar>
    )
  }
}

export default Header