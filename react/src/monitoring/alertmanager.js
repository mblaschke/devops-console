import React from 'react';
import $ from 'jquery';
import moment from 'moment';

import BaseComponent from '../base';
import Spinner from '../spinner';
import Breadcrumb from '../breadcrumb';

import AlertmanagerSilencedelete from "./alertmanager.silencedelete";
import AlertmanagerSilenceedit from "./alertmanager.silenceedit";

class Alertmanager extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            config: this.buildAppConfig(),
            alerts: [],
            silences: [],
            loadingAlerts: true,
            loadingSilences: true,
            filter: {
                silence: {
                    active: true,
                    expired: false,
                    expanded: []
                },
                alert: {
                    active: true,
                    suppressed: true,
                    expanded: []
                }
            },
            instance: "",
            team: "*",
            selectedSilence: false,
            searchValue: "",
        };

        window.App.registerSearch((event) => {
            this.handleSearchChange(event);
        });

        window.App.enableSearch();

        $(document).on('show.bs.modal', ".modal", this.disableRefresh.bind(this));
        $(document).on('hide.bs.modal', ".modal", () => {
            this.setState({
                loadingSilences: true,
                loadingAlerts: true
            });

            // wait until alertmanagers are synced
            setTimeout(this.refresh.bind(this,true), 1250)
        });
    }

    loadAlerts(showSpinner) {
        if (!this.state.instance || this.state.instance === "") {
            return
        }

        this.setState({
            loadingAlerts: showSpinner,
        });


        let jqxhr = this.ajax({
            type: 'GET',
            url: '/_webapi/alertmanager/' + encodeURI(this.state.instance) + '/alerts'
        }).done((jqxhr) => {
            if (this.state.isStartup) {
                this.setInputFocus();
            }

            let alerts = Array.isArray(jqxhr) ? jqxhr : [];

            this.setState({
                alerts: alerts,
                isStartup: false,
                loadingAlerts: false,
            });
        });
    }

    loadSilences(showSpinner) {
        if (!this.state.instance || this.state.instance === "") {
            return
        }

        this.setState({
            loadingSilences: showSpinner,
        });

        let jqxhr = this.ajax({
            type: 'GET',
            url: '/_webapi/alertmanager/' + encodeURI(this.state.instance) + '/silences'
        }).done((jqxhr) => {
            if (this.state.isStartup) {
                this.setInputFocus();
            }

            let silences = Array.isArray(jqxhr) ? jqxhr : [];

            this.setState({
                silences: silences,
                isStartup: false,
                loadingSilences: false,
            });
        });
    }

    componentDidMount() {
        this.loadConfig();
    }

    init() {
        this.initTeam();

        let state = this.state;

        // default team for local storage
        try {
            let lastAlertmangerInstance = "" + localStorage.getItem("alertmanager");
            this.state.config.alertmanager.instances.map((row, value) => {
                if (row === lastAlertmangerInstance) {
                    state.instance = lastAlertmangerInstance;
                }
            });
        } catch {}

        // select first team if no selection available
        if (!state.instance || state.instance === "") {
            if (this.state.config.alertmanager.instances.length > 0) {
                state.instance = this.state.config.alertmanager.instances[0];
            }
        }

        this.setState(state);

        setTimeout(() => {
            this.refresh(true);
        }, 250);
    }

    disableRefresh() {
        try {
            clearTimeout(this.refreshHandler);
        } catch(e) {}
    }

    refresh(showSpinner) {
        this.loadAlerts(showSpinner);
        this.loadSilences(showSpinner);

        try {
            clearTimeout(this.refreshHandler);
        } catch(e) {}

        this.refreshHandler = setTimeout(() =>{
            this.refresh(false);
        }, 15000);
    }

    handleInstanceChange(event) {
        var state = this.state;
        state.instance = event.target.type === 'checkbox' ? String(event.target.checked) : String(event.target.value);
        this.setState(state);

        localStorage.setItem("alertmanager", state.instance);

        this.refresh(false);
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
        let matchers = [];

        // add team matcher (force team selection)
        if (this.state.team !== "*") {
            matchers.push({
                name: "team",
                value: this.state.team,
                isRegex: false,
            });
        }

        // add empty matcher
        matchers.push({
            name: "",
            value: "",
            isRegex: false,
        });

        this.setState({
            selectedSilence: {
                __id__: new Date().toISOString(),
                id: false,
                startsAt: "" + new Date().toISOString(),
                endsAt: "" + new Date( new Date().getTime() + 1*3600*1000).toISOString(),
                comment: "",
                team: this.state.team,
                matchers: matchers,
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
                __id__: new Date().toISOString(),
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

    filterAlerts(alerts) {
        let ret = Array.isArray(alerts) ? alerts : [];

        // filter by quickserach
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

        // filter by team
        if (this.state.team !== "*") {
            ret = ret.filter((row) => {
                if (row.labels && row.labels["team"]) {
                    if (row.labels["team"] === this.state.team) {
                        return true;
                    }
                }
                return false;
            });
        }

        // filter by state:active
        if (!this.state.filter.alert.active) {
            ret = ret.filter((row) => {
                if (row.status.state !== "active") return true;
                return false;
            });
        }

        // filter by state:suppressed
        if (!this.state.filter.alert.suppressed) {
            ret = ret.filter((row) => {
                if (row.status.state !== "suppressed") return true;
                return false;
            });
        }

        return ret;
    }

    getSilenceList() {
        let ret = Array.isArray(this.state.silences) ? this.state.silences : [];

        // filter by quickserach
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

        // filter by team
        if (this.state.team !== "*") {
            ret = ret.filter((row) => {
                if (row.matchers) {
                    for (var i in row.matchers) {
                        if (row.matchers[i].name === "team") {
                            if (row.matchers[i].value === this.state.team) {
                                return true;
                            }
                        }
                    }
                }
                return false;
            });
        }

        // filter by state:active
        if (!this.state.filter.silence.active) {
            ret = ret.filter((row) => {
                if (row.status.state !== "active") return true;
                return false;
            });
        }

        // filter by state:expired
        if (!this.state.filter.silence.expired) {
            ret = ret.filter((row) => {
                if (row.status.state !== "expired") return true;
                return false;
            });
        }

        return ret;
    }

    buildFooter(totalRows, visibleRows) {
        let statsList = {};
        let footerLine = "";


        if (totalRows && Array.isArray(totalRows)) {
            // collect
            totalRows.map((row) => {
                if (!statsList[row.status.state]) {
                    statsList[row.status.state] = {
                        "total": 0,
                        "visible": 0,
                    };
                }
                statsList[row.status.state]["total"]++;
            });
        }

        if (visibleRows && Array.isArray(visibleRows)) {
            // collect
            visibleRows.map((row) => {
                if (!statsList[row.status.state]) {
                    statsList[row.status.state] = {
                        "total": 0,
                        "visible": 0,
                    };
                }
                statsList[row.status.state]["visible"]++;
            });
        }

        if (statsList) {
            // to text
            let footerElements = [];
            for (var i in statsList) {
                footerElements.push(`${i}: ${statsList[i]["visible"]} of ${statsList[i]["total"]}`)
            }

            footerLine = footerElements.join(", ");
        }

        return (
            <span>{footerLine}</span>
        )
    }

    buildGroupHeader(rows) {
        let statsList = {};
        let footerLine = "";

        if (rows && Array.isArray(rows)) {
            // collect
            rows.map((row) => {
                if (!statsList[row.status.state]) {
                    statsList[row.status.state] = {
                        "total": 0
                    };
                }
                statsList[row.status.state]["total"]++;
            });
        }
        if (statsList) {
            // to text
            let footerElements = [];
            for (var i in statsList) {
                footerElements.push(`${i}: ${statsList[i]["total"]}`)
            }

            footerLine = footerElements.join(", ");
        }

        return (
            <span>{footerLine}</span>
        )
    }


    transformTime(time) {
        return (
            <span>
                {moment(time, moment.ISO_8601).fromNow()}<br/>
                <small>{this.highlight(time)}</small>
            </span>
        )
    }

    renderMatch(matcher) {
        if (matcher.isRegexp) {
            return (
                <span className="badge badge-secondary">{matcher.name}=~{this.highlight(matcher.value)}</span>
            )
        } else {
            return (
               <span className="badge badge-secondary">{matcher.name}={this.highlight(matcher.value)}</span>
            )
        }
    }

    render() {
        if (this.state.isStartup) {
            return (
                <div>
                    <Spinner active={this.state.isStartup}/>
                </div>
            )
        }

        let alerts = this.state.alerts;
        let alertsVisible = this.filterAlerts(alerts);
        let silences = this.getSilenceList();

        return (
            <div>
                <Breadcrumb/>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-bell"></i>
                        Alerts
                        <div className="toolbox">
                            <div className="form-group row">
                                <div className="col-sm-4">
                                    <div className="dropdown">
                                        <button className="btn btn-secondary dropdown-toggle" type="button"
                                                id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true"
                                                aria-expanded="false">
                                            Filter
                                        </button>

                                        <div className="dropdown-menu" aria-labelledby="dropdownMenuButton">
                                            <form className="px-4 py-3">
                                                <label>Status</label>
                                                <div className="form-check">
                                                    <input type="checkbox" className="form-check-input" id="alertFilterActive" checked={this.getValueCheckbox("filter.alert.active")}
                                                           onChange={this.setValueCheckbox.bind(this, "filter.alert.active")} />
                                                    <label className="form-check-label" htmlFor="alertFilterActive">
                                                        Active
                                                    </label>
                                                </div>

                                                <div className="form-check">
                                                    <input type="checkbox" className="form-check-input" id="alertFilterSuppressed" checked={this.getValueCheckbox("filter.alert.suppressed")}
                                                           onChange={this.setValueCheckbox.bind(this, "filter.alert.suppressed")} />
                                                    <label className="form-check-label" htmlFor="alertFilterSuppressed">
                                                        Suppressed
                                                    </label>
                                                </div>
                                            </form>
                                        </div>
                                    </div>
                                </div>

                                <div className="col-sm-4">
                                    {this.renderTeamSelector()}
                                </div>

                                <div className="col-sm-4">
                                    {this.renderInstanceSelector()}
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className="card-body card-body-table scrollable spinner-area">
                        {this.renderAlerts(alertsVisible)}
                    </div>
                    <div className="card-footer small text-muted">{this.buildFooter(alertsVisible, alerts)}</div>
                </div>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-bell-slash"></i>
                        Silences
                        <div className="toolbox">
                            <div className="form-group row">
                                <div className="col-sm-4">
                                    <div className="dropdown">
                                        <button className="btn btn-secondary dropdown-toggle" type="button"
                                                id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true"
                                                aria-expanded="false">
                                            Filter
                                        </button>

                                        <div className="dropdown-menu" aria-labelledby="dropdownMenuButton">
                                            <form className="px-4 py-3">
                                                <label>Status</label>
                                                <div className="form-check">
                                                    <input type="checkbox" className="form-check-input" id="silenceFilterActive" checked={this.getValueCheckbox("filter.silence.active")}
                                                           onChange={this.setValueCheckbox.bind(this, "filter.silence.active")} />
                                                    <label className="form-check-label" htmlFor="silenceFilterActive">
                                                        Active
                                                    </label>
                                                </div>

                                                <div className="form-check">
                                                    <input type="checkbox" className="form-check-input" id="silenceFilteExpiredr" checked={this.getValueCheckbox("filter.silence.expired")}
                                                           onChange={this.setValueCheckbox.bind(this, "filter.silence.expired")} />
                                                    <label className="form-check-label" htmlFor="silenceFilteExpiredr">
                                                        Expired
                                                    </label>
                                                </div>
                                            </form>
                                        </div>
                                    </div>
                                </div>

                                <div className="col-sm-4">
                                    {this.renderTeamSelector()}
                                </div>

                                <div className="col-sm-4">
                                    {this.renderInstanceSelector()}
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className="card-body card-body-table scrollable spinner-area">
                        {this.renderSilences(silences)}
                    </div>
                    <div className="card-footer small text-muted">{this.buildFooter(this.state.silences, silences)}</div>
                </div>

                <AlertmanagerSilencedelete instance={this.state.instance} silence={this.state.selectedSilence} config={this.state.config} callback={this.handleSilenceDelete.bind(this)} />
                <AlertmanagerSilenceedit instance={this.state.instance} silence={this.state.selectedSilence} config={this.state.config} callback={this.handleSilenceEdit.bind(this)} />
            </div>
        );
    }


    renderInstanceSelector() {
        let instances = this.state.config.alertmanager.instances ? this.state.config.alertmanager.instances : [];

        return (
            <select className="form-control" required value={this.state.instance} onChange={this.handleInstanceChange.bind(this)}>
                {instances.map((row) =>
                    <option key={row} value={row}>{row}</option>
                )}
            </select>
        )
    }

    alertTriggerCollapse(alertName, event) {
        let state = this.state;

        let collapseIndex = state.filter.alert.expanded.indexOf(alertName);

        if (collapseIndex !== -1 ) {
            // existing -> removing
            state.filter.alert.expanded.splice(collapseIndex, 1);
        } else {
            // not existing -> adding
            state.filter.alert.expanded.push(alertName);
        }
        this.setState(state);
        this.handlePreventEvent(event);
        return false;
    }

    buildGroupedAlerts(alerts) {
        let ret = {
            list: {},
            length: 0
        };

        let ungrouped = "Ungrouped alerts";

        alerts.map((row) => {
            if (row.labels && row.labels.alertname) {
                if (!ret.list[row.labels.alertname]) {
                    ret.list[row.labels.alertname] = [];
                }
                ret.list[row.labels.alertname].push(row);
            } else {
                if (!ret.list[ungrouped]) {
                    ret.list[ungrouped] = [];
                }

                ret.list[ungrouped].push(row);
            }
            ret.length++;
        });

        return ret;
    }

    renderAlerts(alerts) {
        let groupedAlerts = this.buildGroupedList(alerts, (row) => {
            let alertname = false;
            if (row.labels && row.labels.alertname) {
                alertname = row.labels.alertname;
            }
            return alertname;
        }, "Ungrouped silences");

        let groupedAlertList = groupedAlerts.list;

        let htmlTableRows = [];

        Object.keys(groupedAlertList).map((alertName, alertIndex) => {
            let isVisible = false;
            let alertGroupIconClassName = "far fa-caret-square-up";

            if (this.state.filter.alert.expanded.indexOf(alertName) !== -1) {
                isVisible = true;
                alertGroupIconClassName = "far fa-caret-square-down";
            }

            let alertList = groupedAlertList[alertName];

            htmlTableRows.push(
                <tr className="alertmanager-alertname-group">
                    <th colSpan="5">
                        <a href="#" className="group-filter" onClick={this.alertTriggerCollapse.bind(this, alertName)}>
                            <i className={alertGroupIconClassName}></i>&nbsp;
                            {alertName} <span className="group-stats">{this.buildGroupHeader(alertList)}</span>
                        </a>
                    </th>
                </tr>
            );

            if (isVisible) {
                alertList = this.sortDataset(alertList);

                alertList.map((row) => {
                    htmlTableRows.push(
                        <tr className="alertmanager-alertname-item">
                            <td className="detail">
                                <strong>{this.highlight(row.annotations.summary)}</strong><br />
                                <small>{this.highlight(row.annotations.description)}</small>

                                <ul className="alertmanager-label">
                                    {Object.entries(row.labels).map((item) =>
                                        <li>
                                            <span className="badge badge-secondary">{item[0]}={this.highlight(item[1])}</span>
                                        </li>
                                    )}
                                </ul>
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
                    )
                });
            }
        });


        return (
            <div>
                <Spinner active={this.state.loadingAlerts}/>

                <div className="table-responsive">
                    <table className="table table-hover table-sm table-fixed">
                        <colgroup>
                            <col width="*"/>
                            <col width="200rem"/>
                            <col width="200rem"/>
                            <col width="100rem"/>
                            <col width="100rem"/>
                        </colgroup>
                        <thead>
                        <tr>
                            <th>{this.sortBy("annotations.summary", "Alert", (a) => {return a.annotations.summary})}</th>
                            <th>{this.sortBy("startsAt", "Started")}</th>
                            <th>{this.sortBy("updatedAt", "Last update")}</th>
                            <th>{this.sortBy("status.state", "Status", (a) => {return a.status.state})}</th>
                            <th></th>
                        </tr>
                        </thead>
                        <tbody>
                            {htmlTableRows.length === 0 &&
                                <tr>
                                    <td colspan="5" className="not-found">No alerts found.</td>
                                </tr>
                            }
                            {htmlTableRows}
                        </tbody>
                    </table>
                </div>
            </div>
        )
    }


    buildGroupedList(list, groupCallback, ungroupedName) {
        let ret = {
            list: {},
            length: 0
        };

        if (!ungroupedName) {
            ungroupedName = "- Ungrouped -";
        }

        list.map((row) => {
            let groupName = groupCallback(row);

            if (groupName) {
                if (!ret.list[groupName]) {
                    ret.list[groupName] = [];
                }
                ret.list[groupName].push(row);
            } else {
                if (!ret.list[ungroupedName]) {
                    ret.list[ungroupedName] = [];
                }

                ret.list[ungroupedName].push(row);
            }
            ret.length++;
        });

        return ret;
    }

    silenceTriggerCollapse(alertName, event) {
        let state = this.state;

        let collapseIndex = state.filter.silence.expanded.indexOf(alertName);

        if (collapseIndex !== -1 ) {
            // existing -> removing
            state.filter.silence.expanded.splice(collapseIndex, 1);
        } else {
            // not existing -> adding
            state.filter.silence.expanded.push(alertName);
        }
        this.setState(state);
        this.handlePreventEvent(event);
        return false;
    }

    renderSilences(silences) {
        let groupedSilences = this.buildGroupedList(silences, (row) => {
            let alertname = false;
            row.matchers.map((item) => {
                if (item.name === "alertname") {
                    alertname = item.value;
                }
            });

            return alertname;
        }, "Ungrouped silences");

        let groupedSilenceList = groupedSilences.list;

        let htmlTableRows = [];

        Object.keys(groupedSilenceList).map((alertName, silenceIndex) => {
            let isVisible = false;
            let alertGroupIconClassName = "far fa-caret-square-up";

            if (this.state.filter.silence.expanded.indexOf(alertName) !== -1) {
                isVisible = true;
                alertGroupIconClassName = "far fa-caret-square-down";
            }

            let silenceList = groupedSilenceList[alertName];

            htmlTableRows.push(
                <tr className="alertmanager-alertname-group">
                    <th colSpan="5">
                        <a href="#" className="group-filter" onClick={this.silenceTriggerCollapse.bind(this, alertName)}>
                            <i className={alertGroupIconClassName}></i>&nbsp;
                            {alertName} <span className="group-stats">{this.buildGroupHeader(silenceList)}</span>
                        </a>
                    </th>
                </tr>
            );

            if (isVisible) {
                silenceList.map((row) => {
                    htmlTableRows.push(
                        <tr>
                            <td>
                                {this.highlight(row.comment)}
                                <br/>
                                <i><small>created: {this.highlight(row.createdBy)}</small></i>

                                <ul className="alertmanager-matcher">
                                    {row.matchers.map((item) =>
                                        <li>{this.renderMatch(item)}</li>
                                    )}
                                </ul>
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
                                    <button id={"btnGroupDrop-" + silences.id} type="button"
                                            className="btn btn-secondary dropdown-toggle"
                                            data-toggle="dropdown" aria-haspopup="true"
                                            aria-expanded="false">
                                        Action
                                    </button>
                                    {(() => {
                                        switch (row.status.state) {
                                            case "expired":
                                                return <div className="dropdown-menu" aria-labelledby={"btnGroupDrop-" + silences.id}>
                                                    <a className="dropdown-item" onClick={this.silenceEdit.bind(this, row)}>Edit</a>
                                                </div>
                                            default:
                                                return <div className="dropdown-menu" aria-labelledby={"btnGroupDrop-" + silences.id}>
                                                    <a className="dropdown-item" onClick={this.silenceEdit.bind(this, row)}>Edit</a>
                                                    <a className="dropdown-item" onClick={this.silenceDelete.bind(this, row)}>Delete</a>
                                                </div>
                                        }
                                    })()}
                                </div>
                            </td>
                        </tr>
                    )
                });
            }
        });


        return (
            <div>
                <Spinner active={this.state.loadingSilences}/>

                <div className="table-responsive">
                    <table className="table table-hover table-sm table-fixed">
                        <colgroup>
                            <col width="*"/>
                            <col width="200rem"/>
                            <col width="200rem"/>
                            <col width="80rem"/>
                            <col width="80rem"/>
                        </colgroup>
                        <thead>
                        <tr>
                            <th>Silence</th>
                            <th>Started</th>
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
                            {htmlTableRows.length === 0 &&
                            <tr>
                                <td colspan="5" className="not-found">No alerts found.</td>
                            </tr>
                            }
                            {htmlTableRows}
                        </tbody>
                    </table>
                </div>
            </div>
        )
    }
}

export default Alertmanager;
