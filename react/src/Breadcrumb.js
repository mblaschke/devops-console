import React from 'react';
import BaseComponent from './BaseComponent';

class Breadcrumb extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            breadcrumbs: []
        };

        this.refresh();
    }

    refresh() {
        let state = this.state;
        let pathPartsRaw = window.location.pathname.split("/", -1);
        state.breadcrumbs = [];

        pathPartsRaw.map((row) => {
            if (row !== "") {
                state.breadcrumbs.push({
                    title: row,
                    className: "breadcrumb-item"
                })
            }
        });

        if (state.breadcrumbs.length >= 1) {
            state.breadcrumbs[state.breadcrumbs.length-1].className += " active"
        }

        this.setState(state);
    }

    render() {
        return (
            <ol className="breadcrumb">
                {this.state.breadcrumbs.map((row) =>
                    <li className={row.className}>{row.title}</li>
                )}
            </ol>
        )
    }
}

export default Breadcrumb;

