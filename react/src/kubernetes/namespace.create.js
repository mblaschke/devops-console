import React from 'react';
import {CopyToClipboard} from 'react-copy-to-clipboard';

import BaseComponent from '../base';
import NamespaceFormelement from './namespace.formelement';

class NamespaceCreate extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            namespacePreview: "",
            buttonText: "Create namespace",

            environment: {},

            requestRunning: false,
            form: {
                environment: "",
                app: "",
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
        } else if (this.state.form.environment === "" || this.state.form.app === "" || this.state.form.team === "") {
            state = "disabled"
        }

        return state
    }

    previewNamespace() {
        let namespace = "";

        let selectedEnv = this.state.form.environment;
        let envConfig = false;

        this.props.config.kubernetes.environments.forEach((row) => {
            if (row.environment === selectedEnv) {
                envConfig = row;
            }
        });

        if (envConfig && envConfig.template) {
            namespace = envConfig.template;
            namespace = namespace.replace("{env}", selectedEnv);
            namespace = namespace.replace("{user}", this.props.config.user.username);
            namespace = namespace.replace("{team}", this.state.form.team);
            namespace = namespace.replace("{app}", this.state.form.app);
        }

        return namespace.toLowerCase().replace(/_/g, "");
    }

    componentWillMount() {
        this.init();
    }

    init() {
        let state = this.state;

        if (!state.form) {
            state.form = {
                team: "",
                environment: "",
                app: "",
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
        } catch {}

        // select first team if no selection available
        if (!state.form.team || state.form.team === "") {
            if (this.props.config.teams.length > 0) {
                state.form.team = this.props.config.teams[0].name;
            }
        }

        if (!state.form.environment) {
            if (this.props.config.kubernetes.environments.length > 0) {
                state.form.environment = this.props.config.kubernetes.environments[0].environment;
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
                <div className="modal fade" id="createQuestion" tabIndex="-1" role="dialog" aria-labelledby="createQuestion" aria-hidden="true">
                    <div className="modal-dialog" role="document">
                        <div className="modal-content">
                            <div className="modal-header">
                                <h5 className="modal-title" id="exampleModalLabel">Create namespace</h5>
                                <button type="button" className="close" data-bs-dismiss="modal" aria-label="Close">
                                    <span aria-hidden="true">&times;</span>
                                </button>
                            </div>
                                <div className="modal-body">
                                    <div className="row">
                                        <div className="col">
                                            <div className="p-3 mb-2 bg-light text-dark">
                                                <div className="button-copy-box">
                                                    <i>Preview: </i><span id="namespacePreview">{this.previewNamespace()}</span>
                                                    <CopyToClipboard text={this.previewNamespace()}>
                                                        <button className="button-copy" onClick={this.handlePreventEvent.bind(this)}><i className="far fa-copy"></i></button>
                                                    </CopyToClipboard>
                                                </div>
                                            </div>
                                        </div>
                                    </div>

                                    <div className="row">
                                        <div className="col-3">
                                            <label htmlFor="inputNsEnvironment">Environment</label>
                                            <select name="nsEnvironment" id="inputNsEnvironment" className="form-control" required value={this.getValue("form.environment")} onChange={this.setValue.bind(this, "form.environment")}>
                                            {this.props.config.kubernetes.environments.map((row) =>
                                                <option key={row.environment} value={row.environment}>{row.environment} ({row.description})</option>
                                            )}
                                            </select>
                                        </div>
                                        <div>
                                            <label htmlFor="inputNsAreaTeam">Team</label>
                                            <select name="nsAreaTeam" id="inputNsAreaTeam" className="form-control namespace-area-team" value={this.getValue("form.team")} onChange={this.setValue.bind(this, "form.team")}>
                                                {this.props.config.teams.map((row, value) =>
                                                    <option key={row.Id} value={row.name}>{row.name}</option>
                                                )}
                                            </select>
                                        </div>
                                        <div className="col">
                                            <label htmlFor="inputNsApp" className="inputNsApp">Application</label>
                                            <input type="text" name="nsApp" id="inputNsApp" className="form-control" placeholder="Name" required value={this.getValue("form.app")} onChange={this.setValue.bind(this, "form.app")} />
                                        </div>
                                    </div>

                                    <div className="row">
                                        <div className="col">
                                            <input type="text" name="nsDescription" className="form-control" placeholder="Description" value={this.getValue("form.description")} onChange={this.setValue.bind(this, "form.description")} />
                                        </div>
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="inputNsNetpol" className="inputRg">NetworkPolicy</label>
                                        <select id="inputNsNetpol" className="form-control" value={this.getValue("form.networkPolicy")} onChange={this.setValue.bind(this, "form.networkPolicy")}>
                                            {this.props.config.kubernetes.namespace.networkPolicy.map((row) =>
                                                <option key={row.name} value={row.name}>{row.description}</option>
                                            )}
                                        </select>
                                    </div>

                                    {this.kubernetesSettingsConfig().map((setting, value) =>
                                        <NamespaceFormelement setting={setting} value={this.getValue("form.settings." + setting.name)} onchange={this.setValue.bind(this, "form.settings." + setting.name)} />
                                    )}
                                </div>
                                <div className="modal-footer">
                                    <button type="button" className="btn btn-secondary bnt-k8s-namespace-cancel" data-bs-dismiss="modal">Cancel</button>
                                    <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.getButtonState()} onClick={this.createNamespace.bind(this)}>{this.state.buttonText}</button>
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

