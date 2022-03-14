import React from 'react';
import BaseComponent from '../base';
import Spinner from "../spinner";
import Breadcrumb from "../breadcrumb";

class Settings extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartupConfig: true,
            isStartupSettings: true,

            config: this.buildAppConfig(),

            requestRunning: false,

            settings: {
                user: [],
                team: []
            },

            formUser: {},
            formTeam: {}
        };
    }

    componentDidMount() {
        this.loadConfig();
        this.loadSettings();
    }

    init() {
        this.setState({
            isStartupConfig: false
        });
    }

    loadSettings() {
        this.ajax({
            type: 'GET',
            url: '/_webapi/general/settings'
        }).done((jqxhr) => {
            if (jqxhr) {
                var state = this.state;
                state.isStartupSettings = false;

                if (jqxhr.configuration) {
                    state.settings = jqxhr.configuration;
                }

                if (jqxhr.user) {
                    state.formUser = jqxhr.user;
                }

                if (jqxhr.team) {
                    state.formTeam = jqxhr.team;
                }

                this.setState(state);
            }
        });
    }

    stateUpdateButton() {
        let state = "";

        if (this.state.requestRunning) {
            state = "disabled";
        }

        return state
    }

    updateUserSettings(e) {
        e.preventDefault();
        e.stopPropagation();

        this.setState({
            requestRunning: true,
        });

        this.ajax({
            type: 'POST',
            url: "/_webapi/general/settings/user",
            data: JSON.stringify(this.state.formUser)
        }).always(() => {
            this.setState({
                requestRunning: false,
            });
        });
    }

    updateTeamSettings(team, e) {
        e.preventDefault();
        e.stopPropagation();


        this.setState({
            requestRunning: true
        });

        this.ajax({
            type: 'POST',
            url: "/_webapi/general/settings/team/" + encodeURI(team),
            data: JSON.stringify(this.getTeamConfig(team))
        }).always(() => {
            this.setState({
                requestRunning: false,
            });
        });
    }

    getTeamConfig(team) {
        var ret = {};

        if (this.state.formTeam && this.state.formTeam[team]) {
            ret = this.state.formTeam[team];
        }

        return ret;
    }

    isStartup() {
        return this.state.isStartupConfig || this.state.isStartupSettings
    }

    render() {
        if (this.state.isStartupConfig || this.state.isStartupSettings) {
            return (
                <div>
                    <Spinner active={this.isStartup()}/>
                </div>
            );
        }

        return (
            <div>
                <Spinner active={this.isStartup()}/>

                <Breadcrumb/>

                <div className="card mb-3">
                    <div className="card-header">
                        <i className="fas fa-user-cog"></i>
                        Personal settings
                    </div>
                    <div className="card-body">
                        <form method="post">
                            {this.state.settings.user.map((setting, value) =>
                                <div className="form-group">
                                    <label htmlFor={"personal-" + setting.name}
                                           className="inputRg">{setting.label}</label>
                                    <input type="text" name={setting.name} id={"personal-" + setting.name}
                                           className="form-control" placeholder={setting.Plaeholder}
                                           value={this.getValue(["formUser", setting.name])}
                                           onChange={this.setValue.bind(this, ["formUser", setting.name])}/>
                                </div>
                            )}
                            <div className="toolbox">
                                <button type="submit" className="btn btn-primary bnt-k8s-namespace-create"
                                        disabled={this.stateUpdateButton()}
                                        onClick={this.updateUserSettings.bind(this)}>Save
                                </button>
                            </div>
                        </form>
                    </div>
                </div>

                {this.state.config.teams.map((team, value) =>
                    <div className="card mb-3">
                        <div className="card-header">
                            <i className="fas fa-users-cog"></i>
                            Team {team.name} settings
                        </div>
                        <div className="card-body">
                            <form method="post">
                                {this.state.settings.team.map((setting, value) =>
                                    <div className="form-group">
                                        <label htmlFor={"team-" + team.name + "-" + setting.name}
                                               className="inputRg">{setting.label}</label>
                                        <input type="text" name={setting.name}
                                               id={"team-" + team.name + "-" + setting.name} className="form-control"
                                               placeholder={setting.placeholder}
                                               value={this.getValue(["formTeam", team.name, setting.name])}
                                               onChange={this.setValue.bind(this, ["formTeam", team.name, setting.name])}/>
                                    </div>
                                )}
                                <div className="toolbox">
                                    <button type="submit" className="btn btn-primary bnt-k8s-namespace-create"
                                            disabled={this.stateUpdateButton()}
                                            onClick={this.updateTeamSettings.bind(this, team.name)}>Save
                                    </button>
                                </div>
                            </form>
                        </div>
                    </div>
                )}

            </div>
        );
    }
}

export default Settings;
