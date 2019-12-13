import React from 'react';
import _ from 'lodash';
import moment from 'moment';
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

        let jqxhr = this.ajax({
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
    }

    saveCreate(e) {
        let oldButtonText = this.state.buttonText;
        this.setState({
            buttonState: "disabled",
            buttonText: "Creating...",
        });

        let jqxhr = this.ajax({
            type: 'POST',
            contentType: 'application/json; charset=UTF-8',
            url: "/api/alertmanager/" + encodeURI(this.props.instance) + "/silence",
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

        // avoid react updating the form data (internal id)
        if (nextProps.silence.__id__ && this.props.silence && this.props.silence.__id__) {
            if (nextProps.silence.__id__ === this.props.silence.__id__) {
                return;
            }
        }

        // avoid react updating the form data item id (external id)
        if (nextProps.silence.id && this.props.silence && this.props.silence.id) {
            if (nextProps.silence.id === this.props.silence.id) {
                return;
            }
        }


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

    addTime(field, time, unit, event) {
        let value = this.getValue(field, event);

        if (moment(value, moment.ISO_8601).isBefore(new Date())) {
            // time value is before NOW
            value = moment(new Date()).add(time, unit).toISOString();
        } else {
            // time value is after NOW
            value = moment(value, moment.ISO_8601).add(time, unit).toISOString();
        }

        var state = this.state;
        _.set(state, field, value);
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

        let matchers = Array.isArray(this.state.silence.matchers) ? this.state.silence.matchers : [];

        // filter team
        matchers = matchers.filter((row) => {
            if (row.name === "team") {
                return false;
            }

            return true;
        });

        let reltime = (time) => {
            let val = moment(time, moment.ISO_8601).fromNow();
            if (val === "a few seconds ago" || val === "a few seconds") {
                val = "now";
            }

            return (
                <span class="reltime">({val})</span>
            )
        };

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
                                        <label htmlFor="silence-form-team">Team</label>
                                        <select name="inputTeam" id="silence-form-team" className="form-control" value={this.getValue("team")} onChange={this.setValue.bind(this, "team")}>
                                            {this.props.config.Teams.map((row, value) =>
                                                <option key={row.Id} value={row.Name}>{row.Name}</option>
                                            )}
                                        </select>
                                    </div>

                                    <div className="form-group">
                                        <label htmlFor="silence-form-comment" className="inputRg">Description</label>
                                        <textarea id="silence-form-comment" className="form-control" value={this.getValue("silence.comment")} onChange={this.setValue.bind(this, "silence.comment")}  />
                                    </div>

                                    <div className="form-group">
                                        <div className="form-row">
                                            <div className="form-group col-md-6 form-group-rel">
                                                <label htmlFor="silence-form-startsAt" className="inputRg">Starts at {reltime(this.getValue("silence.startsAt"))}</label>
                                                <div className="form-group-rel">
                                                    <input id="silence-form-startsAt" className="form-control" value={this.getValue("silence.startsAt")} onChange={this.setValue.bind(this, "silence.startsAt")}  />
                                                    <div className="btn-group bnt-abs-right" role="group">
                                                        <button id="btnGroupDrop-startsAt" type="button"
                                                                className="btn btn-secondary dropdown-toggle btn-sm"
                                                                data-toggle="dropdown" aria-haspopup="true"
                                                                aria-expanded="false">
                                                            +
                                                        </button>
                                                        <div className="dropdown-menu" aria-labelledby="btnGroupDrop-startsAt">
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 1, "h")}>1 hour</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 2, "h")}>2 hours</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 4, "h")}>4 hours</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 8, "h")}>8 hours</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 1, "d")}>1 day</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 2, "d")}>2 day</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 4, "d")}>4 day</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 1, "w")}>1 week</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 2, "w")}>2 weeks</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 3, "w")}>3 weeks</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.startsAt", 4, "w")}>4 weeks</a>
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>

                                            <div className="form-group col-md-6 form-group-rel">
                                                <label htmlFor="silence-form-endsAt" className="inputRg">Ends at {reltime(this.getValue("silence.endsAt"))}</label>
                                                <div className="form-group-rel">
                                                    <input id="silence-form-endsAt" className="form-control" value={this.getValue("silence.endsAt")} onChange={this.setValue.bind(this, "silence.endsAt")}  />
                                                    <div className="btn-group bnt-abs-right" role="group">
                                                        <button id="btnGroupDrop-endsAt" type="button"
                                                                className="btn btn-secondary dropdown-toggle btn-sm"
                                                                data-toggle="dropdown" aria-haspopup="true"
                                                                aria-expanded="false">
                                                            +
                                                        </button>
                                                        <div className="dropdown-menu" aria-labelledby="btnGroupDrop-endsAt">
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 1, "h")}>1 hour</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 2, "h")}>2 hours</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 4, "h")}>4 hours</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 8, "h")}>8 hours</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 1, "d")}>1 day</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 2, "d")}>2 day</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 4, "d")}>4 day</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 1, "w")}>1 week</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 2, "w")}>2 weeks</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 3, "w")}>3 weeks</a>
                                                            <a className="dropdown-item" onClick={this.addTime.bind(this, "silence.endsAt", 4, "w")}>4 weeks</a>
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>

                                    <table className="table table-sm table-borderless table-striped">
                                        <colgroup>
                                            <col widht="*" />
                                            <col widht="*" />
                                            <col widht="50rem" />
                                            <col widht="20rem" />
                                        </colgroup>
                                        <thead>
                                            <tr>
                                                <th colspan="4">Matchers <small>(Alerts affected by this silence)</small></th>
                                            </tr>
                                            <tr>
                                                <th>Label</th>
                                                <th>Value</th>
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
                                                    <input className="form-control" value={this.getValue("silence.matchers[" + key + "].name")} onChange={this.setValue.bind(this, "silence.matchers[" + key + "].name")}  />
                                                </td>

                                                <td>
                                                    <input className="form-control" value={this.getValue("silence.matchers[" + key + "].value")} onChange={this.setValue.bind(this, "silence.matchers[" + key + "].value")}  />
                                                </td>

                                                <td>
                                                    <div className="form-check">
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

