import React from 'react';
import $ from 'jquery';

import BaseComponent from './BaseComponent';
import Spinner from './Spinner';
import Breadcrumb from './Breadcrumb';

class AzureResourceGroups extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            searchValue: "",
            buttonText: "Create Azure ResourceGroup",
            requestRunning: false,

            resourceGroup: {
                team: "",
                name: "",
                location: "westeurope",
                personal: false,
                tag: {}
            },

            config: {
                User: {
                    Username: '',
                },
                Teams: [],
                NamespaceEnvironments: [],
                Quota: {},
                Azure: {
                    ResourceGroup: {
                        Tags: []
                    }
                }
            },
            isStartup: true
        };

        setInterval(() => {
            this.refresh()
        }, 10000);
    }

    loadConfig() {
        this.ajax({
            type: "GET",
            url: '/api/app/config'
        }).done((jqxhr) => {
            if (jqxhr) {
                if (!jqxhr.Teams) {
                    jqxhr.Teams = [];
                }

                if (!jqxhr.NamespaceEnvironments) {
                    jqxhr.NamespaceEnvironments = [];
                }

                if (this.state.isStartup) {
                    this.setInputFocus();
                }

                this.setState({
                    config: jqxhr,
                    isStartup: false
                });

                this.componentWillMount();
            }
        });
    }

    componentWillMount() {
        let state = this.state;

        // default team for local storage
        try {
            let lastSelectedTeam = "" + localStorage.getItem("team");
            this.state.config.Teams.map((row, value) => {
                if (row.Name === lastSelectedTeam) {
                    state.resourceGroup.team = lastSelectedTeam;
                }
            });
        } catch {}

        // select first team if no selection available
        if (this.state.resourceGroup.team === "") {
            if (this.state.config.Teams.length > 0) {
                state.resourceGroup.team = this.state.config.Teams[0].Name
            }
        }

        state.resourceGroup.tag = {};
        this.azureResourceGroupTagConfig().map((setting) => {
            if (setting.Default) {
                state.resourceGroup.tag[setting.Name] = setting.Default;
            }
        });

        this.setState(state);
    }

    componentDidMount() {
        this.loadConfig();
        this.setInputFocus();
    }

    refresh() {
    }

    createResourceGroup(e) {
        e.preventDefault();
        e.stopPropagation();

        let oldButtonText = this.state.buttonText;
        this.setState({
            requestRunning: true,
            buttonText: "Saving..."
        });

        let jqxhr = this.ajax({
            type: 'POST',
            url: "/api/azure/resourcegroup",
            data: JSON.stringify(this.state.resourceGroup)
        }).done((jqxhr) => {
            let state = this.state;
            state.resourceGroup.name = "";
            this.setState(state);
        }).always(() => {
            this.setState({
                requestRunning: false,
                buttonText: oldButtonText
            });
        });
    }

    stateCreateButton() {
        let state = "";

        if (this.state.requestRunning) {
            state = "disabled";
        } else {
            if (this.state.azResourceGroup === "" || this.state.azTeam === "" || this.state.azResourceGroupLocation === "") {
                state = "disabled"
            }
        }

        return state
    }

    handleResourceGroupInputChange(name, event) {
        let state = this.state;
        state.resourceGroup[name] = event.target.value;
        this.setState(state);

        if (name === "team") {
            try {
                localStorage.setItem("team", event.target.value);
            } catch {}
        }
    }


    handleResourceGroupTagInputChange(name, event) {
        let state = this.state;
        state.resourceGroup["tag"][name] = event.target.value;
        this.setState(state);
    }

    getResourceGroupItem(name) {
        let ret = "";

        if (this.state.resourceGroup && this.state.resourceGroup[name]) {
            ret = this.state.resourceGroup[name];
        }

        return ret;
    }

    handleResourceGroupCheckboxChange(name, event) {
        let state = this.state;
        state.resourceGroup[name] = event.target.checked;
        this.setState(state);
    }

    getResourceGroupItemBool(name) {
        return (this.state.resourceGroup && this.state.resourceGroup[name])
    }

    getResourceGroupTagItem(name) {
        var ret = "";

        if (this.state.resourceGroup.tag && this.state.resourceGroup.tag[name]) {
            ret = this.state.resourceGroup.tag[name];
        }

        return ret;
    }


    handleClickOutside() {
        this.setInputFocus();
    }

    azureResourceGroupTagConfig() {
        let ret = [];

        if (this.state.config.Azure.ResourceGroup.Tags) {
            ret = this.state.config.Azure.ResourceGroup.Tags
        }

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

        return (
            <div>
                <Spinner active={this.state.isStartup}/>

                <Breadcrumb/>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-box"></i>
                        Create Azure ResourceGroup
                    </div>
                    <div className="card-body">
                        <form method="post">
                            <div className="form-group">
                                <label htmlFor="inputNsAreaTeam">Team</label>
                                <select name="nsAreaTeam" id="inputNsAreaTeam" className="form-control namespace-area-team" value={this.getResourceGroupItem("team")} onChange={this.handleResourceGroupInputChange.bind(this, "team")}>
                                    {this.state.config.Teams.map((row, value) =>
                                        <option key={row.Id} value={row.Name}>{row.Name}</option>
                                    )}
                                </select>
                            </div>

                            <div className="form-group">
                                <label htmlFor="inputNsApp" className="inputRg">Azure ResourceGroup</label>
                                <input type="text" name="nsApp" id="inputRg" className="form-control" placeholder="ResourceGroup name" required value={this.getResourceGroupItem("name")} onChange={this.handleResourceGroupInputChange.bind(this, "name")} />
                            </div>
                            <div className="form-group">
                                <label htmlFor="inputNsApp" className="inputRgLocation">Azure Location</label>
                                <input type="text" name="nsApp" id="inputRgLocation" className="form-control" placeholder="ResourceGroup location" required value={this.getResourceGroupItem("location")} onChange={this.handleResourceGroupInputChange.bind(this, "location")} />
                            </div>

                            {this.azureResourceGroupTagConfig().map((setting, value) =>
                                <div className="form-group">
                                    <label htmlFor="inputNsApp" className="inputRg">{setting.Label}</label>
                                    <input type="text" name={setting.Name} id={setting.Name} className="form-control" placeholder={setting.Plaeholder} value={this.getResourceGroupTagItem(setting.Name)} onChange={this.handleResourceGroupTagInputChange.bind(this, setting.Name)} />
                                    <small className="form-text text-muted">{setting.Description}</small>
                                </div>
                            )}

                            <div className="form-group">
                                <div className="form-check">
                                    <input type="checkbox" className="form-check-input" id="az-resourcegroup-personal" checked={this.getResourceGroupItemBool("personal")} onChange={this.handleResourceGroupCheckboxChange.bind(this, "personal")} />
                                    <label className="form-check-label" htmlFor="az-resourcegroup-personal">Personal ResourceGroup (only read access to team)</label>
                                </div>
                            </div>
                            <div className="toolbox">
                                <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.stateCreateButton()} onClick={this.createResourceGroup.bind(this)}>{this.state.buttonText}</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        );
    }
}

export default AzureResourceGroups;

