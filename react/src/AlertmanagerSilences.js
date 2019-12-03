import React from 'react';
import $ from 'jquery';

import BaseComponent from './BaseComponent';
import Spinner from './Spinner';
import Breadcrumb from './Breadcrumb';
import AlertmanagerSilenceModalDelete from "./AlertmanagerSilenceModalDelete";

class AlertmanagerSilences extends BaseComponent {
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
            selectedSilence: false,
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
            url: '/api/alertmanager/' + encodeURI(this.state.instance) + '/silences/'
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

        return (
            <div>
                <Breadcrumb/>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-object-group"></i>
                        Active silences
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
                                <th>Creator</th>
                                <th>Started</th>
                                <th>Ends at</th>
                                <th></th>
                            </tr>
                            </thead>
                            <tbody>
                            {alerts.map((row) =>
                                <tr>
                                    <td>{row.comment}</td>
                                    <td>{row.createdBy}</td>
                                    <td>{row.startsAt}</td>
                                    <td>{row.endsAt}</td>
                                    <td>
                                        <div className="btn-group" role="group">
                                            <button id="btnGroupDrop1" type="button"
                                                    className="btn btn-secondary dropdown-toggle"
                                                    data-toggle="dropdown" aria-haspopup="true"
                                                    aria-expanded="false">
                                                Action
                                            </button>
                                            <div className="dropdown-menu" aria-labelledby="btnGroupDrop1">
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


                <AlertmanagerSilenceModalDelete instance={this.state.instance} silence={this.state.selectedSilence} callback={this.handleSilenceDelete.bind(this)} />

            </div>
        );
    }
}

export default AlertmanagerSilences;

