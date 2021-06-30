import React from 'react';
import {withRouter} from "react-router-dom";
import {LuckySheet} from "../components/LuckySheet";

class SheetView extends React.Component{

    constructor(props) {
        super(props);
        this.state={
            fid:0
        }
    }

    componentDidMount(){
        const query = this.props.location.search;
        const arr = query.split('&');
        const fid = parseInt(arr[0].substr(4));
        this.setState({fid:fid})
    }

    render(){
        return(
            <LuckySheet fid={this.state.fid}/>
        );
    }
}

export default withRouter(SheetView);
