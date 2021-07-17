import React from 'react';
import {Redirect, Route, Router, Switch} from 'react-router-dom';
import PrivateRoute from './PrivateRoute'
import LoginRoute from './LoginRoute'
import HomeView from "../view/HomeView";
import LoginView from '../view/LoginView'
import SheetView from "../view/SheetView";
import RegisterView from "../view/RegisterView";
import {history} from "./history";

class BasicRoute extends React.Component {

    constructor(props) {
        super(props);

        history.listen((location, action) => {
            console.log(location, action);
        });
    }

    render() {
        return (
            <Router history={history}>
                <Switch>
                    <PrivateRoute exact path="/" component={HomeView}/>
                    <LoginRoute exact path="/login" component={LoginView}/>
                    <Route exact path="/register" component={RegisterView}/>
                    <PrivateRoute exact path="/sheet" component={SheetView}/>
                    <Redirect from="/*" to="/"/>
                </Switch>
            </Router>
        )
    }


}

export default BasicRoute;
