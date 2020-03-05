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
            url: "/api/kubernetes/namespace",
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

        this.props.config.NamespaceEnvironments.forEach((row) => {
            if (row.Environment === selectedEnv) {
                envConfig = row;
            }
        });

        if (envConfig && envConfig.Template) {
            namespace = envConfig.Template;
            namespace = namespace.replace("{env}", selectedEnv);
            namespace = namespace.replace("{user}", this.props.config.User.Username);
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
            this.props.config.Teams.map((row, value) => {
                if (row.Name === lastSelectedTeam) {
                    state.namespace.team = lastSelectedTeam;
                }
            });
        } catch {}

        // select first team if no selection available
        if (!state.namespace.team || state.namespace.team === "") {
            if (this.props.config.Teams.length > 0) {
                state.namespace.team = this.props.config.Teams[0].Name;
            }
        }

        if (!state.namespace.environment) {
            if (this.props.config.NamespaceEnvironments.length > 0) {
                state.namespace.environment = this.props.config.NamespaceEnvironments[0].Environment;
            }
        }

        state.namespace.settings = {};
        this.kubernetesSettingsConfig().map((setting) => {
            if (setting.Default) {
                state.namespace.settings[setting.Name] = setting.Default
            } else {
                state.namespace.settings[setting.Name] = "";
            }
        });

        this.setState(state);
    }


    kubernetesSettingsConfig() {
        let ret = [];

        if (this.props.config.Kubernetes.Namespace.Settings) {
            ret = this.props.config.Kubernetes.Namespace.Settings;
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
                                            {this.props.config.NamespaceEnvironments.map((row) =>
                                                <option key={row.Environment} value={row.Environment}>{row.Environment} ({row.Description})</option>
                                            )}
                                            </select>
                                        </div>
                                        <div>
                                            <label htmlFor="inputNsAreaTeam">Team</label>
                                            <select name="nsAreaTeam" id="inputNsAreaTeam" className="form-control namespace-area-team" value={this.getNamespaceItem("team")} onChange={this.handleNamespaceInputChange.bind(this, "team")}>
                                                {this.props.config.Teams.map((row, value) =>
                                                    <option key={row.Id} value={row.Name}>{row.Name}</option>
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
                                        <select id="inputNsNetpol" className="form-control" value={this.getNamespaceItem("netpol")} onChange={this.handleNamespaceInputChange.bind(this, "netpol")}>
                                            <option value="none">no policy</option>
                                            <option value="deny">block all traffic</option>
                                            <option value="namespace">allow same namespace</option>
                                            <option value="allow">allow all</option>
                                        </select>
                                    </div>

                                    {this.kubernetesSettingsConfig().map((setting, value) =>
                                        <K8sNamespaceFormElement setting={setting} value={this.getNamespaceSettingItem(setting.Name)} onchange={this.handleNamespaceSettingInputChange.bind(this, setting.Name)} />
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

