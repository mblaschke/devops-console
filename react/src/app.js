import React, { Component } from 'react';
import { BrowserRouter as Router, Route } from 'react-router-dom';
import Base from './base';
import K8sNamespace from './kubernetes/namespace';
import K8sAccess from './kubernetes/access';
import AzureResourceGroups from './azure/resourcegroup';
import AzureRoleAssignment from './azure/roleassignment';
import MonitoringAlertmanager from "./monitoring/alertmanager";
import Settings from './general/settings';
import GeneralStats from './general/stats';

import '@popperjs/core';
import { usePopper } from 'react-popper';
import $ from "jquery";

class App extends Base {
    constructor(props) {
        super(props);

        this.searchCallback = false;
        this.janitorTimeout = false;

        this.state = {
            loggedIn: false,
            username: "",
            password: "",
            buttonState: "disabled",
            messageList: [],
        };

        // don't publish the whole app to public javascript
        let AppWrapper = {
            MessageExpiry: 15 * 1000,

            pushGlobalMessage: (type, message) => {
                this.pushGlobalMessage(type, message);
            },

            triggerSearch: (event) => {
                if (this.searchCallback) {
                    try {
                        this.searchCallback(event);
                    } catch {}
                }
            },

            registerSearch: (callback) => {
                this.searchCallback = callback;
            },

            enableSearch: () => {
                $("body").removeClass("hide-search")
            }
        };
        window.App = AppWrapper;

        this.startMessageJanitor();
    }

    startMessageJanitor() {
        let timeoutInitial = 10000;
        let timeoutCleanup = 250;

        let janitorFunc = () => {
            let timeout = timeoutInitial;
            let messageList = this.state.messageList;
            let now = new Date();

            if (!messageList) {
                messageList = [];
            }

            if (messageList.length > 0) {
                // filter not visible messages
                messageList = messageList.filter((row, num) => {
                    return row.show;
                });

                // hide elements if expired
                messageList.forEach(function(part, index) {
                    if((now - this[index].created) > window.App.MessageExpiry) {
                        this[index].show = false;
                        timeout = timeoutCleanup;
                    } else {
                        let expiresIn = window.App.MessageExpiry - (now - this[index].created);

                        if (expiresIn < timeout) {
                            timeout = expiresIn;
                        }
                    }
                }, messageList);

                if (!messageList) {
                    messageList = [];
                }

                this.setState({
                    messageList: messageList
                });
            }

            // make sure the timeout is not too small or too big
            if (timeout < timeoutCleanup) {
                timeout = timeoutCleanup;
            } else if (timeout > window.App.MessageExpiry) {
                timeout = window.App.MessageExpiry - timeoutCleanup;
            }

            if (this.janitorTimeout) {
                try {
                    this.janitorTimeout.cancel();
                } catch {}
            }
            this.janitorTimeout = setTimeout(janitorFunc, timeout);
        };

        janitorFunc();
    }

    handleChangeUsername(event) {
        if (event.target.value !== "") {
            this.setState({buttonState: ""});
        } else {
            this.setState({buttonState: "disabled"});
        }
        this.setState({username: event.target.value});
    }

    handleChangePassword(event) {
        if (event.target.value !== "") {
            this.setState({buttonState: ""});
        } else {
            this.setState({buttonState: "disabled"});
        }
        this.setState({password: event.target.value});
    }

    handleLogin() {
        this.ajax({
            type: 'POST',
            url: "/_webapi/_login",
            data: {
                username: this.state.username,
                password: this.state.password
            }
        }).done(() => {
            this.setState({loggedIn: true});
        });
    }

    pushGlobalMessage(type, message) {
        let messageList = this.state.messageList;
        if (!messageList) {
            messageList = [];
        }

        let lastIndex = messageList.length - 1;

        if (messageList.length > 0 && (messageList[lastIndex].type === type && messageList[lastIndex].original === message)) {
            // duplicate check
            messageList[lastIndex].text = messageList[lastIndex].original + " (+" + messageList[lastIndex].counter++ + ")";
            messageList[lastIndex].show = true;
            messageList[lastIndex].created = new Date();
        } else {
            messageList.push({
                type: type,
                original: message,
                text: message,
                counter: 1,
                created: new Date(),
                show: true
            });
        }

        this.setState({
            messageList: messageList
        });
    }

    removeGlobalMessage(num, event) {
        try {
            let messageList = this.state.messageList;
            if (messageList[num]) {
                delete messageList[num];
                this.setState({
                    messageList: messageList
                });
            }
        } catch {}

        this.handlePreventEvent(event);
    }

    renderMessageClass(message) {
        let ret = "alert alert-" + message.type + " alert-dismissible";

        if (message.show) {
            ret += " show";
        } else {
            ret += " fade";
        }

        return ret;
    }

    renderGlobalMessages() {
        if (this.state.messageList.length === 0) {
            return (<div></div>);
        }

        return (
            <div>
                {this.state.messageList.map((row, num) =>
                    <div className={this.renderMessageClass(row)} role="alert">
                        {row.text}
                        <button type="button" className="close" aria-label="Close" onClick={this.removeGlobalMessage.bind(this, num)}>
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                )}
            </div>
        )
    }

    render() {
        return (
            <Router>
                <div>
                    <div className="globalmessages">{this.renderGlobalMessages()}</div>
                    <Route path="/kubernetes/namespaces" component={K8sNamespace} />
                    <Route path="/kubernetes/access" component={K8sAccess} />

                    <Route path="/azure/resourcegroup" component={AzureResourceGroups} />
                    <Route path="/azure/roleassignment" component={AzureRoleAssignment} />

                    <Route path="/monitoring/alertmanager" component={MonitoringAlertmanager} />

                    <Route path="/general/settings" component={Settings} />
                    <Route path="/general/about" component={GeneralStats} />
                </div>
            </Router>
        )
    }
}

export default App;
