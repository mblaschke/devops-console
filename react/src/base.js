import React from 'react';
import { Component } from 'react';
import $ from 'jquery';
import * as utils from "./utils";
import _ from 'lodash';

class Base extends Component {
    constructor(props) {
        super(props);

        this.state = {
            sortField: null,
            sortDir: "asc"
        }

        this.startHeartbeat();
    }

    buildAppConfig() {
        return {
            user: {
                username: '',
            },
            teams: [],
                alertmanager: {
                instances: []
            },
            quota: {},
            azure: {
                roleAssignment: {
                    roleDefinitions: [],
                    ttl: []
                },
                resourceGroup: {
                    tags: []
                }
            },
            kubernetes: {
                environments: [],
                namespace: {
                    settings: [],
                    networkPolicy: []
                }
            }
        };
    }


    loadConfig() {
        this.ajax({
            type: "GET",
            url: '/_webapi/app/config'
        }).done((jqxhr) => {
            if (jqxhr) {
                if (this.state.isStartup) {
                    this.setInputFocus();
                }

                if (!jqxhr.teams) {
                    jqxhr.teams = [];
                }

                this.setState({
                    isStartup: false,
                    config: jqxhr
                });

                // trigger init
                setTimeout(() => {
                    this.init();
                });
            }
        });
    }

    sortBy(field, text, callback) {
        let symbol = <i className="fas fa-sort disabled"></i>;
        if (this.state.sort && this.state.sort.field === field) {
            if (this.state.sort.dir === "asc") {
                symbol = <i className="fas fa-sort-up"></i>;
            } else {
                symbol = <i className="fas fa-sort-down"></i>;
            }
        }

        return (<a className="sortable" onClick={this.triggerSortBy.bind(this, field, callback)}><span>{text}</span>{symbol}</a>)
    }

    triggerSortBy(field, callback) {
        let sort;
        if (!this.state.sort) {
            sort = {
                field: field,
                dir: "desc",
                callback: callback
            };
        } else {
            sort = this.state.sort;
            sort.callback = callback;

            if (sort.field === field) {
                if (sort.dir === "asc") {
                    sort.dir = "desc";
                } else {
                    sort.dir = "asc";
                }
            } else {
                sort.field = field;
                sort.dir = "asc";
            }
        }

        this.setState({
            sort: sort
        })
    }

    sortDataset(list) {
        if (!this.state.sort) {
            return list;
        }

        let sort = this.state.sort;
        list = list.sort(function(a,b) {
            let aVal;
            let bVal;

            if (sort.callback) {
                // value by callback
                aVal = sort.callback(a)
                bVal = sort.callback(b)
            } else if (a[sort.field] && b[sort.field]) {
                // value by field
                aVal = a[sort.field];
                bVal = b[sort.field]
            }

            // do sort
            if (sort.dir === "asc") {
                if (aVal < bVal) return -1;
                if (aVal > bVal) return 1;
            } else {
                if (aVal < bVal) return 1;
                if (aVal > bVal) return -1;
            }

            return 0;
        });

        return list;
    }

    startHeartbeat() {
        if (!this.startHeartbeat.interval) {
            this.startHeartbeat.interval = setInterval(() => {
                this.heartbeat();
            }, 30 * 1000)
        }
    }

    heartbeat() {
        let jqxhr = this.ajax({
            type: "GET",
            url: '/_webapi/heartbeat'
        });

        this.handleXhr(jqxhr);
    }

    setInputFocus() {
        setTimeout( () => {
            $(":input:text:visible:enabled").first().focus();
        }, 500);
    }

    handleXhr(jqxhr) {
        jqxhr.always(() => {
            // update CSRF token if needed
            let csrfToken = jqxhr.getResponseHeader("x-csrf-token");
            if (csrfToken) {
                window.CSRF_TOKEN = csrfToken;
            }
        });

        jqxhr.fail((jqxhr) => {
            if (jqxhr.status === 401) {
                window.location.href = "/logout/forced";
            } else if (jqxhr.responseJSON && jqxhr.responseJSON.message) {
                window.App.pushGlobalMessage("danger", jqxhr.responseJSON.message);
                this.setState({
                    globalError: jqxhr.responseJSON.message,
                    isStartup: false
                });
            } else if (jqxhr.responseText) {
                window.App.pushGlobalMessage("danger", jqxhr.responseText);
            } else {
                window.App.pushGlobalMessage("danger", "Request failed, please check connectivity");
            }
        });

        jqxhr.done((jqxhr) => {
            if (!jqxhr) {
                return
            }

            if (jqxhr.message) {
                window.App.pushGlobalMessage("success", jqxhr.message);
            } else if (jqxhr.responseJSON && jqxhr.responseJSON.message) {
                window.App.pushGlobalMessage("success", jqxhr.message);
            }
        });
    }

    handlePreventEvent(event) {
        event.preventDefault();
        event.stopPropagation();
    }


    highlight(text) {
        let highlight = this.state.searchValue;

        if (highlight && highlight !== "") {
            // Split on higlight term and include term into parts, ignore case
            let parts = text.split(new RegExp(`(${highlight})`, 'gi'));
            return <span> { parts.map((part, i) =>
                <span key={i} className={part.toLowerCase() === highlight.toLowerCase() ? 'highlight' : '' }>
            { part }
        </span>)
            } </span>;
        } else {
            return <span>{text}</span>
        }
    }

    getValue(field) {
        return _.get(this.state, field);
    }

    setValue(field, event) {
        let value;
        if (event.target) {
            value = event.target.type === 'checkbox' ? String(event.target.checked) : String(event.target.value);
        } else {
            value = event;
        }

        let state = this.state;
        _.set(state, field, value);
        this.setState(state);
    }

    getValueCheckbox(field) {
        return utils.translateValueToCheckbox(_.get(this.state, field));
    }

    setValueCheckbox(field, event) {
        let value = event.target.type === 'checkbox' ? event.target.checked : String(event.target.value);

        var state = this.state;
        _.set(state, field, value);
        this.setState(state);
    }

    handleSearchChange(event) {
        this.setState({
            searchValue: event.target.value
        });
    }

    initTeam() {
        let state = this.state;

        // default team for local storage
        try {
            let lastSelectedTeam = "" + localStorage.getItem("team");
            this.state.config.Teams.map((row, value) => {
                if (row.Name === lastSelectedTeam) {
                    state.team = lastSelectedTeam;
                }
            });
        } catch {}

        // select first team if no selection available
        if (state.team === "") {
            if (this.state.config.Teams.length > 0) {
                state.team = this.state.config.Teams[0].Name
            }
        }

        this.setState(state);
    }

    setTeam(field, event) {
        let value;
        if (event.target) {
            value = event.target.type === 'checkbox' ? String(event.target.checked) : String(event.target.value);
        } else {
            value = event;
        }

        try {
            localStorage.setItem("team", value);
        } catch {}

        this.setValue(field, event);
    }

    renderTeamSelector(htmlId) {
        if (!htmlId) {
            htmlId = "formTeamSelector" + Math.random() * 10000;
        }

        return (
            <select className="form-control" id={htmlId} value={this.getValue("team")} onChange={this.setTeam.bind(this, "team")}>
                <option key="*" value="*">All teams</option>
                {this.state.config.teams.map((row, value) =>
                    <option key={row.Id} value={row.name}>{row.name}</option>
                )}
            </select>
        )
    }

    renderTeamSelectorWithlabel() {
        let htmlId = "formTeamSelector" + Math.random() * 10000;

        return (
            <div className="form-group">
                <label htmlFor={htmlId}>Team</label>
                {this.renderTeamSelector(htmlId)}
            </div>
        )
    }

    ajax(opts) {
        if (!opts.headers) {
            opts.headers = [];
        }

        if (window.CSRF_TOKEN) {
            opts.headers["X-CSRF-Token"] = window.CSRF_TOKEN;
        }

        if (!opts.contentType) {
            opts.contentType = "application/json";
        }

        let jqxhr = $.ajax(opts);
        this.handleXhr(jqxhr);
        return jqxhr;
    }

}

export default Base;
