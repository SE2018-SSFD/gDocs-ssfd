import React from 'react';
import {Route, Redirect} from 'react-router-dom'
import * as userService from "../api/userService"

export class LoginRoute extends React.Component{
    constructor(props) {
        super(props);
        this.state = {
            isAuthed: false,
            hasAuthed: false,
        };
    }

    checkAuth = (data) => {
        if (data.success === true) {
            this.setState({isAuthed: true, hasAuthed: true});
        } else {
            // localStorage.removeItem('token');
            this.setState({isAuthed: false, hasAuthed: true});
        }
    };


    componentDidMount() {
        userService.checkSession(this.checkAuth);
    }


    render() {

        const {component: Component, path="/",exact=false,strict=false} = this.props;

        if (!this.state.hasAuthed) {
            return null;
        }

        return <Route path={path} exact={exact} strict={strict} render={props => (
            this.state.isAuthed ? (
                <Redirect to={{
                    pathname: '/',
                    state: {from: props.location}
                }}/>
            ) : (
                <Component {...props}/>
            )
        )}/>
    }
}

export default LoginRoute
