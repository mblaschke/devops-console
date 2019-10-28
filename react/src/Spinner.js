import React, { Component } from 'react';

class Spinner extends Component {
    render() {
        return (
            <div className={this.props.active ? 'spinner-wrapper' : 'spinner-wrapper hidden'}>
                <div className="spinner">
                    <div className="rect1"></div>
                    <div className="rect2"></div>
                    <div className="rect3"></div>
                    <div className="rect4"></div>
                    <div className="rect5"></div>
                </div>
            </div>
        )
    }
}

export default Spinner;
