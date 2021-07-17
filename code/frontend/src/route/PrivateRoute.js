import React from 'react';
import {Redirect, Route} from 'react-router-dom'
import {checkSession} from "../api/userService";

export default class PrivateRoute extends React.Component {

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
        checkSession(this.checkAuth);
    }


    render() {

        const {component: Component, path = "/", exact = false, strict = false} = this.props;

        if (!this.state.hasAuthed) {
            return null;
        }

        return <Route path={path} exact={exact} strict={strict} render={props => (
            this.state.isAuthed ? (
                <Component {...props}/>
            ) : (
                <Redirect to={{
                    pathname: '/login',
                    state: {from: props.location}
                }}/>
            )
        )}/>
    }
}

