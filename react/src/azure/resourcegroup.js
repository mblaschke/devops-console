import React from 'react';
import BaseComponent from '../base';
import Spinner from '../spinner';
import Breadcrumb from '../breadcrumb';

class Resourcegroup extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartup: true,
            config: this.buildAppConfig(),

            searchValue: "",
            buttonText: "Create Azure ResourceGroup",

            requestRunning: false,
            form: {
                team: "",
                name: "",
                location: "westeurope",
                tag: {}
            }
        };
    }

    init() {
        this.componentWillMount();
    }

    componentWillMount() {
        let state = this.state;

        // default team for local storage
        try {
            let lastSelectedTeam = "" + localStorage.getItem("team");
            this.state.config.teams.map((row, value) => {
                if (row.name === lastSelectedTeam) {
                    state.form.team = lastSelectedTeam;
                }
            });
        } catch {}

        // select first team if no selection available
        if (this.state.form.team === "") {
            if (this.state.config.teams.length > 0) {
                state.form.team = this.state.config.teams[0].name
            }
        }

        state.form.tag = {};
        this.azureResourceGroupTagConfig().map((setting) => {
            if (setting.Default) {
                state.form.tag[setting.Name] = setting.Default;
            }
        });

        this.setState(state);
    }

    componentDidMount() {
        this.loadConfig();
        this.setInputFocus();
    }

    createResourceGroup(e) {
        e.preventDefault();
        e.stopPropagation();

        let oldButtonText = this.state.buttonText;
        this.setState({
            requestRunning: true,
            buttonText: "Saving..."
        });

        this.ajax({
            type: 'POST',
            url: "/_webapi/azure/resourcegroup",
            data: JSON.stringify(this.state.form)
        }).done(() => {
            let state = this.state;
            state.form.name = "";
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
        } else if (this.state.form.name === "" || this.state.form.team === "" || this.state.form.location === "") {
            state = "disabled"
        }

        return state
    }

    handleClickOutside() {
        this.setInputFocus();
    }

    azureResourceGroupTagConfig() {
        let ret = [];

        if (this.state.config.azure.resourceGroup.tags) {
            ret = this.state.config.azure.resourceGroup.tags
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
                                <select name="nsAreaTeam" id="inputNsAreaTeam" className="form-control namespace-area-team" value={this.getValue("form.team")} onChange={this.setValue.bind(this, "form.team")}>
                                    {this.state.config.teams.map((row, value) =>
                                        <option key={row.Id} value={row.name}>{row.name}</option>
                                    )}
                                </select>
                            </div>

                            <div className="form-group">
                                <label htmlFor="inputNsApp" className="inputRg">Azure ResourceGroup</label>
                                <input type="text" name="nsApp" id="inputRg" className="form-control" placeholder="ResourceGroup name" required value={this.getValue("form.name")} onChange={this.setValue.bind(this, "form.name")} />
                            </div>
                            <div className="form-group">
                                <label htmlFor="inputNsApp" className="inputRgLocation">Azure Location</label>
                                <input type="text" name="nsApp" id="inputRgLocation" className="form-control" placeholder="ResourceGroup location" required value={this.getValue("form.location")} onChange={this.setValue.bind(this, "form.location")} />
                            </div>

                            {this.azureResourceGroupTagConfig().map((setting, value) =>
                                <div className="form-group">
                                    <label htmlFor="inputNsApp" className="inputRg">{setting.label}</label>
                                    <input type="text" name={setting.name} id={setting.name} className="form-control" placeholder={setting.plaeholder} value={this.getValue("form.tag." + setting.name)} onChange={this.setValue.bind(this, "form.tag." + setting.name)} />
                                    <small className="form-text text-muted">{setting.description}</small>
                                </div>
                            )}
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

export default Resourcegroup;

