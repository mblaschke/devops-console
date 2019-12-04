import React from 'react';
import $ from 'jquery';
import moment from 'moment';

import BaseComponent from './BaseComponent';
import Spinner from './Spinner';
import Breadcrumb from './Breadcrumb';
import MonitoringAlertmanagerModalSilenceDelete from "./MonitoringAlertmanagerModalSilenceDelete";
import MonitoringAlertmanagerModalSilenceEdit from "./MonitoringAlertmanagerModalSilenceEdit";
import * as utils from "./utils";

class MonitoringAlertmanager extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            config: {
                User: {
                    Username: '',
                },
                Teams: [],
                Alertmanager: {
                    Instances: []
                },
            },
            alerts: [],
            silences: [],
            filter: {
                silence: {
                    expired: false
                }
            },
            instance: "",
            selectedSilence: false,
            searchValue: "",
        };

        window.App.registerSearch((event) => {
            this.handleSearchChange(event);
        });

        window.App.enableSearch();

        $(document).on('show.bs.modal', ".modal", this.disableRefresh.bind(this));
        $(document).on('hide.bs.modal', ".modal", this.refresh.bind(this));
    }

    loadAlerts() {
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

    loadSilences() {
        if (!this.state.instance || this.state.instance === "") {
            return
        }

        let jqxhr = $.get({
            url: '/api/alertmanager/' + encodeURI(this.state.instance) + '/silences/'
        }).done((jqxhr) => {
            if (this.state.isStartup) {
                this.setInputFocus();
            }

            this.setState({
                silences: jqxhr,
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

    disableRefresh() {
        try {
            clearTimeout(this.refreshHandler);
        } catch(e) {}
    }

    refresh() {
        this.loadAlerts();
        this.loadSilences();

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

    silenceDelete(row) {
        this.setState({
            selectedSilence: row
        });

        setTimeout(() => {
            $("#deleteQuestion").modal('show');
        }, 200);
    }

    handleSilenceDelete() {
        $("#deleteQuestion").modal('hide');
        this.setState({
            selectedSilence: false,
        });
    }

    silenceEdit(row) {
        this.setState({
            selectedSilence: row
        });

        setTimeout(() => {
            $("#editQuestion").modal('show');
        }, 200);
    }

    silenceNew() {
        this.setState({
            selectedSilence: {
                id: false,
                startsAt: "" + new Date().toISOString(),
                endsAt: "" + new Date( new Date().getTime() + 1*3600*1000).toISOString(),
                comment: "",
                matchers: [{
                    name: "",
                    value: "",
                    isRegex: false,
                }],
            }
        });

        setTimeout(() => {
            $("#editQuestion").modal('show');
        }, 200);
    }

    silenceNewFromAlert(alert) {
        let matchers = [];

        Object.entries(alert.labels).map((item) => {
            matchers.push({
                name: item[0],
                value: item[1],
                isRegex: false,
            });
        });

        this.setState({
            selectedSilence: {
                id: false,
                startsAt: "" + new Date().toISOString(),
                endsAt: "" + new Date( new Date().getTime() + 1*3600*1000).toISOString(),
                comment: "Silence alert: " + alert.annotations.summary + "\n" + alert.annotations.description,
                matchers: matchers,
            }
        });

        setTimeout(() => {
            $("#editQuestion").modal('show');
        }, 200);
    }

    handleSilenceEdit() {
        $("#editQuestion").modal('hide');
        this.setState({
            selectedSilence: false,
        });
    }

    getAlertList() {
        let ret = this.state.alerts ? this.state.alerts : [];

        if (this.state.searchValue !== "") {
            let term = this.state.searchValue.replace(/[.?*+^$[\]\\(){}|-]/g, "\\$&");
            let re = new RegExp(term, "i");

            ret = ret.filter((row) => {
                if (row.annotations.summary.search(re) !== -1) return true;
                if (row.annotations.description.search(re) !== -1) return true;
                if (row.startsAt.search(re) !== -1) return true;
                if (row.updatedAt.search(re) !== -1) return true;
                if (row.status.state.search(re) !== -1) return true;

                if (row.labels) {
                    for(var i in row.labels) {
                        if (row.labels[i].search(re) !== -1) return true;
                    }
                }


                return false;
            });
        }

        return ret;
    }

    getSilenceList() {
        let ret = this.state.silences ? this.state.silences : [];

        if (this.state.searchValue !== "") {
            let term = this.state.searchValue.replace(/[.?*+^$[\]\\(){}|-]/g, "\\$&");
            let re = new RegExp(term, "i");

            ret = ret.filter((row) => {
                if (row.comment.search(re) !== -1) return true;
                if (row.createdBy.search(re) !== -1) return true;
                if (row.startsAt.search(re) !== -1) return true;
                if (row.endsAt.search(re) !== -1) return true;
                if (row.status.state.search(re) !== -1) return true;

                if (row.matchers) {
                    for(var i in row.matchers) {
                        if (row.matchers[i].value.search(re) !== -1) return true;
                    }
                }

                return false;
            });
        }

        if (!this.state.filter.silence.expired) {
            ret = ret.filter((row) => {
                if (row.status.state !== "expired") return true;
                return false;
            });
        }

        return ret;
    }

    transformTime(time) {
        return (
            <span>
                {moment(time, moment.ISO_8601).fromNow()}<br/>
                <small>{this.highlight(time)}</small>
            </span>
        )
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
        let alerts = this.getAlertList();
        let silcenes = this.getSilenceList();
        let instances = this.state.config.Alertmanager.Instances ? this.state.config.Alertmanager.Instances : [];

        return (
            <div>
                <Breadcrumb/>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-bell"></i>
                        Alerts
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
                                <col width="200rem"/>
                                <col width="80rem"/>
                            </colgroup>
                            <thead>
                            <tr>
                                <th>Alert</th>
                                <th>Labels</th>
                                <th>Started</th>
                                <th>Last update</th>
                                <th>Status</th>
                                <th></th>
                            </tr>
                            </thead>
                            <tbody>
                            {alerts.map((row) =>
                                <tr>
                                    <td>
                                        <strong>{this.highlight(row.annotations.summary)}</strong><br />
                                        <small>{this.highlight(row.annotations.description)}</small>
                                    </td>
                                    <td>
                                        {Object.entries(row.labels).map((item) =>
                                            <span>
                                                <span className="badge badge-secondary">{item[0]}: {this.highlight(item[1])}</span>
                                                <br />
                                            </span>
                                        )}
                                    </td>
                                    <td>{this.transformTime(row.startsAt)}</td>
                                    <td>{this.transformTime(row.updatedAt)}</td>
                                    <td>
                                        {(() => {
                                            switch (row.status.state) {
                                                case "active":
                                                    return <span className="badge badge-danger blinking">{this.highlight(row.status.state)}</span>
                                                case "suppressed":
                                                    return <span className="badge badge-warning">{this.highlight(row.status.state)}</span>
                                                default:
                                                    return <span className="badge badge-secondary">{this.highlight(row.status.state)}</span>
                                            }
                                        })()}
                                    </td>
                                    <td className="toolbox">
                                        {(() => {
                                            switch (row.status.state) {
                                                case "active":
                                                    return <button type="button" className="btn btn-secondary" onClick={this.silenceNewFromAlert.bind(this, row)}>
                                                                Silence
                                                            </button>
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

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-bell-slash"></i>
                        Silences
                        <div className="toolbox">
                            <div className="form-group row">
                                <div className="col">
                                    <input type="checkbox" className="form-check-input" id="silenceFilterExpired"
                                           checked={this.getValueCheckbox("filter.silence.expired")}
                                           onChange={this.setValueCheckbox.bind(this, "filter.silence.expired")}/>
                                    <label className="form-check-label" htmlFor="silenceFilterExpired">Expired</label>
                                </div>

                                <div className="col">
                                    <select className="form-control" required value={this.state.instance} onChange={this.handleInstanceChange.bind(this)}>
                                        {instances.map((row) =>
                                            <option key={row} value={row}>{row}</option>
                                        )}
                                    </select>
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className="card-body">
                        <table className="table table-hover table-sm">
                            <colgroup>
                                <col width="*"/>
                                <col width="200rem"/>
                                <col width="200rem"/>
                                <col width="200rem"/>
                                <col width="100rem"/>
                                <col width="100rem"/>
                                <col width="80rem"/>
                            </colgroup>
                            <thead>
                            <tr>
                                <th>Alert</th>
                                <th>Matchers</th>
                                <th>From</th>
                                <th>Until</th>
                                <th></th>
                                <th className="toolbox">
                                    <button type="button" className="btn btn-secondary" onClick={this.silenceNew.bind(this)}>
                                        <i className="fas fa-plus"></i>
                                    </button>
                                </th>
                            </tr>
                            </thead>
                            <tbody>
                            {silcenes.map((row) =>
                                <tr>
                                    <td>
                                        {this.highlight(row.comment)}
                                        <br/>
                                        <i><small>created: {this.highlight(row.createdBy)}</small></i>
                                    </td>
                                    <td>
                                        {row.matchers.map((item) =>
                                            <span>
                                                <span className="badge badge-secondary">{item.name}: {this.highlight(item.value)}</span>
                                                <br />
                                            </span>
                                        )}
                                    </td>
                                    <td>{this.transformTime(row.startsAt)}</td>
                                    <td>{this.transformTime(row.endsAt)}</td>
                                    <td>
                                        {(() => {
                                            switch (row.status.state) {
                                                case "active":
                                                    return <span className="badge badge-success blinking">{this.highlight(row.status.state)}</span>
                                                case "expired":
                                                    return <span className="badge badge-warning">{this.highlight(row.status.state)}</span>
                                                default:
                                                    return <span className="badge badge-secondary">{this.highlight(row.status.state)}</span>
                                            }
                                        })()}
                                    </td>
                                    <td className="toolbox">
                                        <div className="btn-group" role="group">
                                            <button id="btnGroupDrop1" type="button"
                                                    className="btn btn-secondary dropdown-toggle"
                                                    data-toggle="dropdown" aria-haspopup="true"
                                                    aria-expanded="false">
                                                Action
                                            </button>
                                            <div className="dropdown-menu" aria-labelledby="btnGroupDrop1">
                                                <a className="dropdown-item" onClick={self.silenceEdit.bind(self, row)}>Edit</a>
                                                <a className="dropdown-item" onClick={self.silenceDelete.bind(self, row)}>Delete</a>
                                            </div>
                                        </div>
                                    </td>
                                </tr>
                            )}
                            </tbody>
                        </table>
                    </div>
                    <div className="card-footer small text-muted"></div>
                </div>

                <MonitoringAlertmanagerModalSilenceDelete instance={this.state.instance} silence={this.state.selectedSilence} config={this.state.config} callback={this.handleSilenceDelete.bind(this)} />
                <MonitoringAlertmanagerModalSilenceEdit instance={this.state.instance} silence={this.state.selectedSilence} config={this.state.config} callback={this.handleSilenceEdit.bind(this)} />
            </div>
        );
    }
}

export default MonitoringAlertmanager;

