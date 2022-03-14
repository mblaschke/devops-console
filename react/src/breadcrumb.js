import React from 'react';
import Base from './base';

class Breadcrumb extends Base {
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

        pathPartsRaw.forEach((row) => {
            if (row !== "") {
                state.breadcrumbs.push({
                    title: row,
                    className: "breadcrumb-item"
                })
            }
        });

        if (state.breadcrumbs.length >= 1) {
            state.breadcrumbs[state.breadcrumbs.length - 1].className += " active"
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

