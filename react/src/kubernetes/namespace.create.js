import React from 'react';
import {CopyToClipboard} from 'react-copy-to-clipboard';

import BaseComponent from '../base';
import NamespaceFormelement from './namespace.formelement';

class NamespaceCreate extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            buttonText: "Create namespace",

            requestRunning: false,
            form: {
                name: "",
                team: "",
                description: "",
                settings: {}
            }
        };
    }

    createNamespace(e) {
        e.preventDefault();
        e.stopPropagation();

        let oldButtonText = this.state.buttonText;
        this.setState({
            requestRunning: true,
            buttonText: "Saving...",
        });

        this.ajax({
            type: 'POST',
            url: "/_webapi/kubernetes/namespace",
            data: JSON.stringify(this.state.form)
        }).done(() => {
            this.setState({
                form: false
            });
            this.init();

            if (this.props.callback) {
                this.props.callback()
            }
        }).always(() => {
            this.setState({
                requestRunning: false,
                buttonText: oldButtonText
            });
        });
    }

    getButtonState() {
        let state = "";

        if (this.state.requestRunning) {
            state = "disabled";
        } else if (this.state.form.name === "" || this.state.form.team === "") {
            state = "disabled"
        }

        return state
    }

    componentWillMount() {
        this.init();
    }

    init() {
        let state = this.state;

        if (!state.form) {
            state.form = {
                name: "",
                team: "",
                description: ""
            };
        }

        // default team for local storage
        try {
            let lastSelectedTeam = "" + localStorage.getItem("team");
            this.props.config.teams.map((row, value) => {
                if (row.name === lastSelectedTeam) {
                    state.form.team = lastSelectedTeam;
                }
            });
        } catch (e) {}

        // select first team if no selection available
        if (!state.form.team || state.form.team === "") {
            if (this.props.config.teams.length > 0) {
                state.form.team = this.props.config.teams[0].name;
            }
        }

        state.form.settings = {};
        this.kubernetesSettingsConfig().map((setting) => {
            if (setting.default) {
                state.form.settings[setting.name] = setting.default
            } else {
                state.form.settings[setting.name] = "";
            }
        });

        this.setState(state);
    }


    kubernetesSettingsConfig() {
        let ret = [];

        if (this.props.config.kubernetes.namespace.settings) {
            ret = this.props.config.kubernetes.namespace.settings;
        }

        return ret;
    }

    render() {
        return (
            <div>
                <form method="post">
                    <div className="modal fade" id="createQuestion" tabIndex="-1" role="dialog"
                         aria-labelledby="createQuestion" aria-hidden="true">
                        <div className="modal-dialog" role="document">
                            <div className="modal-content">
                                <div className="modal-header">
                                    <h5 className="modal-title" id="exampleModalLabel">Create namespace</h5>
                                    <button type="button" className="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                                </div>
                                <div className="modal-body">
                                    <div className="row">
                                        <div className="col">
                                            <label htmlFor="inputNsName" className="inputRg">Namespace</label>
                                            <input type="text" id="inputNsName" name="nsName" className="form-control"
                                                   placeholder="Namespace" value={this.getValue("form.name")}
                                                   onChange={this.setValue.bind(this, "form.name")}/>
                                        </div>
                                    </div>

                                    <div className="row">
                                        <div className="col">
                                            <label htmlFor="inputNsTeamr">Team</label>
                                            <select name="nsTeam" id="inputNsTeam"
                                                    className="form-control namespace-team"
                                                    value={this.getValue("form.team")}
                                                    onChange={this.setValue.bind(this, "form.team")}>
                                                {this.props.config.teams.map((row, value) =>
                                                    <option key={row.Id} value={row.name}>{row.name}</option>
                                                )}
                                            </select>
                                        </div>
                                    </div>

                                    <div className="row">
                                        <div className="col">
                                            <label htmlFor="inputNsDescription">Description</label>
                                            <input type="text" id="inputNsDescription" name="nsDescription" className="form-control"
                                                   placeholder="Description" value={this.getValue("form.description")}
                                                   onChange={this.setValue.bind(this, "form.description")}/>
                                        </div>
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="inputNsNetpol" className="inputRg">NetworkPolicy</label>
                                        <select id="inputNsNetpol" className="form-control"
                                                value={this.getValue("form.networkPolicy")}
                                                onChange={this.setValue.bind(this, "form.networkPolicy")}>
                                            {this.props.config.kubernetes.namespace.networkPolicy.map((row) =>
                                                <option key={row.name} value={row.name}>{row.description}</option>
                                            )}
                                        </select>
                                    </div>

                                    {this.kubernetesSettingsConfig().map((setting, value) =>
                                        <NamespaceFormelement setting={setting}
                                                              value={this.getValue("form.settings." + setting.name)}
                                                              onchange={this.setValue.bind(this, "form.settings." + setting.name)}/>
                                    )}
                                </div>
                                <div className="modal-footer">
                                    <button type="button" className="btn btn-secondary bnt-k8s-namespace-cancel"
                                            data-bs-dismiss="modal">Cancel
                                    </button>
                                    <button type="submit" className="btn btn-primary bnt-k8s-namespace-create"
                                            disabled={this.getButtonState()}
                                            onClick={this.createNamespace.bind(this)}>{this.state.buttonText}</button>
                                </div>
                            </div>
                        </div>
                    </div>
                </form>
            </div>
        );
    }
}

export default NamespaceCreate;

