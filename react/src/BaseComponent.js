import React from 'react';
import { Component } from 'react';
import $ from 'jquery';
import * as utils from "./utils";
import _ from 'lodash';

class BaseComponent extends Component {

    setInputFocus() {
        setTimeout( () => {
            $(":input:text:visible:enabled").first().focus();
        }, 500);
    }

    handleXhr(jqxhr) {
        jqxhr.always((resp) => {
            // update CSRF token if needed
            let csrfToken = jqxhr.getResponseHeader("x-csrf-token");
            if (csrfToken) {
                window.CSRF_TOKEN = csrfToken;
            }
        });

        jqxhr.fail((jqxhr) => {
            if (jqxhr.status === 401) {
                this.setState({
                    globalError: "Login expired, please reload",
                    isStartup: false
                });
            } else if (jqxhr.responseJSON && jqxhr.responseJSON.Message) {
                window.App.pushGlobalMessage("danger", jqxhr.responseJSON.Message);
                this.setState({
                    globalError: jqxhr.responseJSON.Message,
                    isStartup: false
                });
            } else if (jqxhr.responseText) {
                window.App.pushGlobalMessage("danger", jqxhr.responseText);
            } else {
                window.App.pushGlobalMessage("danger", "Request failed, please check connectivity");
            }
        });

        jqxhr.done((jqxhr) => {
            if (jqxhr.Message) {
                window.App.pushGlobalMessage("success", jqxhr.Message);
            } else if (jqxhr.responseJSON && jqxhr.responseJSON.Message) {
                window.App.pushGlobalMessage("success", jqxhr.Message);
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
        let value = false;
        if (event.target) {
            value = event.target.type === 'checkbox' ? String(event.target.checked) : String(event.target.value);
        } else {
            value = event;
        }

        var state = this.state;
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

    ajax(opts) {
        if (!opts.headers) {
            opts.headers = [];
        }

        if (window.CSRF_TOKEN) {
            opts.headers["X-CSRF-Token"] = window.CSRF_TOKEN;
        }

        return $.ajax(opts);
    }

}

export default BaseComponent;
