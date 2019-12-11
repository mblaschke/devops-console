import React from 'react';
import $ from 'jquery';

import BaseComponent from './BaseComponent';
import Spinner from './Spinner';

class GeneralStats extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            stats: {},
            updateDate: false,
            isStartup: true
        };

        setTimeout(() => {
            this.refresh()
        }, 500);

        setInterval(() => {
            this.refresh()
        }, 10000);
    }

    refresh() {
        let jqxhr = this.ajax({
            type: "GET",
            url: '/api/general/stats'
        }).done((jqxhr) => {
            this.setState({
                stats: jqxhr,
                updateDate: new Date(),
                isStartup: false
            });
        });

        this.handleXhr(jqxhr);
    }

    render() {
        if (this.state.isStartup) {
            return (
                <div>
                    <Spinner active={this.state.isStartup}/>
                </div>
            )
        }

        return (
            <div>
                <Spinner active={this.state.isStartup}/>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-info-circle"></i>
                        Application statistics
                    </div>
                    <div className="card-body">
                        {this.state.stats.map((category) =>
                            <table className="table table-hover table-sm">
                                <colgroup>
                                    <col width="200rem" />
                                    <col width="*" />
                                </colgroup>
                                <thead>
                                    <tr>
                                        <th colspan="2">{category.Name}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                {category.Stats.map((stat) =>
                                    <tr>
                                        <td>{stat.Name}</td>
                                        <td>{stat.Value}</td>
                                    </tr>
                                )}
                                </tbody>
                            </table>
                        )}
                    </div>
                    <div className="card-footer small text-muted">Last update: {this.state.updateDate ? this.state.updateDate.toLocaleString() : 'updating...' }</div>
                </div>
            </div>
        );
    }
}

export default GeneralStats;

