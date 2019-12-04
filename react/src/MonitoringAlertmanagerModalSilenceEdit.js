import React from 'react';
import $ from 'jquery';
import _ from 'lodash';

import * as utils from './utils.js';
import BaseComponent from './BaseComponent';

class MonitoringAlertmanagerModalSilenceEdit extends BaseComponent {
    constructor(props) {
        super(props);

        this.state = {
            namespacePreview: "",
            buttonText: "Save",
            buttonState: "",
            reload: true,
            team: "",
            silence: {}
        };
    }

    componentWillMount() {
        this.init();
    }

    init() {
        let state = this.state;

        // default team for local storage
        try {
            let lastSelectedTeam = "" + localStorage.getItem("team");
            this.props.config.Teams.map((row, value) => {
                if (row.Name === lastSelectedTeam) {
                    state.team = lastSelectedTeam;
                }
            });
        } catch {}

        // select first team if no selection available
        if (!state.team || state.team === "") {
            if (this.props.config.Teams.length > 0) {
                state.team = this.props.config.Teams[0].Name;
            }
        }

        this.setState(state);
    }

    save(e) {
        e.preventDefault();
        e.stopPropagation();

        if (this.props.silence.id) {
            this.saveUpdate(e);
        } else {
            this.saveCreate(e);
        }
    }

    saveUpdate(e) {
        let oldButtonText = this.state.buttonText;
        this.setState({
            buttonState: "disabled",
            buttonText: "Saving...",
        });

        let jqxhr = $.ajax({
            type: 'PUT',
            contentType: 'application/json; charset=UTF-8',
            url: "/api/alertmanager/" + encodeURI(this.props.instance) + "/silence/" + encodeURI(this.props.silence.id),
            data: JSON.stringify({
                team: this.state.team,
                silence: this.state.silence
            })
        }).done((jqxhr) => {
            this.setState({
                silence: false,
                reload: true,
            });

            if (this.props.callback) {
                this.props.callback()
            }
        }).always(() => {
            this.setState({
                buttonState: "",
                buttonText: oldButtonText
            });
        });

        this.handleXhr(jqxhr);
    }

    saveCreate(e) {
        let oldButtonText = this.state.buttonText;
        this.setState({
            buttonState: "disabled",
            buttonText: "Creating...",
        });

        let jqxhr = $.ajax({
            type: 'POST',
            contentType: 'application/json; charset=UTF-8',
            url: "/api/alertmanager/" + encodeURI(this.props.instance) + "/silence/",
            data: JSON.stringify({
                team: this.state.team,
                silence: this.state.silence
            })
        }).done((jqxhr) => {
            this.setState({
                silence: false,
                reload: true,
            });

            if (this.props.callback) {
                this.props.callback()
            }
        }).always(() => {
            this.setState({
                buttonState: "",
                buttonText: oldButtonText
            });
        });

        this.handleXhr(jqxhr);
    }

    cancelEdit() {
        this.setState({
            silence: false,
            reload: true,
        });
    }

    componentWillReceiveProps(nextProps) {
        if (!nextProps.silence) {
            // invalid silence
            return
        }

        if (!this.props.silence || this.props.silence.id !== nextProps.silence.id) {
            // make copy
            let silence = JSON.parse(JSON.stringify(nextProps.silence));

            let team = this.state.team;

            try {
                if (silence.matchers) {
                    let matcherTeam = false;
                    let matchersFiltered = [];
                    silence.matchers.map((matcher) => {
                        if (matcher.name === "team") {
                            matcherTeam = matcher.value;
                        } else {
                            matchersFiltered.push(matcher);
                        }
                    });
                    silence.matchers = matchersFiltered;

                    if (matcherTeam) {
                        this.props.config.Teams.map((row, value) => {
                            if (row.Name === matcherTeam) {
                                team = matcherTeam;
                            }
                        });
                    }
                }

            } catch {}

            // set to state
            this.setState({
                silence: silence,
                team: team,
                reload: false
            });
        }
    }

    deleteMatcher(num) {
        var state = this.state;
        state.silence.matchers.splice(num, 1 );
        this.setState(state);
    }

    addMatcher() {
        var state = this.state;
        state.silence.matchers.push({
            name: "",
            value: "",
            regexp: false,
        });
        this.setState(state);
    }

    htmlIdMatcher(key) {
        return "form-element-matcher" + key;
    }

    render() {
        if (!this.state.silence) {
            return (
                <div>
                    <form method="post">
                        <div className="modal fade" id="editQuestion" tabIndex="-1" role="dialog" aria-labelledby="editQuestion" aria-hidden="true">
                        </div>
                    </form>
                </div>
            )
        }

        let matchers = this.state.silence.matchers ? this.state.silence.matchers : [];
        // filter team
        matchers = matchers.filter((row) => {
            if (row.name === "team") {
                return false;
            }

            return true;
        });

        return (
            <div>
                <form method="post">
                <div className="modal fade" id="editQuestion" tabIndex="-1" role="dialog" aria-labelledby="editQuestion" aria-hidden="true">
                    <div className="modal-dialog" role="document">
                        <div className="modal-content">
                            <div className="modal-header">
                                <h5 className="modal-title" id="exampleModalLabel">Silence</h5>
                                <button type="button" className="close" data-dismiss="modal" aria-label="Close">
                                    <span aria-hidden="true">&times;</span>
                                </button>
                            </div>
                                <div className="modal-body">

                                    <div className="form-group">
                                        <label htmlFor="inputTeam">Team</label>
                                        <select name="inputTeam" id="inputTeam" className="form-control" value={this.getValue("team")} onChange={this.setValue.bind(this, "team")}>
                                            {this.props.config.Teams.map((row, value) =>
                                                <option key={row.Id} value={row.Name}>{row.Name}</option>
                                            )}
                                        </select>
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="inputNsDescription" className="inputRg">Description</label>
                                        <textarea className="form-control" value={this.getValue("silence.comment")} onChange={this.setValue.bind(this, "silence.comment")}  />
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="inputNsDescription" className="inputRg">Starts at</label>
                                        <input className="form-control" value={this.getValue("silence.startsAt")} onChange={this.setValue.bind(this, "silence.startsAt")}  />
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="inputNsDescription" className="inputRg">Ends at</label>
                                        <input className="form-control" value={this.getValue("silence.endsAt")} onChange={this.setValue.bind(this, "silence.endsAt")}  />
                                    </div>

                                    <table className="table table-sm table-borderless">
                                        <colgroup>
                                            <col widht="*" />
                                            <col widht="*" />
                                            <col widht="50rem" />
                                            <col widht="20rem" />
                                        </colgroup>
                                        <thead>
                                            <tr>
                                                <th>Match</th>
                                                <th>Filter</th>
                                                <th></th>
                                                <th>
                                                    <button type="button" className="btn btn-secondary btn-sm" onClick={this.addMatcher.bind(this)}>
                                                        <i className="fas fa-plus"></i>
                                                    </button>
                                                </th>
                                            </tr>
                                        </thead>

                                        <tbody>
                                        {matchers.map((item,key) =>
                                            <tr>
                                                <td>
                                                    <div className="form-group">
                                                        <input className="form-control" value={this.getValue("silence.matchers[" + key + "].name")} onChange={this.setValue.bind(this, "silence.matchers[" + key + "].name")}  />
                                                    </div>
                                                </td>

                                                <td>
                                                    <div className="form-group">
                                                        <input className="form-control" value={this.getValue("silence.matchers[" + key + "].value")} onChange={this.setValue.bind(this, "silence.matchers[" + key + "].value")}  />
                                                    </div>
                                                </td>

                                                <td>
                                                    <div className="form-group form-check">
                                                        <input type="checkbox" id={this.htmlIdMatcher(key)}
                                                               className="form-check-input"
                                                               checked={this.getValueCheckbox("silence.matchers[" + key + "].isRegex")}
                                                               onChange={this.setValueCheckbox.bind(this, "silence.matchers[" + key + "].isRegex")}/>
                                                        <label className="form-check-label"
                                                               htmlFor={this.htmlIdMatcher(key)}>
                                                            Regexp
                                                        </label>
                                                    </div>
                                                </td>

                                                <td>
                                                    <button type="button" className="btn btn-secondary btn-sm" onClick={this.deleteMatcher.bind(this, key)}>
                                                        <i className="fas fa-trash-alt"></i>
                                                    </button>
                                                </td>
                                            </tr>
                                        )}
                                        </tbody>
                                    </table>

                                    <div className="modal-footer">
                                        <button type="button" className="btn btn-secondary bnt-k8s-namespace-cancel" data-dismiss="modal">Cancel</button>
                                        <button type="submit" className="btn btn-primary bnt-k8s-namespace-create" disabled={this.state.buttonState} onClick={this.save.bind(this)}>{this.state.buttonText}</button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </form>
            </div>
        );
    }
}

export default MonitoringAlertmanagerModalSilenceEdit;

