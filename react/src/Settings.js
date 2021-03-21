import React from 'react';
import BaseComponent from './BaseComponent';
import Spinner from "./Spinner";
import Breadcrumb from "./Breadcrumb";

class Settings extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            isStartupConfig: true,
            isStartupSettings: true,
            buttonState: "",

            config: this.buildAppConfig(),

            requestRunning: false,

            settingConfig: {
                User: [],
                Team: []
            },

            user: {},

            team: {}
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
        let jqxhr = this.ajax({
            type: 'GET',
            url: '/_webapi/general/settings'
        }).done((jqxhr) => {
            if (jqxhr) {
                var state = this.state;
                state.isStartupSettings = false;

                if (jqxhr.Configuration) {
                    state.settingConfig = jqxhr.Configuration;
                }

                if (jqxhr.User) {
                    state.user = jqxhr.User;
                }

                if (jqxhr.Team) {
                    state.team = jqxhr.Team;
                }

                this.setState(state);
            }
        });
    }

    handlePersonalInputChange(name, event) {
        var state = this.state.user;
        state[name] = event.target.value;
        this.setState(state);
    }

    handleTeamInputChange(team, name, event) {
        var state = this.state.team;

        if (!state[team]) {
            state[team] = {};
        }

        state[team][name] = event.target.value;
        this.setState(state);
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

        let jqxhr = this.ajax({
            type: 'POST',
            url: "/_webapi/general/settings/user",
            data: JSON.stringify(this.state.user)
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

        let jqxhr = this.ajax({
            type: 'POST',
            url: "/_webapi/general/settings/team/" + encodeURI(team),
            data: JSON.stringify(this.getTeamConfig(team))
        }).always(() => {
            this.setState({
                requestRunning: false,
            });
        });
    }

    getUserConfigItem(name) {
        var ret = "";

        if (this.state.user && this.state.user[name]) {
            ret = this.state.user[name];
        }

        return ret;
    }

    getTeamConfig(team) {
        var ret = {};

        if (this.state.team && this.state.team[team]) {
            ret = this.state.team[team];
        }

        return ret;
    }
    getTeamConfigItem(team, name) {
        var ret = "";

        if (this.state.team && this.state.team[team] && this.state.team[team][name]) {
            ret = this.state.team[team][name];
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
                            {this.state.settingConfig.User.map((setting, value) =>
                                <div className="form-group">
                                    <label htmlFor={"personal-" + setting.Name} className="inputRg">{setting.Label}</label>
                                    <input type="text" name={setting.Name} id={"personal-" + setting.Name} className="form-control" placeholder={setting.Plaeholder} value={this.getUserConfigItem(setting.Name)} onChange={this.handlePersonalInputChange.bind(this, setting.Name)} />
                                </div>
                            )}
                            <div className="toolbox">
                                <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.stateUpdateButton()} onClick={this.updateUserSettings.bind(this)}>Save</button>
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
                                {this.state.settingConfig.Team.map((setting, value) =>
                                    <div className="form-group">
                                        <label htmlFor={"team-" + team.name + "-" + setting.Name} className="inputRg">{setting.Label}</label>
                                        <input type="text" name={setting.Name} id={"team-" + team.name + "-" + setting.Name} className="form-control" placeholder={setting.Plaeholder} value={this.getTeamConfigItem(team.name, setting.Name)} onChange={this.handleTeamInputChange.bind(this, team.name, setting.Name)} />
                                    </div>
                                )}
                                <div className="toolbox">
                                    <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.stateUpdateButton()} onClick={this.updateTeamSettings.bind(this, team.name)}>Save</button>
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
