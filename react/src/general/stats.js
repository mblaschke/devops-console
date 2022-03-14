import React from 'react';
import Base from '../base';
import Spinner from '../spinner';

class Stats extends Base {
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
        this.ajax({
            type: "GET",
            url: '/_webapi/general/stats'
        }).done((jqxhr) => {
            this.setState({
                stats: jqxhr,
                updateDate: new Date(),
                isStartup: false
            });
        });
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
                                    <col width="200rem"/>
                                    <col width="*"/>
                                </colgroup>
                                <thead>
                                <tr>
                                    <th colSpan="2">{category.name}</th>
                                </tr>
                                </thead>
                                <tbody>
                                {category.stats.map((stat) =>
                                    <tr>
                                        <td>{stat.name}</td>
                                        <td>{stat.value}</td>
                                    </tr>
                                )}
                                </tbody>
                            </table>
                        )}
                    </div>
                    <div className="card-footer small text-muted">Last
                        update: {this.state.updateDate ? this.state.updateDate.toLocaleString() : 'updating...'}</div>
                </div>
            </div>
        );
    }
}

export default Stats;

