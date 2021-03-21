import React from 'react';
import {CopyToClipboard} from 'react-copy-to-clipboard';

import BaseComponent from './BaseComponent';
import K8sNamespaceFormElement from './K8sNamespaceFormElement';

class K8sNamespaceModalCreate extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            namespacePreview: "",
            buttonText: "Create namespace",
            buttonState: "disabled",

            environment: {},

            namespace: {
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
            buttonState: "disabled",
            buttonText: "Saving...",
        });

        let jqxhr = this.ajax({
            type: 'POST',
            url: "/_webapi/kubernetes/namespace",
            data: JSON.stringify(this.state.namespace)
        }).done((jqxhr) => {
            this.setState({
                namespace: false
            });
            this.init();

            if (this.props.callback) {
                this.props.callback()
            }
        }).always(() => {
            this.setState({
                buttonState: "",
                buttonText: oldButtonText
            });
        });
    }

    handleNamespaceInputChange(name, event) {
        var state = this.state;
        state.namespace[name] = event.target.type === 'checkbox' ? String(event.target.checked) : String(event.target.value);
        this.setState(state);

        this.handleButtonState();

        if (name === "team") {
            try {
                localStorage.setItem("team", state.namespace[name]);
            } catch {}
        }
    }


    handleNamespaceSettingInputChange(name, event) {
        var state = this.state;

        if (!state.namespace.settings) {
            state.namespace.settings = {}
        }

        state.namespace["settings"][name] = event.target.type === 'checkbox' ? String(event.target.checked) : String(event.target.value);
        this.setState(state);
    }

    getNamespaceItem(name) {
        var ret = "";

        if (this.state.namespace && this.state.namespace[name]) {
            ret = this.state.namespace[name];
        }

        return ret;
    }

    getNamespaceSettingItem(name) {
        var ret = "";

        if (this.state.namespace.settings && this.state.namespace.settings[name]) {
            ret = this.state.namespace.settings[name];
        }

        return ret;
    }

    handleButtonState(event) {
        let buttonState = "disabled";

        if (this.state.namespace.environment !== "" && this.state.namespace.app !== "" && this.state.namespace.team !== "") {
            buttonState = ""
        }

        this.setState({
            buttonState: buttonState,
        });
    }

    handleNsDescriptionChange(event) {
        let state = this.state;
        state.namespace.description = event.target.type === 'checkbox' ? String(event.target.checked) : String(event.target.value);
        this.setState(state);
    }

    previewNamespace() {
        let namespace = "";

        let selectedEnv = this.state.namespace.environment;
        let envConfig = false;

        this.props.config.kubernetes.environments.forEach((row) => {
            if (row.environment === selectedEnv) {
                envConfig = row;
            }
        });

        if (envConfig && envConfig.Template) {
            namespace = envConfig.Template;
            namespace = namespace.replace("{env}", selectedEnv);
            namespace = namespace.replace("{user}", this.props.config.user.username);
            namespace = namespace.replace("{team}", this.state.namespace.team);
            namespace = namespace.replace("{app}", this.state.namespace.app);
        }

        return namespace.toLowerCase().replace(/_/g, "");
    }

    componentWillMount() {
        this.init();
    }

    init() {
        let state = this.state;

        if (!state.namespace) {
            state.namespace = {
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
                    state.namespace.team = lastSelectedTeam;
                }
            });
        } catch {}

        // select first team if no selection available
        if (!state.namespace.team || state.namespace.team === "") {
            if (this.props.config.teams.length > 0) {
                state.namespace.team = this.props.config.teams[0].name;
            }
        }

        if (!state.namespace.environment) {
            if (this.props.config.kubernetes.environments.length > 0) {
                state.namespace.environment = this.props.config.kubernetes.environments[0].environment;
            }
        }

        state.namespace.settings = {};
        this.kubernetesSettingsConfig().map((setting) => {
            if (setting.default) {
                state.namespace.settings[setting.name] = setting.default
            } else {
                state.namespace.settings[setting.name] = "";
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
                                <button type="button" className="close" data-dismiss="modal" aria-label="Close">
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
                                            <select name="nsEnvironment" id="inputNsEnvironment" className="form-control" required value={this.getNamespaceItem("environment")} onChange={this.handleNamespaceInputChange.bind(this, "environment")}>
                                            {this.props.config.kubernetes.environments.map((row) =>
                                                <option key={row.environment} value={row.environment}>{row.environment} ({row.description})</option>
                                            )}
                                            </select>
                                        </div>
                                        <div>
                                            <label htmlFor="inputNsAreaTeam">Team</label>
                                            <select name="nsAreaTeam" id="inputNsAreaTeam" className="form-control namespace-area-team" value={this.getNamespaceItem("team")} onChange={this.handleNamespaceInputChange.bind(this, "team")}>
                                                {this.props.config.teams.map((row, value) =>
                                                    <option key={row.Id} value={row.name}>{row.name}</option>
                                                )}
                                            </select>
                                        </div>
                                        <div className="col">
                                            <label htmlFor="inputNsApp" className="inputNsApp">Application</label>
                                            <input type="text" name="nsApp" id="inputNsApp" className="form-control" placeholder="Name" required value={this.getNamespaceItem("app")} onChange={this.handleNamespaceInputChange.bind(this, "app")} />
                                        </div>
                                    </div>

                                    <div className="row">
                                        <div className="col">
                                            <input type="text" name="nsDescription" className="form-control" placeholder="Description" value={this.getNamespaceItem("description")} onChange={this.handleNamespaceInputChange.bind(this, "description")} />
                                        </div>
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="inputNsNetpol" className="inputRg">NetworkPolicy</label>
                                        <select id="inputNsNetpol" className="form-control" value={this.getNamespaceItem("networkPolicy")} onChange={this.handleNamespaceInputChange.bind(this, "networkPolicy")}>
                                            {this.props.config.kubernetes.namespace.networkPolicy.map((row) =>
                                                <option key={row.name} value={row.name}>{row.description}</option>
                                            )}
                                        </select>
                                    </div>

                                    {this.kubernetesSettingsConfig().map((setting, value) =>
                                        <K8sNamespaceFormElement setting={setting} value={this.getNamespaceSettingItem(setting.name)} onchange={this.handleNamespaceSettingInputChange.bind(this, setting.name)} />
                                    )}
                                </div>
                                <div className="modal-footer">
                                    <button type="button" className="btn btn-secondary bnt-k8s-namespace-cancel" data-dismiss="modal">Cancel</button>
                                    <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.state.buttonState} onClick={this.createNamespace.bind(this)}>{this.state.buttonText}</button>
                                </div>
                            </div>
                        </div>
                    </div>
                </form>
            </div>
        );
    }
}

export default K8sNamespaceModalCreate;

