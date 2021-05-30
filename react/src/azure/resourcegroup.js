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

            resourceId: false,

            requestRunning: false,
            form: {
                team: "",
                name: "",
                location: "westeurope",
                tags: {}
            }
        };
    }

    init() {
        this.componentWillMount();
    }

    componentWillMount() {
        this.initTeamSelection('form.team');

        let tags = {};
        this.azureResourceGroupTagConfig().map((setting) => {
            if (setting.default) {
                tags[setting.name] = setting.default;
            }
        });

        this.setValue('form.tags', tags)
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
        }).done((jqxhr) => {
            let state = this.state;

            if (jqxhr && jqxhr.resourceId) {
                state.resourceId = "" + jqxhr.resourceId;
            }

            state.form.name = "";
            console.log(state)
            this.setState(state);
        }).always(() => {
            this.setState({
                requestRunning: false,
                buttonText: oldButtonText
            });
        });
    }


    openResourceGroup(e) {
        e.preventDefault();
        e.stopPropagation();
        if (this.state.resourceId) {
            let tenantId = encodeURI(this.state.config.azure.tenantId)
            let resourceId = encodeURI(this.state.resourceId)
            window.open(`https://portal.azure.com/#@${tenantId}/resource/${resourceId}`, '_blank');
        }
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

    stateOpenButton() {
        let state = "disabled";

        if (!this.state.requestRunning && this.state.resourceId) {
            state = "";
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

                            {this.azureResourceGroupTagConfig().map((setting) =>
                                <div className="form-group">
                                    <label htmlFor="inputNsApp" className="inputRg">{setting.label}</label>
                                    <input type="text" name={setting.name} id={setting.name} className="form-control" placeholder={setting.placeholder} value={this.getValue("form.tags." + setting.name)} onChange={this.setValue.bind(this, "form.tags." + setting.name)} />
                                    <small className="form-text text-muted">{setting.description}</small>
                                </div>
                            )}
                            <div className="toolbox">
                                <button type="submit" className="btn btn-primary bnt-azure-resourcegroup-create" disabled={this.stateCreateButton()} onClick={this.createResourceGroup.bind(this)}>{this.state.buttonText}</button>
                                <button type="button" className="btn btn-secondary bnt-azure-resourcegroup-open" disabled={this.stateOpenButton()} onClick={this.openResourceGroup.bind(this)}>Goto ResourceGroup</button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        );
    }
}

export default Resourcegroup;

