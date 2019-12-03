import React from 'react';
import $ from 'jquery';

import BaseComponent from './BaseComponent';
import Spinner from './Spinner';
import Breadcrumb from './Breadcrumb';

class AlertmanagerAlerts extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            alerts: [],
            config: {
                User: {
                    Username: '',
                },
                Teams: [],
                Alertmanager: {
                    Instances: []
                },
            },
            instance: "",
            searchValue: "",
        };

        window.App.registerSearch((event) => {
            this.handleSearchChange(event);
        });

        window.App.enableSearch();
    }

    load() {
        if (!this.state.instance || this.state.instance === "") {
            return
        }

        let jqxhr = $.get({
            url: '/api/alertmanager/' + encodeURI(this.state.instance) + '/alerts/'
        }).done((jqxhr) => {
            if (this.state.isStartup) {
                this.setInputFocus();
            }

            this.setState({
                alerts: jqxhr,
                isStartup: false
            });
        });

        this.handleXhr(jqxhr);
    }

    loadConfig() {
        let jqxhr = $.get({
            url: '/api/app/config'
        }).done((jqxhr) => {
            if (jqxhr) {
                if (!jqxhr.Teams) {
                    jqxhr.Teams = [];
                }

                if (!jqxhr.NamespaceEnvironments) {
                    jqxhr.NamespaceEnvironments = [];
                }

                this.setState({
                    config: jqxhr,
                    isStartup: false
                });

                // trigger init
                setTimeout(() => {
                    this.init();
                });
            }
        });

        this.handleXhr(jqxhr);
    }

    componentDidMount() {
        this.loadConfig();
    }

    init() {
        let state = this.state;

        // default team for local storage
        try {
            let lastAlertmangerInstance = "" + localStorage.getItem("alertmanager");
            this.state.config.Alertmanager.Instances.map((row, value) => {
                if (row === lastAlertmangerInstance) {
                    state.instance = lastAlertmangerInstance;
                }
            });
        } catch {}

        // select first team if no selection available
        if (!state.instance || state.instance === "") {
            if (this.state.config.Alertmanager.Instances.length > 0) {
                state.instance = this.state.config.Alertmanager.Instances[0];
            }
        }

        this.setState(state);

        setTimeout(() => {
            this.refresh();
        }, 100);
    }

    refresh() {
        this.load();

        try {
            clearTimeout(this.refreshHandler);
        } catch(e) {}

        this.refreshHandler = setTimeout(() =>{
            this.refresh();
        }, 10000);
    }

    handleInstanceChange(event) {
        var state = this.state;
        state.instance = event.target.type === 'checkbox' ? String(event.target.checked) : String(event.target.value);
        this.setState(state);

        localStorage.setItem("alertmanager", state.instance);

        this.refresh();
    }

    getList() {
        let ret = this.state.alerts;
        return ret;
    }

    render() {
        if (this.state.isStartup) {
            return (
                <div>
                    <Spinner active={this.state.isStartup}/>
                </div>
            )
        }

        let self = this;
        let alerts = this.getList();
        let instances = this.state.config.Alertmanager.Instances ? this.state.config.Alertmanager.Instances : [];

        let labelList = (labels) => {
            let ret = "";


            return ret;
        };

        return (
            <div>
                <Breadcrumb/>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-object-group"></i>
                        Active alerts
                        <div className="toolbox">
                            <select className="form-control" required value={this.state.instance} onChange={this.handleInstanceChange.bind(this)}>
                                {instances.map((row) =>
                                    <option key={row} value={row}>{row}</option>
                                )}
                            </select>
                        </div>
                    </div>
                    <div className="card-body">
                        <table className="table table-hover table-sm">
                            <colgroup>
                                <col width="*"/>
                                <col width="200rem"/>
                                <col width="200rem"/>
                                <col width="200rem"/>
                                <col width="80rem"/>
                            </colgroup>
                            <thead>
                            <tr>
                                <th>Alert</th>
                                <th>Labels</th>
                                <th>Started</th>
                                <th>Updated</th>
                                <th>Status</th>
                                <th></th>
                            </tr>
                            </thead>
                            <tbody>
                            {alerts.map((row) =>
                                <tr>
                                    <td>
                                        <strong>{row.annotations.summary}</strong><br />
                                        {row.annotations.description}
                                    </td>
                                    <td>
                                        {Object.entries(row.labels).map((item) =>
                                            <span>
                                                <span className="badge badge-secondary">{item[0]}: {item[1]}</span>
                                                <br />
                                            </span>
                                        )}
                                    </td>
                                    <td>
                                        {row.startsAt}
                                    </td>
                                    <td>
                                        {row.updatedAt}
                                    </td>
                                    <td>
                                        {(() => {
                                            switch (row.status.state) {
                                                case "active":
                                                    return <span className="badge badge-danger">active</span>
                                                case "suppressed":
                                                    return <span className="badge badge-warning">suppressed</span>
                                            }
                                        })()}

                                    </td>
                                </tr>
                            )}
                            </tbody>
                        </table>
                    </div>
                    <div className="card-footer small text-muted"></div>
                </div>

            </div>
        );
    }
}

export default AlertmanagerAlerts;

